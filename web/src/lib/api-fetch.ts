'use server'

import 'server-only'
import { auth } from "@clerk/nextjs/server";
import * as Sentry from '@sentry/nextjs';
import type { BaseResponse } from "./base-response";
import { snakeToCamel } from "./case";

type ApiFetchOptions = {
  baseUrl?: string;
  withCredentials?: boolean;
  query?: Record<string, any>;
} & RequestInit;

export async function apiFetch<T = any>(
  url: string,
  options?: ApiFetchOptions
): Promise<BaseResponse<T>> {
  const startedAt = Date.now();

  try {
    const {
      withCredentials = false,
      baseUrl = process.env.API_URL,
      query,
      ...fetchOptions
    } = options || {};

    const method = fetchOptions.method || "GET";

    if (!baseUrl) {
      Sentry.logger.error("API_URL is not configured", {
        area: "api-fetch",
        url,
        method,
      });

      throw new Error(
        "Server API_URL is not configured. Please set API_URL environment variable."
      );
    }

    const headers: Record<string, any> = {
      ...fetchOptions?.headers,
      apikey: process.env.API_KEY || "",
    };

    if (withCredentials) {
      const { getToken } = await auth();
      const token = await getToken();

      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      } else {
        Sentry.logger.warn("Clerk token missing for credentialed API request", {
          area: "api-fetch",
          url,
          method,
        });
      }
    }

    const searchParams = new URLSearchParams();

    if (query) {
      Object.entries(query).forEach(([key, value]) => {
        if (value === undefined || value === null || value === "") return;

        if (typeof value === "object") {
          searchParams.append(key, JSON.stringify(value));
        } else {
          searchParams.append(key, String(value));
        }
      });
    }

    const queryString = searchParams.toString();
    const fullUrl = `${baseUrl}${url}${queryString ? `?${queryString}` : ""}`;

    const response = await fetch(fullUrl, {
      method,
      ...fetchOptions,
      headers,
    });

    const durationMs = Date.now() - startedAt;

    if (!response.ok) {
      let message = "Unknown error";
      let errorData: any;

      try {
        errorData = await response.json();
        message = errorData.error?.message || errorData.message || message;
      } catch (_) {
        Sentry.logger.warn("Failed to parse backend error response", {
          area: "api-fetch",
          url,
          method,
          status: response.status,
          durationMs,
        });
      }

      if (response.status >= 500 || response.status === 429) {
        Sentry.logger.error("Backend API returned unexpected error", {
          area: "api-fetch",
          url,
          method,
          status: response.status,
          durationMs,
        });

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
        });
      } else {
        Sentry.logger.info("Backend API returned expected non-OK response", {
          area: "api-fetch",
          url,
          method,
          status: response.status,
          durationMs,
        });
      }

      return {
        error: {
          code: response.status,
          message,
        },
        meta: undefined,
      } as BaseResponse<T>;
    }

    let rawData: any;
    const contentType = response.headers.get("content-type");

    if (contentType && contentType.includes("application/json")) {
      const text = await response.text();

      try {
        rawData = text ? JSON.parse(text) : {};
      } catch (error) {
        const durationMs = Date.now() - startedAt;

        Sentry.logger.error("Failed to parse backend JSON response", {
          area: "api-fetch",
          url,
          method,
          status: response.status,
          durationMs,
        });

        throw error;
      }
    } else {
      rawData = {};
    }

    const data = snakeToCamel<any>(rawData);

    if (data.pagination || data.meta) {
      return {
        data: data.data ? (data.data as T) : ([] as T),
        meta: {
          pagination: data.pagination,
          ...data.meta,
        },
      } as BaseResponse<T>;
    }

    return {
      data: data.data !== undefined ? (data.data as T) : (data as T),
    } as BaseResponse<T>;
  } catch (error: any) {
    const durationMs = Date.now() - startedAt;

    Sentry.logger.error("apiFetch crashed", {
      area: "api-fetch",
      url,
      method: options?.method || "GET",
      durationMs,
    });

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
    });

    return {
      error: {
        code: error.code || 500,
        message: error.message || "Unknown error",
      },
      meta: undefined,
    } as BaseResponse<T>;
  }
}