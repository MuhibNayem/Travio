<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { auth } from "$lib/runes/auth.svelte";
    import { goto } from "$app/navigation";
    import { Loader2 } from "@lucide/svelte";
    import { toast } from "svelte-sonner";

    import { subscriptionApi } from "$lib/api/subscription";

    let email = $state("");
    let password = $state("");

    async function handleLogin() {
        if (!email || !password) {
            return;
        }

        const success = await auth.login(email, password);
        if (success) {
            toast.success("Welcome back!", {
                description: "You have successfully signed in.",
            });

            // Role-based redirect
            if (auth.user?.role === "user") {
                goto("/search");
            } else if (
                auth.user?.role === "admin" &&
                auth.user?.organizationId
            ) {
                try {
                    // Check subscription status
                    const sub = await subscriptionApi.getSubscription(
                        auth.user.organizationId,
                    );
                    // If on free plan, redirect to onboarding to upsell/setup
                    // Assuming 'plan_free' is the ID for free tier
                    if (sub.plan_id === "plan_free") {
                        goto("/onboarding");
                    } else {
                        goto("/dashboard");
                    }
                } catch (e) {
                    // Fallback if subscription fetch fails (e.g. fresh org? or error)
                    // If 404, it might mean no subscription -> Onboarding
                    goto("/onboarding");
                }
            } else {
                // Default fallback
                goto("/dashboard");
            }
        } else {
            toast.error("Login failed", {
                description: auth.error || "Please check your credentials.",
            });
        }
    }

    function handleKeyDown(event: KeyboardEvent) {
        if (event.key === "Enter") {
            handleLogin();
        }
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

        {#if auth.error}
            <div
                class="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-3 text-sm text-red-600 dark:text-red-400"
            >
                {auth.error}
            </div>
        {/if}

        <div class="flex flex-col gap-4 text-left">
            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    for="email">Email</label
                >
                <Input
                    id="email"
                    type="email"
                    bind:value={email}
                    class="bg-white/50 backdrop-blur-sm"
                    placeholder="you@example.com"
                    disabled={auth.isLoading}
                    onkeydown={handleKeyDown}
                />
            </div>
            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    for="password">Password</label
                >
                <Input
                    id="password"
                    type="password"
                    bind:value={password}
                    class="bg-white/50 backdrop-blur-sm"
                    placeholder="••••••••"
                    disabled={auth.isLoading}
                    onkeydown={handleKeyDown}
                />
            </div>
        </div>

        <Button
            class="w-full h-12 text-lg font-bold shadow-lg shadow-blue-500/20"
            onclick={handleLogin}
            disabled={auth.isLoading || !email || !password}
        >
            {#if auth.isLoading}
                <Loader2 class="mr-2 h-5 w-5 animate-spin" />
                Signing in...
            {:else}
                Sign In
            {/if}
        </Button>

        <p class="text-sm text-gray-500">
            Don't have an account? <a
                href="/register"
                class="font-bold text-primary hover:underline">Register</a
            >
        </p>
    </div>
</div>
