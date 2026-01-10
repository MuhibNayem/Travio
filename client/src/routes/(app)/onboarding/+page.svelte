<script lang="ts">
    import { auth } from "$lib/runes/auth.svelte";
    import { subscriptionApi, type Plan } from "$lib/api/subscription";
    import { Button } from "$lib/components/ui/button";
    import { toast } from "svelte-sonner";
    import {
        Check,
        Bus,
        ShieldCheck,
        Zap,
        Users,
        ArrowRight,
        Loader2,
    } from "@lucide/svelte";
    import { goto } from "$app/navigation";
    import { onMount } from "svelte";

    let step = $state(1);
    let plans = $state<Plan[]>([]);
    let isLoading = $state(true);
    let isProcessing = $state(false);
    let selectedPlanId = $state<string | null>(null);

    const STEPS = [
        { id: 1, title: "Welcome" },
        { id: 2, title: "Select Plan" },
        { id: 3, title: "Success" },
    ];

    onMount(async () => {
        try {
            plans = await subscriptionApi.listPlans();
            // Pre-select popular plan if needed, e.g. 'Goti'
            const goti = plans.find((p) => p.name.includes("Goti"));
            if (goti) selectedPlanId = goti.id;
        } catch (error) {
            console.error("Failed to load plans", error);
            toast.error("Failed to load subscription plans");
        } finally {
            isLoading = false;
        }
    });

    async function handleSelectPlan(plan: Plan) {
        selectedPlanId = plan.id;
    }

    async function handleSubscribe() {
        if (!selectedPlanId || !auth.user?.organizationId) return;
        isProcessing = true;

        try {
            // First check if current plan is 'free'.
            // In a real flow we might need to Upgrade/Swap.
            // For MVP: We try to create subscription.
            // If checking existing sub fails or conflicts, we might need a cancel flow.
            // Assuming Upgrade logic:
            // 1. Cancel current (if any)
            try {
                await subscriptionApi.cancelSubscription(
                    auth.user.organizationId,
                );
            } catch (ignore) {
                // If no sub or failed to cancel (maybe already canceled), ignore
            }

            // 2. Create new
            await subscriptionApi.createSubscription(
                auth.user.organizationId,
                selectedPlanId,
            );

            step = 3;
            toast.success("Subscription activated successfully!");
        } catch (error: any) {
            console.error("Subscription failed", error);
            // Handle specific errors like 'already has active subscription'
            toast.error(
                "Failed to upgrade plan: " + (error.message || "Unknown error"),
            );
        } finally {
            isProcessing = false;
        }
    }

    function formatPrice(paisa: number) {
        return (paisa / 100)
            .toLocaleString("en-BD", { style: "currency", currency: "BDT" })
            .replace("BDT", "৳");
    }
</script>

<div class="min-h-screen bg-muted/20 pb-20 pt-20">
    <div class="container mx-auto max-w-5xl px-4">
        <!-- Progress -->
        <div class="mb-12 flex items-center justify-center gap-4">
            {#each STEPS as s}
                <div class="flex items-center gap-2">
                    <div
                        class={`flex h-8 w-8 items-center justify-center rounded-full text-xs font-bold transition-all ${
                            step >= s.id
                                ? "bg-primary text-primary-foreground scale-110 shadow-lg shadow-primary/20"
                                : "bg-muted text-muted-foreground"
                        }`}
                    >
                        {#if step > s.id}
                            <Check size={14} />
                        {:else}
                            {s.id}
                        {/if}
                    </div>
                    <span
                        class={`text-sm font-medium ${
                            step >= s.id
                                ? "text-foreground"
                                : "text-muted-foreground"
                        }`}
                    >
                        {s.title}
                    </span>
                    {#if s.id < STEPS.length}
                        <div class="h-[1px] w-12 bg-border mx-2"></div>
                    {/if}
                </div>
            {/each}
        </div>

        {#if step === 1}
            <!-- Step 1: Welcome -->
            <div
                class="glass-panel mx-auto max-w-2xl p-10 text-center animate-in fade-in slide-in-from-bottom-4 duration-500"
            >
                <div
                    class="mb-6 inline-flex h-20 w-20 items-center justify-center rounded-3xl bg-primary/10 text-primary"
                >
                    <Bus size={40} />
                </div>
                <h1 class="mb-4 text-3xl font-black tracking-tight">
                    Welcome to TicketNation
                </h1>
                <p class="text-lg text-muted-foreground mb-8">
                    You've successfully created your operator account. Let's get
                    your business set up with the right tools to grow.
                </p>
                <div class="flex flex-col gap-4">
                    <Button
                        size="lg"
                        class="w-full text-lg h-12 font-bold"
                        onclick={() => (step = 2)}
                    >
                        Set Up Subscription <ArrowRight class="ml-2 size-5" />
                    </Button>
                    <Button variant="ghost" onclick={() => goto("/dashboard")}
                        >Skip for now (Free Tier)</Button
                    >
                </div>
            </div>
        {:else if step === 2}
            <!-- Step 2: Plans -->
            <div class="animate-in fade-in slide-in-from-bottom-4 duration-500">
                <div class="text-center mb-10">
                    <h2 class="text-3xl font-bold mb-2">Choose your plan</h2>
                    <p class="text-muted-foreground">
                        Select the package that fits your fleet best
                    </p>
                </div>

                {#if isLoading}
                    <div class="flex justify-center py-20">
                        <Loader2 class="animate-spin text-primary size-10" />
                    </div>
                {:else}
                    <div class="grid gap-8 md:grid-cols-3">
                        {#each plans as plan}
                            <button
                                class={`group relative flex flex-col rounded-2xl border-2 bg-card p-6 text-left transition-all hover:border-primary/50 hover:shadow-2xl ${
                                    selectedPlanId === plan.id
                                        ? "border-primary shadow-xl shadow-primary/10 ring-1 ring-primary"
                                        : "border-border/50 shadow-sm"
                                }`}
                                onclick={() => handleSelectPlan(plan)}
                            >
                                {#if plan.name.toLowerCase().includes("goti")}
                                    <div
                                        class="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-gradient-to-r from-orange-500 to-pink-500 px-3 py-1 text-[10px] font-bold uppercase tracking-wider text-white shadow-lg"
                                    >
                                        Most Popular
                                    </div>
                                {/if}

                                <h3 class="text-lg font-bold">{plan.name}</h3>
                                <p
                                    class="mb-4 text-sm text-muted-foreground line-clamp-2"
                                >
                                    {plan.description}
                                </p>

                                <div class="mb-6 flex items-baseline gap-1">
                                    <span class="text-3xl font-black"
                                        >{formatPrice(plan.price)}</span
                                    >
                                    <span
                                        class="text-sm font-medium text-muted-foreground"
                                        >/mo</span
                                    >
                                </div>

                                <div class="mb-6 flex-1 space-y-3">
                                    {#if plan.max_users < 1000}
                                        <div
                                            class="flex items-center gap-2 text-sm"
                                        >
                                            <Users
                                                size={16}
                                                class="text-primary"
                                            />
                                            <span
                                                >Max {plan.max_users} staff</span
                                            >
                                        </div>
                                    {/if}
                                    {#each Object.entries(plan.features) as [feature, val]}
                                        <div
                                            class="flex items-center gap-2 text-sm text-muted-foreground/80"
                                        >
                                            <Check
                                                size={16}
                                                class="text-green-500"
                                            />
                                            <span class="capitalize"
                                                >{feature.replace(/_/g, " ")}: {val}</span
                                            >
                                        </div>
                                    {/each}
                                </div>

                                <div
                                    class={`mt-auto h-2 w-full rounded-full bg-primary/10 transition-all ${selectedPlanId === plan.id ? "opacity-100" : "opacity-0"}`}
                                >
                                    <div
                                        class="h-full w-full rounded-full bg-primary"
                                    ></div>
                                </div>
                            </button>
                        {/each}
                    </div>

                    <div class="mt-10 flex justify-end">
                        <Button
                            size="lg"
                            class="w-full md:w-auto text-lg h-14 font-bold px-10 shadow-xl shadow-primary/20"
                            disabled={!selectedPlanId || isProcessing}
                            onclick={handleSubscribe}
                        >
                            {#if isProcessing}
                                <Loader2 class="mr-2 animate-spin" /> Upgrading...
                            {:else}
                                Activate Plan
                            {/if}
                        </Button>
                    </div>
                {/if}
            </div>
        {:else if step === 3}
            <!-- Step 3: Success -->
            <div
                class="glass-panel mx-auto max-w-lg p-12 text-center animate-in zoom-in-50 duration-500"
            >
                <div
                    class="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-full bg-green-100 text-green-600 dark:bg-green-900/30"
                >
                    <Check size={40} />
                </div>
                <h2 class="mb-2 text-2xl font-black">All Set!</h2>
                <p class="mb-8 text-muted-foreground">
                    Your subscription has been activated. You can now access all
                    features of your selected plan.
                </p>
                <Button
                    size="lg"
                    class="w-full font-bold"
                    onclick={() => goto("/dashboard")}
                >
                    Go to Dashboard
                </Button>
            </div>
        {/if}
    </div>
</div>
