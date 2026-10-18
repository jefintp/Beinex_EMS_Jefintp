import { HttpErrorResponse } from '@angular/common/http';

/**
 * Error body the Go API is expected to return:
 *   { "error": "validation failed", "fields": { "email": "email already exists" } }
 */
interface ApiErrorBody {
  error?: string;
  message?: string;
  fields?: Record<string, string>;
}

export interface ParsedApiError {
  status: number;
  message: string;
  /** Per-field messages, keyed by JSON field name (e.g. `email`, `joining_date`). */
  fields: Record<string, string>;
}

export function parseApiError(err: unknown): ParsedApiError {
  if (!(err instanceof HttpErrorResponse)) {
    return { status: -1, message: 'Something went wrong. Please try again.', fields: {} };
  }

  if (err.status === 0) {
    return {
      status: 0,
      message: 'Cannot reach the server. Check your connection and that the backend is running.',
      fields: {},
    };
  }

  let message: string | undefined;
  let fields: Record<string, string> = {};
  const body: unknown = err.error;

  if (body && typeof body === 'object') {
    const parsed = body as ApiErrorBody;
    message = parsed.error ?? parsed.message;
    fields = parsed.fields ?? {};
  } else if (typeof body === 'string') {
    const text = body.trim();
    // Ignore HTML error pages (e.g. proxy 502s); only keep short plain-text messages.
    if (text && text.length < 300 && !text.startsWith('<')) message = text;
  }

  return { status: err.status, message: message ?? defaultMessage(err.status), fields };
}

function defaultMessage(status: number): string {
  switch (status) {
    case 400:
    case 422:
      return 'The submitted data is invalid. Please review it and try again.';
    case 404:
      return 'The requested employee was not found.';
    case 409:
      return 'This record conflicts with an existing one.';
    default:
      return status >= 500
        ? 'The server ran into a problem. Please try again in a moment.'
        : 'The request failed. Please try again.';
  }
}
