/**
 * Authentication API client
 */

import { api, ApiError } from './api';

export interface TokenPair {
    access_token: string;
    refresh_token: string;
    expires_in: number;
}

export interface LoginResponse extends TokenPair { }

export interface RegisterResponse {
    user_id: string;
}

export interface CreateOrgResponse {
    organization_id: string;
    status: string;
}

export interface Session {
    id: string;
    device_info: string;
    ip_address: string;
    created_at: string;
    expires_at: string;
}

export interface UserContext {
    id: string;
    organization_id: string;
    role: string;
}

/**
 * Login with email and password
 */
export async function login(email: string, password: string): Promise<LoginResponse> {
    return api.post<LoginResponse>('/v1/auth/login', { email, password });
}

export interface OrgDetails {
    address?: string;
    phone?: string;
    email?: string;
    website?: string;
}

export interface CreateOrgInput {
    name: string;
    address?: string;
    phone?: string;
    email?: string;
    website?: string;
}

/**
 * Register a new user
 */
export async function register(
    email: string,
    password: string,
    name: string,
    organizationId?: string,
    newOrganization?: CreateOrgInput
): Promise<RegisterResponse> {
    return api.post<RegisterResponse>('/v1/auth/register', {
        email,
        password,
        name,
        organization_id: organizationId,
        new_organization: newOrganization,
    });
}

/**
 * Create a new organization (must be done before registering if org required)
 */
export async function createOrganization(
    name: string,
    planId: string = 'free',
    details: OrgDetails = {}
): Promise<CreateOrgResponse> {
    return api.post<CreateOrgResponse>('/v1/orgs', {
        name,
        plan_id: planId,
        ...details
    });
}

/**
 * Logout - invalidate the refresh token (or cookie)
 */
export async function logout(refreshToken?: string): Promise<void> {
    return api.post('/v1/auth/logout', { refresh_token: refreshToken || "" });
}

/**
 * Logout from all devices
 */
export async function logoutAll(accessToken: string): Promise<void> {
    return api.post('/v1/auth/logout-all', {}, accessToken);
}

/**
 * Refresh access token using refresh token check (or cookie)
 */
export async function refreshTokens(refreshToken?: string): Promise<TokenPair> {
    return api.post<TokenPair>('/v1/auth/refresh', { refresh_token: refreshToken || "" });
}

/**
 * Get current user context from session
 */
export async function getMe(): Promise<UserContext> {
    return api.get<UserContext>('/v1/auth/me');
}

/**
 * Get active sessions for current user
 */
export async function getSessions(accessToken: string): Promise<Session[]> {
    return api.get<Session[]>('/v1/auth/sessions', accessToken);
}

export { ApiError };
