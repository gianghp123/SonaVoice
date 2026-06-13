"use server"

import { auth } from "@clerk/nextjs/server"
import * as Sentry from "@sentry/nextjs"
import "server-only"
import type { BaseResponse, PaginatedMeta } from "./base-response"
import { camelToSnake, snakeToCamel } from "./case"

type ApiFetchOptions = {
  baseUrl?: string
  withCredentials?: boolean
  query?: Record<string, unknown>
  body?: unknown
} & Omit<RequestInit, "body">

type ApiResponseShape<T> = {
  data?: T
  meta?: PaginatedMeta
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null
}

function isBodyInit(value: unknown): value is BodyInit {
  return (
    typeof value === "string" ||
    value instanceof FormData ||
    value instanceof Blob ||
    value instanceof ArrayBuffer ||
    value instanceof URLSearchParams ||
    value instanceof ReadableStream
  )
}

function getErrorMessage(errorData: unknown): string {
  if (!isRecord(errorData)) return "Unknown error"
  const message = errorData.message
  if (typeof message === "string") return message

  const error = errorData.error

  if (typeof error === "string") return error
  if (isRecord(error) && typeof error.message === "string") {
    return error.message
  }
  return "Unknown error"
}

function getErrorCode(error: unknown): number {
  if (!isRecord(error)) return 500
  const code = error.code

  if (typeof code === "number") return code
  if (typeof code === "string") {
    const parsed = Number(code)
    return Number.isFinite(parsed) ? parsed : 500
  }

  return 500
}

function buildQueryString(query?: Record<string, unknown>) {
  const searchParams = new URLSearchParams()

  if (!query) return ""

  Object.entries(query).forEach(([key, value]) => {
    if (value === undefined || value === null || value === "") return
    const snakeKey = key.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`)
    if (typeof value === "object") {
      searchParams.append(snakeKey, JSON.stringify(camelToSnake(value)))
    } else {
      searchParams.append(snakeKey, String(value))
    }
  })

  return searchParams.toString()
}

export async function apiFetch<T>(
  url: string,
  options?: ApiFetchOptions
): Promise<BaseResponse<T>> {
  const startedAt = Date.now()

  try {
    const {
      withCredentials = false,
      baseUrl = process.env.API_URL,
      query,
      body: inputBody,
      ...fetchOptions
    } = options || {}

    const method = fetchOptions.method || "GET"

    if (!baseUrl) {
      Sentry.logger.error("API_URL is not configured", {
        area: "api-fetch",
        url,
        method,
      })

      throw new Error(
        "Server API_URL is not configured. Please set API_URL environment variable."
      )
    }

    const headers = new Headers(fetchOptions.headers)

    headers.set("apikey", process.env.API_KEY || "")

    let body: BodyInit | null | undefined

    if (inputBody !== undefined && inputBody !== null) {
      if (isBodyInit(inputBody)) {
        body = inputBody
      } else {
        headers.set("Content-Type", "application/json")
        body = JSON.stringify(camelToSnake(inputBody))
      }
    } else if (inputBody === null) {
      body = null
    }

    if (withCredentials) {
      const { getToken } = await auth()
      const token = await getToken()

      if (token) {
        headers.set("Authorization", `Bearer ${token}`)
      } else {
        Sentry.logger.warn("Clerk token missing for credentialed API request", {
          area: "api-fetch",
          url,
          method,
        })
      }
    }

    const queryString = buildQueryString(query)
    const fullUrl = `${baseUrl}${url}${queryString ? `?${queryString}` : ""}`

    const response = await fetch(fullUrl, {
      ...fetchOptions,
      method,
      headers,
      body,
    })

    const durationMs = Date.now() - startedAt

    if (!response.ok) {
      let errorData: unknown

      try {
        errorData = await response.json()
      } catch {
        Sentry.logger.warn("Failed to parse backend error response", {
          area: "api-fetch",
          url,
          method,
          status: response.status,
          durationMs,
        })
      }

      const message = getErrorMessage(errorData)

      if (response.status >= 500 || response.status === 429) {
        Sentry.logger.error("Backend API returned unexpected error", {
          area: "api-fetch",
          url,
          method,
          status: response.status,
          durationMs,
        })

        Sentry.captureException(new Error(message), {
          tags: {
            area: "api-fetch",
            type: "backend-response-error",
            status: String(response.status),
          },
          extra: {
            url,
            method,
            status: response.status,
            durationMs,
            responseBody: errorData,
          },
        })
      } else {
        Sentry.logger.info("Backend API returned expected non-OK response", {
          area: "api-fetch",
          url,
          method,
          status: response.status,
          durationMs,
        })
      }

      return {
        error: {
          code: response.status,
          message,
        },
        meta: undefined,
      }
    }

    let rawData: unknown
    const contentType = response.headers.get("content-type")

    if (contentType?.includes("application/json")) {
      const text = await response.text()

      try {
        rawData = text ? JSON.parse(text) : {}
      } catch (error) {
        Sentry.logger.error("Failed to parse backend JSON response", {
          area: "api-fetch",
          url,
          method,
          status: response.status,
          durationMs: Date.now() - startedAt,
        })

        throw error
      }
    } else {
      rawData = {}
    }

    const data = snakeToCamel<unknown>(rawData)

    if (isRecord(data)) {
      const responseData = data as ApiResponseShape<T>

      if (responseData.meta) {
        return {
          data: responseData.data as T,
          meta: responseData.meta,
        }
      }

      if ("data" in responseData) {
        return {
          data: responseData.data as T,
        }
      }
    }

    return {
      data: data as T,
    }
  } catch (error: unknown) {
    const durationMs = Date.now() - startedAt

    Sentry.logger.error("apiFetch crashed", {
      area: "api-fetch",
      url,
      method: options?.method || "GET",
      durationMs,
    })

    Sentry.captureException(error, {
      tags: {
        area: "api-fetch",
        type: "fetch-crash",
      },
      extra: {
        url,
        method: options?.method || "GET",
        queryKeys: options?.query ? Object.keys(options.query) : undefined,
        durationMs,
      },
    })

    return {
      error: {
        code: getErrorCode(error),
        message: error instanceof Error ? error.message : "Unknown error",
      },
      meta: undefined,
    }
  }
}