import { login as apiLogin, logout as apiLogout, refreshTokens, register as apiRegister, type TokenPair, type OrgDetails, type CreateOrgInput } from '$lib/api';
import { jwtDecode } from 'jwt-decode';

interface User {
    id: string;
    name: string;
    email: string;
    role: 'user' | 'admin';
    organizationId?: string;
}

interface AuthState {
    user: User | null;
    accessToken: string | null;
    refreshToken: string | null;
    isLoading: boolean;
    error: string | null;
}

class AuthStore {
    user = $state<User | null>(null);
    accessToken = $state<string | null>(null);
    refreshToken = $state<string | null>(null);
    isLoading = $state(false);
    error = $state<string | null>(null);

    isAuthenticated = $derived(!!this.user && !!this.accessToken);

    constructor() {
        // Check local storage on initialization (if in browser)
        if (typeof localStorage !== 'undefined') {
            const storedUser = localStorage.getItem('user');
            const storedAccessToken = localStorage.getItem('accessToken');
            const storedRefreshToken = localStorage.getItem('refreshToken');

            if (storedAccessToken && storedRefreshToken) {
                this.accessToken = storedAccessToken;
                this.refreshToken = storedRefreshToken;
                this.decodeToken();
            }
        }
    }

    private saveToStorage() {
        if (typeof localStorage !== 'undefined') {
            if (this.user) {
                localStorage.setItem('user', JSON.stringify(this.user));
            }
            if (this.accessToken) {
                localStorage.setItem('accessToken', this.accessToken);
            }
            if (this.refreshToken) {
                localStorage.setItem('refreshToken', this.refreshToken);
            }
        }
    }

    private clearStorage() {
        if (typeof localStorage !== 'undefined') {
            localStorage.removeItem('user');
            localStorage.removeItem('accessToken');
            localStorage.removeItem('refreshToken');
        }
    }

    /**
     * Login with email and password
     */
    async login(email: string, password: string): Promise<boolean> {
        this.isLoading = true;
        this.error = null;

        try {
            const response = await apiLogin(email, password);

            // Set tokens
            this.accessToken = response.access_token;
            this.refreshToken = response.refresh_token;

            // Decode token to get user details
            this.decodeToken();

            this.saveToStorage();
            return true;
        } catch (e) {
            this.error = e instanceof Error ? e.message : 'Login failed';
            return false;
        } finally {
            this.isLoading = false;
        }
    }

    private decodeToken() {
        if (!this.accessToken) return;

        try {
            const decoded: any = jwtDecode(this.accessToken);
            this.user = {
                id: decoded.sub || decoded.user_id,
                name: decoded.name || decoded.email?.split('@')[0] || 'User',
                email: decoded.email,
                role: decoded.role || 'user',
                organizationId: decoded.org_id || decoded.organization_id
            };
        } catch (e) {
            console.error('Failed to decode token', e);
        }
    }

    /**
     * Register a new user (creates org transactionally if needed)
     */
    async register(
        email: string,
        password: string,
        name: string,
        orgName?: string,
        orgDetails: OrgDetails = {}
    ): Promise<boolean> {
        this.isLoading = true;
        this.error = null;

        try {
            let newOrganization: CreateOrgInput | undefined;

            // Prepare new organization payload if name provided
            if (orgName) {
                newOrganization = {
                    name: orgName,
                    ...orgDetails
                };
            }

            // Register user with optional new organization
            await apiRegister(email, password, name, undefined, newOrganization);

            return true;
        } catch (e) {
            this.error = e instanceof Error ? e.message : 'Registration failed';
            return false;
        } finally {
            this.isLoading = false;
        }
    }

    /**
     * Logout - invalidates refresh token on server
     */
    async logout(): Promise<void> {
        try {
            if (this.refreshToken) {
                await apiLogout(this.refreshToken);
            }
        } catch (e) {
            console.error('Logout error:', e);
        } finally {
            this.user = null;
            this.accessToken = null;
            this.refreshToken = null;
            this.clearStorage();
        }
    }

    /**
     * Refresh access token using stored refresh token
     */
    async refresh(): Promise<boolean> {
        if (!this.refreshToken) {
            return false;
        }

        try {
            const response = await refreshTokens(this.refreshToken);
            this.accessToken = response.access_token;
            this.refreshToken = response.refresh_token;
            this.saveToStorage();
            return true;
        } catch (e) {
            // Refresh failed, clear auth state
            await this.logout();
            return false;
        }
    }

    /**
     * Get the current access token, refreshing if needed
     */
    async getValidToken(): Promise<string | null> {
        // For simplicity, just return current token
        // In production, check expiry and refresh if needed
        return this.accessToken;
    }

    /**
     * Legacy method for mock login (keep for backward compatibility during transition)
     * @deprecated Use login() instead
     */
    legacyLogin(user: User, token: string) {
        this.user = user;
        this.accessToken = token;
        this.saveToStorage();
    }
}

export const auth = new AuthStore();
