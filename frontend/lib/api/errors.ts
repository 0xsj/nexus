// lib/api/errors.ts

export const ApiErrorKind = {
  // Client errors
  BAD_REQUEST: "BAD_REQUEST",
  UNAUTHORIZED: "UNAUTHORIZED",
  FORBIDDEN: "FORBIDDEN",
  NOT_FOUND: "NOT_FOUND",
  CONFLICT: "CONFLICT",
  VALIDATION: "VALIDATION",

  // Server errors
  INTERNAL: "INTERNAL",
  SERVICE_UNAVAILABLE: "SERVICE_UNAVAILABLE",

  // Network errors
  NETWORK: "NETWORK",
  TIMEOUT: "TIMEOUT",
} as const;

export type ApiErrorKind = (typeof ApiErrorKind)[keyof typeof ApiErrorKind];

// Base shape all API errors share
interface ApiErrorBase {
  kind: ApiErrorKind;
  message: string;
  status: number;
}

// Specific error types with discriminated payloads
export interface BadRequestError extends ApiErrorBase {
  kind: typeof ApiErrorKind.BAD_REQUEST;
}

export interface UnauthorizedError extends ApiErrorBase {
  kind: typeof ApiErrorKind.UNAUTHORIZED;
}

export interface ForbiddenError extends ApiErrorBase {
  kind: typeof ApiErrorKind.FORBIDDEN;
}

export interface NotFoundError extends ApiErrorBase {
  kind: typeof ApiErrorKind.NOT_FOUND;
  resource?: string;
}

export interface ConflictError extends ApiErrorBase {
  kind: typeof ApiErrorKind.CONFLICT;
  conflictingField?: string;
}

export interface ValidationError extends ApiErrorBase {
  kind: typeof ApiErrorKind.VALIDATION;
  fields: Record<string, string[]>;
}

export interface InternalError extends ApiErrorBase {
  kind: typeof ApiErrorKind.INTERNAL;
}

export interface ServiceUnavailableError extends ApiErrorBase {
  kind: typeof ApiErrorKind.SERVICE_UNAVAILABLE;
  retryAfter?: number;
}

export interface NetworkError extends ApiErrorBase {
  kind: typeof ApiErrorKind.NETWORK;
}

export interface TimeoutError extends ApiErrorBase {
  kind: typeof ApiErrorKind.TIMEOUT;
}

// Discriminated union of all API errors
export type ApiError =
  | BadRequestError
  | UnauthorizedError
  | ForbiddenError
  | NotFoundError
  | ConflictError
  | ValidationError
  | InternalError
  | ServiceUnavailableError
  | NetworkError
  | TimeoutError;

// Result type for API operations
export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: ApiError };

// Factory functions
export const ApiErrors = {
  badRequest: (message: string, status = 400): BadRequestError => ({
    kind: ApiErrorKind.BAD_REQUEST,
    message,
    status,
  }),

  unauthorized: (message = "Unauthorized", status = 401): UnauthorizedError => ({
    kind: ApiErrorKind.UNAUTHORIZED,
    message,
    status,
  }),

  forbidden: (message = "Forbidden", status = 403): ForbiddenError => ({
    kind: ApiErrorKind.FORBIDDEN,
    message,
    status,
  }),

  notFound: (
    message = "Not found",
    resource?: string,
    status = 404
  ): NotFoundError => ({
    kind: ApiErrorKind.NOT_FOUND,
    message,
    resource,
    status,
  }),

  conflict: (
    message: string,
    conflictingField?: string,
    status = 409
  ): ConflictError => ({
    kind: ApiErrorKind.CONFLICT,
    message,
    conflictingField,
    status,
  }),

  validation: (
    message: string,
    fields: Record<string, string[]>,
    status = 422
  ): ValidationError => ({
    kind: ApiErrorKind.VALIDATION,
    message,
    fields,
    status,
  }),

  internal: (message = "Internal server error", status = 500): InternalError => ({
    kind: ApiErrorKind.INTERNAL,
    message,
    status,
  }),

  serviceUnavailable: (
    message = "Service unavailable",
    retryAfter?: number,
    status = 503
  ): ServiceUnavailableError => ({
    kind: ApiErrorKind.SERVICE_UNAVAILABLE,
    message,
    retryAfter,
    status,
  }),

  network: (message = "Network error"): NetworkError => ({
    kind: ApiErrorKind.NETWORK,
    message,
    status: 0,
  }),

  timeout: (message = "Request timeout"): TimeoutError => ({
    kind: ApiErrorKind.TIMEOUT,
    message,
    status: 0,
  }),
} as const;

// Type guards
export function isApiError(error: unknown): error is ApiError {
  return (
    typeof error === "object" &&
    error !== null &&
    "kind" in error &&
    "message" in error &&
    "status" in error
  );
}

export function isUnauthorized(error: ApiError): error is UnauthorizedError {
  return error.kind === ApiErrorKind.UNAUTHORIZED;
}

export function isValidationError(error: ApiError): error is ValidationError {
  return error.kind === ApiErrorKind.VALIDATION;
}

export function isNetworkError(
  error: ApiError
): error is NetworkError | TimeoutError {
  return (
    error.kind === ApiErrorKind.NETWORK || error.kind === ApiErrorKind.TIMEOUT
  );
}