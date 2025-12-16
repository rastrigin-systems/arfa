/**
 * API Error Utilities
 *
 * Centralized error handling for API responses.
 */

/**
 * Standard API error response structure
 */
export interface ApiError {
  error?: string;
  message?: string;
  status?: number;
}

/**
 * Type guard to check if an error is an API error
 */
export function isApiError(error: unknown): error is ApiError {
  return (
    typeof error === 'object' &&
    error !== null &&
    ('error' in error || 'message' in error)
  );
}

/**
 * Extract error message from API error response
 *
 * Handles various error formats returned by openapi-fetch
 */
export function getErrorMessage(error: unknown, fallback: string): string {
  if (!error) {
    return fallback;
  }

  if (isApiError(error)) {
    return error.message || error.error || fallback;
  }

  if (error instanceof Error) {
    return error.message;
  }

  if (typeof error === 'string') {
    return error;
  }

  return fallback;
}
