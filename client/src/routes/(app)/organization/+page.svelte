<script lang="ts">
    import { auth } from "$lib/runes/auth.svelte";
    import { BarChart3, TrendingUp, Users, Wallet } from "lucide-svelte";

    const stats = [
        {
            label: "Today's Revenue",
            value: "৳ 24,500",
            icon: Wallet,
            trend: "+12%",
        },
        { label: "Bookings", value: "142", icon: TicketIcon, trend: "+5%" }, // Using local TicketIcon definition or lucide
        { label: "Active Trips", value: "8", icon: BusIcon, trend: "0%" },
        { label: "Occupancy", value: "78%", icon: Users, trend: "+2%" },
    ];

    import { Ticket as TicketIcon, Bus as BusIcon } from "lucide-svelte";
</script>

<div class="space-y-8">
    <!-- Header -->
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-extrabold tracking-tight text-gradient">
                Dashboard
            </h1>
            <p class="text-muted-foreground mt-1">
                Welcome back, {auth.user?.name}
            </p>
        </div>
        <div class="flex gap-2">
            <button
                class="h-9 rounded-md bg-primary/10 px-4 text-sm font-medium text-primary hover:bg-primary/20"
            >
                Refresh Data
            </button>
        </div>
    </div>

    <!-- Quick Stats -->
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {#each stats as stat}
            <div
                class="glass-card rounded-xl p-6 transition-all hover:scale-[1.02]"
            >
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-sm font-medium text-muted-foreground">
                            {stat.label}
                        </p>
                        <h3 class="mt-2 text-2xl font-bold">{stat.value}</h3>
                    </div>
                    <div class="rounded-full bg-primary/10 p-3 text-primary">
                        <stat.icon size={20} />
                    </div>
                </div>
                <div class="mt-4 flex items-center text-xs text-green-500">
                    <TrendingUp size={14} class="mr-1" />
                    <span class="font-medium">{stat.trend}</span>
                    <span class="ml-1 text-muted-foreground"
                        >from yesterday</span
                    >
                </div>
            </div>
        {/each}
    </div>

    <!-- Main Content Grid -->
    <div class="grid gap-8 lg:grid-cols-3">
        <!-- Revenue Chart Placeholder -->
        <div class="glass-card col-span-2 rounded-xl p-6 lg:col-span-2">
            <div class="mb-6 flex items-center justify-between">
                <h3 class="text-lg font-bold">Revenue Overview</h3>
                <BarChart3 class="text-muted-foreground" size={20} />
            </div>
            <div
                class="flex h-[300px] items-center justify-center rounded-lg border border-dashed border-gray-300 dark:border-gray-700 bg-black/5 dark:bg-white/5"
            >
                <p class="text-muted-foreground">
                    Revenue Chart Component Placeholder
                </p>
            </div>
        </div>

        <!-- Recent Activity Placeholder -->
        <div class="glass-card rounded-xl p-6">
            <h3 class="mb-4 text-lg font-bold">Recent Activity</h3>
            <div class="space-y-4">
                {#each [1, 2, 3, 4] as i}
                    <div
                        class="flex items-center gap-3 border-b border-white/5 pb-3 last:border-0 hover:bg-white/5 p-2 rounded-lg transition-colors"
                    >
                        <div class="h-2 w-2 rounded-full bg-green-500"></div>
                        <div class="flex-1">
                            <p class="text-sm font-medium">
                                New Booking #ORD-{1000 + i}
                            </p>
                            <p class="text-xs text-muted-foreground">
                                Dhaka to Chittagong • 2 Seats
                            </p>
                        </div>
                        <span class="text-xs text-muted-foreground">2m ago</span
                        >
                    </div>
                {/each}
            </div>
        </div>
    </div>
</div>
