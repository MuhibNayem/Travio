import { login as apiLogin, logout as apiLogout, refreshTokens, createOrganization, register as apiRegister, type TokenPair } from '$lib/api';

interface User {
    id: string;
    name: string;
    email: string;
    role: 'user' | 'admin';
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

            if (storedUser && storedAccessToken && storedRefreshToken) {
                try {
                    this.user = JSON.parse(storedUser);
                    this.accessToken = storedAccessToken;
                    this.refreshToken = storedRefreshToken;
                } catch (e) {
                    console.error('Failed to parse user from local storage', e);
                    this.clearStorage();
                }
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

            // Create user from email (in a real app, you'd decode the JWT or fetch user profile)
            this.user = {
                id: 'user-id', // Would come from JWT claims
                name: email.split('@')[0],
                email: email,
                role: 'user',
            };

            this.saveToStorage();
            return true;
        } catch (e) {
            this.error = e instanceof Error ? e.message : 'Login failed';
            return false;
        } finally {
            this.isLoading = false;
        }
    }

    /**
     * Register a new user (creates org first, then user)
     */
    async register(email: string, password: string, orgName?: string): Promise<boolean> {
        this.isLoading = true;
        this.error = null;

        try {
            let organizationId: string | undefined;

            // Create organization first if name provided
            if (orgName) {
                const orgResponse = await createOrganization(orgName, 'free');
                organizationId = orgResponse.organization_id;
            }

            // Register user
            await apiRegister(email, password, organizationId);

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
