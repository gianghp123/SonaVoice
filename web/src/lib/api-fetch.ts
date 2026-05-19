'server-only'
import type { BaseResponse } from "./base-response";
import { snakeToCamel } from "./case";
import { auth } from "@clerk/nextjs/server";

type ApiFetchOptions = {
  baseUrl?: string;
  withCredentials?: boolean;
  query?: Record<string, any>;
} & RequestInit;

export async function apiFetch<T = any>(
  url: string,
  options?: ApiFetchOptions
): Promise<BaseResponse<T>> {

  try {
    const {
      withCredentials = false,
      baseUrl = process.env.API_URL,
      query,
      ...fetchOptions
    } = options || {};

    if (!baseUrl) {
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
      }
    }

    const searchParams = new URLSearchParams();
    if (query) {
      Object.entries(query).forEach(([key, value]) => {
        if (value === undefined || value === null || value === "") return;

        if (typeof value === "object") {
          // If it's a DynamoDB object, we MUST stringify it properly
          searchParams.append(key, JSON.stringify(value));
        } else {
          searchParams.append(key, String(value));
        }
      });
    }

    const queryString = searchParams.toString();

    const fullUrl = `${baseUrl}${url}${queryString ? `?${queryString}` : ""}`;

    console.log("🚀 Calling API:", fullUrl);

    const response = await fetch(fullUrl, {
      method: fetchOptions.method || "GET",
      ...fetchOptions,
      headers
    });

    if (!response.ok) {
      let message = "Unknown error";
      try {
        const errorData = await response.json();
        message = errorData.error.message || errorData.message || message;
      } catch (_) { }
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
      rawData = text ? JSON.parse(text) : {};
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
    console.error(error);
    return {
      error: {
        code: error.code || 500,
        message: error.message || "Unknown error",
      },
      meta: undefined,
    } as BaseResponse<T>;
  }
}