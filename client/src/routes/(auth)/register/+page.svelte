<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { auth } from "$lib/runes/auth.svelte";
    import { goto } from "$app/navigation";
    import { Loader2, User, Bus, Check, X, AlertCircle } from "@lucide/svelte";
    import { toast } from "svelte-sonner";
    import { cn } from "$lib/utils";
    import { z } from "zod";

    let name = $state("");
    let email = $state("");
    let password = $state("");
    let confirmPassword = $state("");
    let orgName = $state("");
    let orgAddress = $state("");
    let orgPhone = $state("");
    let orgWebsite = $state("");
    let accountType = $state<"traveller" | "operator">("traveller");

    let errors = $state<Record<string, string[]>>({});
    let touched = $state<Record<string, boolean>>({});

    const phoneRegex = /^(\+88)?01[3-9]\d{8}$/;

    // Dynamic Schema based on accountType (or just huge schema with refinements)
    // We'll validate on submit or blur.

    function getSchema() {
        const base = z.object({
            name: z.string().min(2, "Name must be at least 2 characters"),
            email: z.string().email("Invalid email address"),
            password: z
                .string()
                .min(8, "Password must be at least 8 characters")
                .regex(/[A-Z]/, "Must contain an uppercase letter")
                .regex(/[a-z]/, "Must contain a lowercase letter")
                .regex(/[0-9]/, "Must contain a number")
                .regex(/[^A-Za-z0-9]/, "Must contain a special character"),
            confirmPassword: z.string(),
        });

        if (accountType === "operator") {
            return base
                .extend({
                    orgName: z.string().min(2, "Company Name is required"),
                    orgAddress: z.string().min(5, "Address is required"),
                    orgPhone: z
                        .string()
                        .regex(
                            phoneRegex,
                            "Invalid Bangladeshi phone number (+8801...)",
                        ),
                    orgWebsite: z
                        .string()
                        .url("Invalid URL")
                        .optional()
                        .or(z.literal("")),
                })
                .refine((data) => data.password === data.confirmPassword, {
                    message: "Passwords do not match",
                    path: ["confirmPassword"],
                });
        }

        return base.refine((data) => data.password === data.confirmPassword, {
            message: "Passwords do not match",
            path: ["confirmPassword"],
        });
    }

    async function validateField(field: string) {
        touched[field] = true;
        const schema = getSchema();
        const formData = {
            name,
            email,
            password,
            confirmPassword,
            orgName,
            orgAddress,
            orgPhone,
            orgWebsite,
        };

        // We parse entire schema to catch refinement errors (like confirmPassword)
        // Check strict? No, safeParse.
        const result = schema.safeParse(formData);

        if (!result.success) {
            const formatted = result.error.flatten().fieldErrors;
            errors = formatted;
        } else {
            errors = {};
        }
    }

    // Password Strength Calc
    let passwordStrength = $derived.by(() => {
        let score = 0;
        if (!password) return 0;
        if (password.length >= 8) score += 20;
        if (/[A-Z]/.test(password)) score += 20;
        if (/[a-z]/.test(password)) score += 20;
        if (/[0-9]/.test(password)) score += 20;
        if (/[^A-Za-z0-9]/.test(password)) score += 20;
        return score;
    });

    async function handleRegister() {
        const schema = getSchema();
        const formData = {
            name,
            email,
            password,
            confirmPassword,
            orgName,
            orgAddress,
            orgPhone,
            orgWebsite,
        };
        const result = schema.safeParse(formData);

        if (!result.success) {
            errors = result.error.flatten().fieldErrors;
            toast.error("Please fix the errors in the form");
            return;
        }

        errors = {}; // Clear errors

        // Use name as org name if not provided (fallback, though validation enforces it for operator)
        // If traveller, orgName should be undefined/empty to avoid creating org.
        const organizationName =
            accountType === "operator" ? orgName : undefined;

        const orgDetails =
            accountType === "operator"
                ? {
                      address: orgAddress,
                      phone: orgPhone,
                      website: orgWebsite,
                      email: email, // Use user email as contact email for now
                  }
                : {};

        const success = await auth.register(
            email,
            password,
            name,
            organizationName,
            orgDetails,
        );
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
        if (event.key === "Enter") {
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
            <h1 class="text-3xl font-black tracking-tight mb-2">Join Travio</h1>
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
                <div class="relative">
                    <Input
                        id="name"
                        type="text"
                        bind:value={name}
                        class={cn(
                            "bg-white/50 backdrop-blur-sm transition-all",
                            touched.name && errors.name
                                ? "border-red-500 ring-red-500/20"
                                : "",
                        )}
                        placeholder="John Doe"
                        disabled={auth.isLoading}
                        onblur={() => validateField("name")}
                        oninput={() => validateField("name")}
                        onkeydown={handleKeyDown}
                    />
                    {#if touched.name && !errors.name && name}
                        <Check
                            class="absolute right-3 top-1/2 -translate-y-1/2 text-green-500"
                            size={16}
                        />
                    {/if}
                </div>
                {#if touched.name && errors.name}
                    <p
                        class="text-xs text-red-500 flex items-center gap-1 mt-1 animate-in slide-in-from-top-1"
                    >
                        <AlertCircle size={12} />
                        {errors.name[0]}
                    </p>
                {/if}
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
                        class={cn(
                            "bg-white/50 backdrop-blur-sm",
                            touched.orgName && errors.orgName
                                ? "border-red-500"
                                : "",
                        )}
                        placeholder="Green Line Paribahan"
                        disabled={auth.isLoading}
                        onblur={() => validateField("orgName")}
                        oninput={() => validateField("orgName")}
                        onkeydown={handleKeyDown}
                    />
                    {#if touched.orgName && errors.orgName}
                        <p class="text-xs text-red-500 mt-1">
                            {errors.orgName[0]}
                        </p>
                    {/if}
                </div>
                <div class="space-y-2">
                    <label
                        class="text-sm font-bold text-gray-700 dark:text-gray-300"
                        for="orgAddress">Company Address</label
                    >
                    <Input
                        id="orgAddress"
                        type="text"
                        bind:value={orgAddress}
                        class={cn(
                            "bg-white/50 backdrop-blur-sm",
                            touched.orgAddress && errors.orgAddress
                                ? "border-red-500"
                                : "",
                        )}
                        placeholder="123 Example Street, Dhaka"
                        disabled={auth.isLoading}
                        onblur={() => validateField("orgAddress")}
                        oninput={() => validateField("orgAddress")}
                        onkeydown={handleKeyDown}
                    />
                    {#if touched.orgAddress && errors.orgAddress}
                        <p class="text-xs text-red-500 mt-1">
                            {errors.orgAddress[0]}
                        </p>
                    {/if}
                </div>
                <div class="grid grid-cols-2 gap-4">
                    <div class="space-y-2">
                        <label
                            class="text-sm font-bold text-gray-700 dark:text-gray-300"
                            for="orgPhone">Contact Phone</label
                        >
                        <Input
                            id="orgPhone"
                            type="tel"
                            bind:value={orgPhone}
                            class={cn(
                                "bg-white/50 backdrop-blur-sm",
                                touched.orgPhone && errors.orgPhone
                                    ? "border-red-500"
                                    : "",
                            )}
                            placeholder="+880 17..."
                            disabled={auth.isLoading}
                            onblur={() => validateField("orgPhone")}
                            oninput={() => validateField("orgPhone")}
                            onkeydown={handleKeyDown}
                        />
                        {#if touched.orgPhone && errors.orgPhone}
                            <p class="text-xs text-red-500 mt-1">
                                {errors.orgPhone[0]}
                            </p>
                        {/if}
                    </div>
                    <div class="space-y-2">
                        <label
                            class="text-sm font-bold text-gray-700 dark:text-gray-300"
                            for="orgWebsite">Website (Optional)</label
                        >
                        <Input
                            id="orgWebsite"
                            type="url"
                            bind:value={orgWebsite}
                            class={cn(
                                "bg-white/50 backdrop-blur-sm",
                                touched.orgWebsite && errors.orgWebsite
                                    ? "border-red-500"
                                    : "",
                            )}
                            placeholder="https://..."
                            disabled={auth.isLoading}
                            onblur={() => validateField("orgWebsite")}
                            oninput={() => validateField("orgWebsite")}
                            onkeydown={handleKeyDown}
                        />
                        {#if touched.orgWebsite && errors.orgWebsite}
                            <p class="text-xs text-red-500 mt-1">
                                {errors.orgWebsite[0]}
                            </p>
                        {/if}
                    </div>
                </div>
            {/if}

            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    for="email">Email</label
                >
                <div class="relative">
                    <Input
                        id="email"
                        type="email"
                        bind:value={email}
                        class={cn(
                            "bg-white/50 backdrop-blur-sm",
                            touched.email && errors.email
                                ? "border-red-500"
                                : "",
                        )}
                        placeholder="you@example.com"
                        disabled={auth.isLoading}
                        onblur={() => validateField("email")}
                        oninput={() => validateField("email")}
                        onkeydown={handleKeyDown}
                    />
                    {#if touched.email && !errors.email && email}
                        <Check
                            class="absolute right-3 top-1/2 -translate-y-1/2 text-green-500"
                            size={16}
                        />
                    {/if}
                </div>
                {#if touched.email && errors.email}
                    <p class="text-xs text-red-500 mt-1">{errors.email[0]}</p>
                {/if}
            </div>

            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    for="password">Password</label
                >
                <div class="relative">
                    <Input
                        id="password"
                        type="password"
                        bind:value={password}
                        class={cn(
                            "bg-white/50 backdrop-blur-sm pr-10",
                            touched.password && errors.password
                                ? "border-red-500"
                                : "",
                        )}
                        placeholder="••••••••"
                        disabled={auth.isLoading}
                        onblur={() => validateField("password")}
                        oninput={() => validateField("password")}
                        onkeydown={handleKeyDown}
                    />
                </div>
                {#if touched.password && errors.password}
                    <p class="text-xs text-red-500 mt-1">
                        {errors.password[0]}
                    </p>
                {/if}
                <!-- Password Strength Meter -->
                {#if password}
                    <div class="mt-2 space-y-1">
                        <div
                            class="flex justify-between text-xs text-muted-foreground"
                        >
                            <span>Strength</span>
                            <span
                                class={cn(
                                    passwordStrength < 40
                                        ? "text-red-500"
                                        : passwordStrength < 80
                                          ? "text-yellow-500"
                                          : "text-green-500",
                                )}
                            >
                                {passwordStrength < 40
                                    ? "Weak"
                                    : passwordStrength < 80
                                      ? "Medium"
                                      : "Strong"}
                            </span>
                        </div>
                        <div
                            class="h-1.5 w-full bg-muted/50 rounded-full overflow-hidden"
                        >
                            <div
                                class={cn(
                                    "h-full transition-all duration-500",
                                    passwordStrength < 40
                                        ? "bg-red-500"
                                        : passwordStrength < 80
                                          ? "bg-yellow-500"
                                          : "bg-green-500",
                                )}
                                style="width: {passwordStrength}%"
                            ></div>
                        </div>
                    </div>
                {/if}
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
                    class={cn(
                        "bg-white/50 backdrop-blur-sm",
                        touched.confirmPassword && errors.confirmPassword
                            ? "border-red-500"
                            : "",
                    )}
                    placeholder="••••••••"
                    disabled={auth.isLoading}
                    onblur={() => validateField("confirmPassword")}
                    oninput={() => validateField("confirmPassword")}
                    onkeydown={handleKeyDown}
                />
                {#if touched.confirmPassword && errors.confirmPassword}
                    <p class="text-xs text-red-500 mt-1">
                        {errors.confirmPassword[0]}
                    </p>
                {/if}
            </div>
        </div>

        <Button
            class="w-full h-12 text-lg font-bold shadow-lg shadow-purple-500/20 bg-purple-600 hover:bg-purple-700"
            onclick={handleRegister}
            disabled={auth.isLoading}
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
