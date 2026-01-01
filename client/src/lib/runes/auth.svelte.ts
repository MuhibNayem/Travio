interface User {
    id: string;
    name: string;
    email: string;
    role: "user" | "admin";
}

class AuthStore {
    user = $state<User | null>(null);
    token = $state<string | null>(null);

    isAuthenticated = $derived(!!this.user);

    constructor() {
        // Check local storage on initialization (if in browser)
        if (typeof localStorage !== "undefined") {
            const storedUser = localStorage.getItem("user");
            const storedToken = localStorage.getItem("token");
            if (storedUser && storedToken) {
                try {
                    this.user = JSON.parse(storedUser);
                    this.token = storedToken;
                } catch (e) {
                    console.error("Failed to parse user from local storage", e);
                    this.logout();
                }
            }
        }
    }

    login(user: User, token: string) {
        this.user = user;
        this.token = token;
        if (typeof localStorage !== "undefined") {
            localStorage.setItem("user", JSON.stringify(user));
            localStorage.setItem("token", token);
        }
    }

    logout() {
        this.user = null;
        this.token = null;
        if (typeof localStorage !== "undefined") {
            localStorage.removeItem("user");
            localStorage.removeItem("token");
        }
    }
}

export const auth = new AuthStore();
