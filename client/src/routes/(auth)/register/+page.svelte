<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { auth } from "$lib/runes/auth.svelte";
    import { goto } from "$app/navigation";
    import { Loader2, User, Bus } from "@lucide/svelte";
    import { toast } from "svelte-sonner";
    import { cn } from "$lib/utils";

    let name = $state("");
    let email = $state("");
    let password = $state("");
    let confirmPassword = $state("");
    let orgName = $state("");
    let accountType = $state<"traveller" | "operator">("traveller");

    let passwordError = $derived(
        confirmPassword && password !== confirmPassword
            ? "Passwords do not match"
            : null,
    );

    let isFormValid = $derived(
        name &&
            email &&
            password &&
            confirmPassword &&
            !passwordError &&
            (accountType === "traveller" ||
                (accountType === "operator" && orgName)),
    );

    async function handleRegister() {
        if (!isFormValid) return;

        // Use name as org name if not provided (fallback, though validation enforces it for operator)
        // If traveller, orgName should be undefined/empty to avoid creating org.
        const organizationName =
            accountType === "operator" ? orgName : undefined;

        const success = await auth.register(email, password, organizationName);
        if (success) {
            toast.success("Account created!", {
                description: "Please sign in with your credentials.",
            });
            // Redirect to login after successful registration
            goto("/login?registered=true");
        } else {
            toast.error("Registration failed", {
                description: auth.error || "Please try again.",
            });
        }
    }

    function handleKeyDown(event: KeyboardEvent) {
        if (event.key === "Enter" && isFormValid) {
            handleRegister();
        }
    }
</script>

<div class="glass-panel w-full p-8 relative overflow-hidden">
    <div
        class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-purple-500 to-pink-500"
    ></div>

    <div class="flex flex-col gap-6 text-center">
        <div>
            <h1 class="text-3xl font-black tracking-tight mb-2">
                Join TicketNation
            </h1>
            <p class="text-muted-foreground">
                {accountType === "traveller"
                    ? "Create your traveller account"
                    : "Register your bus company"}
            </p>
        </div>

        {#if auth.error}
            <div
                class="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-3 text-sm text-red-600 dark:text-red-400"
            >
                {auth.error}
            </div>
        {/if}

        <!-- Account Type Selector -->
        <div class="grid grid-cols-2 gap-2 p-1 bg-muted/50 rounded-lg">
            <button
                class={cn(
                    "flex items-center justify-center gap-2 py-2 text-sm font-bold rounded-md transition-all",
                    accountType === "traveller"
                        ? "bg-white text-primary shadow-sm dark:bg-gray-800 dark:text-white"
                        : "text-muted-foreground hover:text-foreground",
                )}
                onclick={() => (accountType = "traveller")}
            >
                <User size={16} />
                Traveller
            </button>
            <button
                class={cn(
                    "flex items-center justify-center gap-2 py-2 text-sm font-bold rounded-md transition-all",
                    accountType === "operator"
                        ? "bg-white text-primary shadow-sm dark:bg-gray-800 dark:text-white"
                        : "text-muted-foreground hover:text-foreground",
                )}
                onclick={() => (accountType = "operator")}
            >
                <Bus size={16} />
                Operator
            </button>
        </div>

        <div class="flex flex-col gap-4 text-left">
            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    for="name">Full Name</label
                >
                <Input
                    id="name"
                    type="text"
                    bind:value={name}
                    class="bg-white/50 backdrop-blur-sm"
                    placeholder="John Doe"
                    disabled={auth.isLoading}
                    onkeydown={handleKeyDown}
                />
            </div>

            {#if accountType === "operator"}
                <div class="space-y-2">
                    <label
                        class="text-sm font-bold text-gray-700 dark:text-gray-300"
                        for="orgName">Company Name</label
                    >
                    <Input
                        id="orgName"
                        type="text"
                        bind:value={orgName}
                        class="bg-white/50 backdrop-blur-sm"
                        placeholder="Green Line Paribahan"
                        disabled={auth.isLoading}
                        onkeydown={handleKeyDown}
                    />
                </div>
            {/if}

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
            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    for="confirmPassword">Confirm Password</label
                >
                <Input
                    id="confirmPassword"
                    type="password"
                    bind:value={confirmPassword}
                    class="bg-white/50 backdrop-blur-sm {passwordError
                        ? 'border-red-500'
                        : ''}"
                    placeholder="••••••••"
                    disabled={auth.isLoading}
                    onkeydown={handleKeyDown}
                />
                {#if passwordError}
                    <p class="text-xs text-red-500">{passwordError}</p>
                {/if}
            </div>
        </div>

        <Button
            class="w-full h-12 text-lg font-bold shadow-lg shadow-purple-500/20 bg-purple-600 hover:bg-purple-700"
            onclick={handleRegister}
            disabled={auth.isLoading || !isFormValid}
        >
            {#if auth.isLoading}
                <Loader2 class="mr-2 h-5 w-5 animate-spin" />
                Creating account...
            {:else}
                {accountType === "operator"
                    ? "Register Company"
                    : "Create Account"}
            {/if}
        </Button>

        <p class="text-sm text-gray-500">
            Already have an account? <a
                href="/login"
                class="font-bold text-primary hover:underline">Sign In</a
            >
        </p>
    </div>
</div>
