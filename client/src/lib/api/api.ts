/**
 * Base API client for Travio frontend
 */

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8888';
export class ApiError extends Error {
    constructor(
        public status: number,
        public statusText: string,
        message: string
    ) {
        super(message);
        this.name = 'ApiError';
    }
}

interface RequestOptions {
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
    body?: unknown;
    headers?: Record<string, string>;
    _retry?: boolean;
}

/**
 * Makes an API request with JSON handling and error wrapping
 */
export async function apiRequest<T>(
    endpoint: string,
    options: RequestOptions = {}
): Promise<T> {
    const { method = 'GET', body, headers = {}, _retry } = options;

    const requestHeaders: Record<string, string> = {
        'Content-Type': 'application/json',
        ...headers,
    };

    const config: RequestInit = {
        method,
        headers: requestHeaders,
        credentials: 'include', // Enable Cookies
    };

    if (body && method !== 'GET') {
        config.body = JSON.stringify(body);
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, config);

    if (response.status === 401) {
        if (!_retry) {
            const refreshed = await tryRefreshAccessToken();
            if (refreshed) {
                return apiRequest<T>(endpoint, { ...options, _retry: true });
            }
        }
        emitAuthCleared();
    }

    if (!response.ok) {
        let errorMessage = response.statusText;
        try {
            const errorBody = await response.json();
            errorMessage = errorBody.message || errorBody.error || errorMessage;
        } catch {
            // Ignore JSON parse errors
        }
        throw new ApiError(response.status, response.statusText, errorMessage);
    }

    // Handle 204 No Content
    if (response.status === 204) {
        return undefined as T;
    }

    return response.json();
}

/**
 * Convenience methods
 */
export const api = {
    get: <T>(endpoint: string) =>
        apiRequest<T>(endpoint, { method: 'GET' }),

    post: <T>(endpoint: string, body: unknown) =>
        apiRequest<T>(endpoint, { method: 'POST', body }),

    put: <T>(endpoint: string, body: unknown) =>
        apiRequest<T>(endpoint, { method: 'PUT', body }),

    delete: <T>(endpoint: string) =>
        apiRequest<T>(endpoint, { method: 'DELETE' }),
};

async function tryRefreshAccessToken(): Promise<boolean> {
    try {
        const response = await fetch(`${API_BASE_URL}/v1/auth/refresh`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({}),
        });

        if (!response.ok) {
            return false;
        }

        if (typeof window !== 'undefined') {
            window.dispatchEvent(new CustomEvent('auth:tokens-refreshed'));
        }
        return true;
    } catch {
        return false;
    }
}

function emitAuthCleared() {
    if (typeof window !== 'undefined') {
        window.dispatchEvent(new CustomEvent('auth:tokens-cleared'));
    }
}
