// lib/api/client.ts

import {
  type ApiError,
  type ApiResult,
  ApiErrorKind,
  ApiErrors,
  isApiError,
} from "./errors";

export interface RequestConfig extends Omit<RequestInit, "body"> {
  body?: unknown;
  params?: Record<string, string | number | boolean | undefined>;
  timeout?: number;
}

export interface ApiClientConfig {
  baseUrl: string;
  timeout?: number;
  getAuthToken?: () => string | null;
  onUnauthorized?: () => void;
}

// Shape we expect from backend error responses
interface BackendErrorResponse {
  code?: string;
  message?: string;
  details?: Record<string, unknown>;
  fields?: Record<string, string[]>;
}

export function createApiClient(config: ApiClientConfig) {
  const { baseUrl, timeout = 30000, getAuthToken, onUnauthorized } = config;

  async function request<T>(
    method: string,
    path: string,
    options: RequestConfig = {}
  ): Promise<ApiResult<T>> {
    const { body, params, timeout: requestTimeout, ...init } = options;

    // Build URL with query params
    const url = new URL(path, baseUrl);
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          url.searchParams.set(key, String(value));
        }
      });
    }

    // Build headers
    const headers = new Headers(init.headers);
    if (!headers.has("Content-Type") && body !== undefined) {
      headers.set("Content-Type", "application/json");
    }

    const token = getAuthToken?.();
    if (token && !headers.has("Authorization")) {
      headers.set("Authorization", `Bearer ${token}`);
    }

    // Create abort controller for timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(
      () => controller.abort(),
      requestTimeout ?? timeout
    );

    try {
      const response = await fetch(url.toString(), {
        ...init,
        method,
        headers,
        body: body !== undefined ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        const error = await parseErrorResponse(response);

        if (error.kind === ApiErrorKind.UNAUTHORIZED) {
          onUnauthorized?.();
        }

        return { ok: false, error };
      }

      // Handle 204 No Content
      if (response.status === 204) {
        return { ok: true, data: undefined as T };
      }

      const data = (await response.json()) as T;
      return { ok: true, data };
    } catch (error) {
      clearTimeout(timeoutId);

      if (isApiError(error)) {
        return { ok: false, error };
      }

      if (error instanceof DOMException && error.name === "AbortError") {
        return { ok: false, error: ApiErrors.timeout() };
      }

      return {
        ok: false,
        error: ApiErrors.network(
          error instanceof Error ? error.message : "Unknown error"
        ),
      };
    }
  }

  async function parseErrorResponse(response: Response): Promise<ApiError> {
    let body: BackendErrorResponse = {};

    try {
      body = await response.json();
    } catch {
      // Failed to parse JSON, use status text
    }

    const message = body.message ?? response.statusText ?? "An error occurred";

    switch (response.status) {
      case 400:
        return ApiErrors.badRequest(message, response.status);

      case 401:
        return ApiErrors.unauthorized(message, response.status);

      case 403:
        return ApiErrors.forbidden(message, response.status);

      case 404:
        return ApiErrors.notFound(
          message,
          body.details?.resource as string | undefined,
          response.status
        );

      case 409:
        return ApiErrors.conflict(
          message,
          body.details?.field as string | undefined,
          response.status
        );

      case 422:
        return ApiErrors.validation(
          message,
          body.fields ?? {},
          response.status
        );

      case 503:
        return ApiErrors.serviceUnavailable(
          message,
          body.details?.retryAfter as number | undefined,
          response.status
        );

      default:
        return ApiErrors.internal(message, response.status);
    }
  }

  return {
    get: <T>(path: string, options?: RequestConfig) =>
      request<T>("GET", path, options),

    post: <T>(path: string, body?: unknown, options?: RequestConfig) =>
      request<T>("POST", path, { ...options, body }),

    put: <T>(path: string, body?: unknown, options?: RequestConfig) =>
      request<T>("PUT", path, { ...options, body }),

    patch: <T>(path: string, body?: unknown, options?: RequestConfig) =>
      request<T>("PATCH", path, { ...options, body }),

    delete: <T>(path: string, options?: RequestConfig) =>
      request<T>("DELETE", path, options),
  };
}

export type ApiClient = ReturnType<typeof createApiClient>;