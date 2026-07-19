/**
 * Extracts a safe, loggable message from an HTTP/unknown error without
 * echoing the full error object — which for HttpErrorResponse can include
 * the raw response body (potentially containing tokens or other sensitive
 * data returned by the API) and full request URL.
 */
export function safeErrorMessage(error: unknown): string {
  if (error && typeof error === 'object') {
    const err = error as { message?: unknown; status?: unknown };
    if (typeof err.message === 'string') {
      return typeof err.status === 'number' ? `${err.message} (status ${err.status})` : err.message;
    }
  }
  return String(error);
}
