// lib/query/client.ts

import { QueryClient } from "@tanstack/react-query";
import { ApiErrorKind, isApiError } from "@/lib/api";

export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 60 * 1000, // 1 minute
        gcTime: 5 * 60 * 1000, // 5 minutes
        retry: (failureCount, error) => {
          // Don't retry on client errors
          if (isApiError(error)) {
            switch (error.kind) {
              case ApiErrorKind.UNAUTHORIZED:
              case ApiErrorKind.FORBIDDEN:
              case ApiErrorKind.NOT_FOUND:
              case ApiErrorKind.VALIDATION:
              case ApiErrorKind.CONFLICT:
                return false;
            }
          }

          // Retry network errors up to 3 times
          return failureCount < 3;
        },
        refetchOnWindowFocus: false,
      },
      mutations: {
        retry: false,
      },
    },
  });
}

// Singleton for client-side usage
let browserQueryClient: QueryClient | undefined;

export function getQueryClient(): QueryClient {
  // Server: always create new client
  if (typeof window === "undefined") {
    return createQueryClient();
  }

  // Browser: reuse singleton
  if (!browserQueryClient) {
    browserQueryClient = createQueryClient();
  }

  return browserQueryClient;
}