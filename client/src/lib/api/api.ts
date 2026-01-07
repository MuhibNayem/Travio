/**
 * Base API client for Travio frontend
 */

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

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
    token?: string;
}

/**
 * Makes an API request with JSON handling and error wrapping
 */
export async function apiRequest<T>(
    endpoint: string,
    options: RequestOptions = {}
): Promise<T> {
    const { method = 'GET', body, headers = {}, token } = options;

    const requestHeaders: Record<string, string> = {
        'Content-Type': 'application/json',
        ...headers,
    };

    if (token) {
        requestHeaders['Authorization'] = `Bearer ${token}`;
    }

    const config: RequestInit = {
        method,
        headers: requestHeaders,
    };

    if (body && method !== 'GET') {
        config.body = JSON.stringify(body);
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, config);

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
    get: <T>(endpoint: string, token?: string) =>
        apiRequest<T>(endpoint, { method: 'GET', token }),

    post: <T>(endpoint: string, body: unknown, token?: string) =>
        apiRequest<T>(endpoint, { method: 'POST', body, token }),

    put: <T>(endpoint: string, body: unknown, token?: string) =>
        apiRequest<T>(endpoint, { method: 'PUT', body, token }),

    delete: <T>(endpoint: string, token?: string) =>
        apiRequest<T>(endpoint, { method: 'DELETE', token }),
};
