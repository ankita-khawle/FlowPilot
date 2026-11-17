
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '';

const REQUEST_TIMEOUT = 120_000;

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly code?: 'NETWORK' | 'TIMEOUT' | 'INVALID_RESPONSE'
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const controller = new AbortController();

  const timeoutId = window.setTimeout(
    () => controller.abort(),
    REQUEST_TIMEOUT
  );

  try {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      signal: controller.signal,
    });

    let payload: unknown;

    try {
      payload = await response.json();
    } catch {
      throw new ApiError(
        'The server returned an invalid response.',
        response.status,
        'INVALID_RESPONSE'
      );
    }

    if (!response.ok) {
      const message =
        isRecord(payload) && typeof payload.error === 'string'
          ? payload.error
          : 'Something went wrong. Please try again.';

      throw new ApiError(message, response.status);
    }

    return payload as T;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    if (error instanceof DOMException && error.name === 'AbortError') {
      throw new ApiError(
        'The request took too long. Please try again.',
        0,
        'TIMEOUT'
      );
    }

    throw new ApiError(
      'Unable to connect to FlowPilot. Check your connection and try again.',
      0,
      'NETWORK'
    );
  } finally {
    window.clearTimeout(timeoutId);
  }
}
