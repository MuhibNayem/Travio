<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { auth } from "$lib/runes/auth.svelte";
    import { goto } from "$app/navigation";

    let email = $state("user@example.com");
    let password = $state("password");

    function handleLogin() {
        // Mock login
        auth.login(
            {
                id: "1",
                name: "John Doe",
                email: email,
                role: "user",
            },
            "mock_token",
        );
        goto("/dashboard");
    }
</script>

<div class="glass-panel w-full p-8 relative overflow-hidden">
    <div
        class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-500 to-indigo-500"
    ></div>

    <div class="flex flex-col gap-6 text-center">
        <div>
            <h1 class="text-3xl font-black tracking-tight mb-2">
                Welcome Back
            </h1>
            <p class="text-muted-foreground">Sign in to manage your tickets</p>
        </div>

        <div class="flex flex-col gap-4 text-left">
            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    >Email</label
                >
                <Input
                    type="email"
                    bind:value={email}
                    class="bg-white/50 backdrop-blur-sm"
                />
            </div>
            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    >Password</label
                >
                <Input
                    type="password"
                    bind:value={password}
                    class="bg-white/50 backdrop-blur-sm"
                />
            </div>
        </div>

        <Button
            class="w-full h-12 text-lg font-bold shadow-lg shadow-blue-500/20"
            onclick={handleLogin}
        >
            Sign In
        </Button>

        <p class="text-sm text-gray-500">
            Don't have an account? <a
                href="/register"
                class="font-bold text-primary hover:underline">Register</a
            >
        </p>
    </div>
</div>
