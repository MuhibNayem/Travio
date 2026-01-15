import { login as apiLogin, logout as apiLogout, register as apiRegister, getMe, type OrgDetails, type CreateOrgInput, type UserContext } from '$lib/api';

interface User {
    id: string;
    name: string; // Not in UserContext, might need to fetch profile or decoded from ID?
    // Wait, GetMe only returns ID, OrgID, Role.
    // Name is in Profile service or I need to update GetMe to return Name.
    // Claims usually have Name. Check Auth Middleware?
    // Middleware extracts "sub", "org_id", "role".
    // Name is NOT in context.
    // I should update GetMe to fetch name or just use Email/ID for now.
    // Or Name is optional.
    email: string; // Not in Context.
    role: 'user' | 'admin' | 'operator';
    organizationId?: string;
}

// Update UserContext in api/auth.ts to include more info?
// Middleware parses JWT. JWT has name/email if I put them there.
// Gateway `Login` output has tokens (which have claims).
// But `GetMe` reads from Context.
// Context only has ID, Org, Role.
// Use these for now. Name/Email will be empty or placeholders.
// Later I should fetch Profile.

class AuthStore {
    user = $state<User | null>(null);
    isLoading = $state(true); // Start loading to check session
    error = $state<string | null>(null);

    isAuthenticated = $derived(!!this.user);

    constructor() {
        if (typeof window !== 'undefined') {
            this.fetchUser();
            window.addEventListener('auth:tokens-cleared', () => {
                this.user = null;
            });
        }
    }

    async fetchUser() {
        this.isLoading = true;
        try {
            const context = await getMe();
            this.user = {
                id: context.id,
                name: 'User', // Placeholder until Profile API
                email: '',    // Placeholder
                role: context.role as any,
                organizationId: context.organization_id
            };
        } catch (e) {
            // Not authenticated
            this.user = null;
        } finally {
            this.isLoading = false;
        }
    }

    async login(email: string, password: string): Promise<boolean> {
        this.isLoading = true;
        this.error = null;

        try {
            await apiLogin(email, password);
            // Login successful (Cookies set). Now fetch user details.
            await this.fetchUser();
            return true;
        } catch (e) {
            this.error = e instanceof Error ? e.message : 'Login failed';
            return false;
        } finally {
            this.isLoading = false;
        }
    }

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
            if (orgName) {
                newOrganization = { name: orgName, ...orgDetails };
            }

            await apiRegister(email, password, name, undefined, newOrganization);

            // Auto-login after register
            await this.login(email, password);
            return true;
        } catch (e) {
            this.error = e instanceof Error ? e.message : 'Registration failed';
            return false;
        } finally {
            this.isLoading = false;
        }
    }

    async logout(): Promise<void> {
        try {
            await apiLogout();
        } catch (e) {
            console.error('Logout error:', e);
        } finally {
            this.user = null;
            // Force reload to clear any JS state/sockets? Or just clear state.
            if (typeof window !== 'undefined') {
                // optional: window.location.href = '/login';
            }
        }
    }
}

export const auth = new AuthStore();
