<script lang="ts">
    import { auth } from "$lib/runes/auth.svelte";
    import {
        BarChart3,
        TrendingUp,
        Users,
        Wallet,
        Ticket,
        Bus,
        ArrowUpRight,
        ArrowDownRight,
        Calendar,
        Loader2,
        RefreshCw,
    } from "@lucide/svelte";
    import { Button } from "$lib/components/ui/button";
    import { toast } from "svelte-sonner";

    interface DashboardStats {
        todayRevenue: number;
        todayBookings: number;
        activeTrips: number;
        occupancyRate: number;
        revenueTrend: string;
        bookingsTrend: string;
        tripsTrend: string;
        occupancyTrend: string;
    }

    interface RecentBooking {
        id: string;
        route: string;
        seats: number;
        amount: number;
        time: string;
        status: string;
    }

    let stats = $state<DashboardStats>({
        todayRevenue: 0,
        todayBookings: 0,
        activeTrips: 0,
        occupancyRate: 0,
        revenueTrend: "+0%",
        bookingsTrend: "+0%",
        tripsTrend: "0%",
        occupancyTrend: "+0%",
    });

    let recentBookings = $state<RecentBooking[]>([]);
    let isLoading = $state(true);
    let dateRange = $state<"today" | "7d" | "30d" | "custom">("today");

    const statsConfig = [
        {
            label: "Today's Revenue",
            value: () => `৳ ${stats.todayRevenue.toLocaleString()}`,
            icon: Wallet,
            trend: () => stats.revenueTrend,
            color: "text-blue-600",
            bgColor: "bg-blue-100 dark:bg-blue-900/30",
        },
        {
            label: "Bookings",
            value: () => stats.todayBookings.toString(),
            icon: Ticket,
            trend: () => stats.bookingsTrend,
            color: "text-green-600",
            bgColor: "bg-green-100 dark:bg-green-900/30",
        },
        {
            label: "Active Trips",
            value: () => stats.activeTrips.toString(),
            icon: Bus,
            trend: () => stats.tripsTrend,
            color: "text-purple-600",
            bgColor: "bg-purple-100 dark:bg-purple-900/30",
        },
        {
            label: "Occupancy",
            value: () => `${stats.occupancyRate}%`,
            icon: Users,
            trend: () => stats.occupancyTrend,
            color: "text-orange-600",
            bgColor: "bg-orange-100 dark:bg-orange-900/30",
        },
    ];

    async function fetchData() {
        isLoading = true;
        try {
            // Simulate API call - in production, call Reporting Service
            await new Promise((resolve) => setTimeout(resolve, 800));

            // Mock data - replace with real API calls
            stats = {
                todayRevenue: 24500,
                todayBookings: 142,
                activeTrips: 8,
                occupancyRate: 78,
                revenueTrend: "+12%",
                bookingsTrend: "+5%",
                tripsTrend: "0%",
                occupancyTrend: "+2%",
            };

            recentBookings = [
                {
                    id: "ORD-1001",
                    route: "Dhaka → Chittagong",
                    seats: 2,
                    amount: 1600,
                    time: "2m ago",
                    status: "confirmed",
                },
                {
                    id: "ORD-1002",
                    route: "Chittagong → Sylhet",
                    seats: 1,
                    amount: 800,
                    time: "15m ago",
                    status: "confirmed",
                },
                {
                    id: "ORD-1003",
                    route: "Dhaka → Cox's Bazar",
                    seats: 4,
                    amount: 3200,
                    time: "1h ago",
                    status: "pending",
                },
                {
                    id: "ORD-1004",
                    route: "Sylhet → Dhaka",
                    seats: 2,
                    amount: 1600,
                    time: "2h ago",
                    status: "confirmed",
                },
            ];
        } catch (error) {
            toast.error("Failed to load dashboard data");
        } finally {
            isLoading = false;
        }
    }

    function refreshData() {
        toast.info("Refreshing data...");
        fetchData();
    }

    function getTrendColor(trend: string) {
        if (trend.startsWith("+")) return "text-green-500";
        if (trend.startsWith("-")) return "text-red-500";
        return "text-muted-foreground";
    }

    function getTrendIcon(trend: string) {
        if (trend.startsWith("+")) return ArrowUpRight;
        if (trend.startsWith("-")) return ArrowDownRight;
        return TrendingUp;
    }

    $effect(() => {
        fetchData();
    });
</script>

<div class="space-y-8">
    <!-- Header -->
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
            <h1 class="text-3xl font-extrabold tracking-tight text-gradient">
                Dashboard
            </h1>
            <p class="text-muted-foreground mt-1">
                Welcome back, {auth.user?.name || "Operator"}
            </p>
        </div>
        <div class="flex gap-2">
            <div class="flex rounded-lg bg-muted p-1">
                {#each [
                    { key: "today", label: "Today" },
                    { key: "7d", label: "7 Days" },
                    { key: "30d", label: "30 Days" },
                ] as range}
                    <button
                        class="rounded-md px-3 py-1.5 text-sm font-medium transition-all {dateRange === range.key ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground hover:text-foreground'}"
                        onclick={() => (dateRange = range.key)}
                    >
                        {range.label}
                    </button>
                {/each}
            </div>
            <Button variant="outline" size="sm" onclick={refreshData}>
                {#if isLoading}
                    <Loader2 class="mr-1 h-4 w-4 animate-spin" />
                {:else}
                    <RefreshCw class="mr-1 h-4 w-4" />
                {/if}
                Refresh
            </Button>
        </div>
    </div>

    <!-- Loading State -->
    {#if isLoading}
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {#each Array(4) as _}
                <div class="glass-card rounded-xl p-6 animate-pulse">
                    <div class="flex items-center justify-between">
                        <div class="space-y-2">
                            <div class="h-4 w-24 rounded bg-muted"></div>
                            <div class="h-8 w-32 rounded bg-muted"></div>
                        </div>
                        <div class="size-12 rounded-full bg-muted"></div>
                    </div>
                    <div class="mt-4 h-4 w-20 rounded bg-muted"></div>
                </div>
            {/each}
        </div>
    {:else}
        <!-- Stats Grid -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {#each statsConfig as stat}
                <div
                    class="glass-card rounded-xl p-6 transition-all hover:shadow-md"
                >
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-sm font-medium text-muted-foreground">
                                {stat.label}
                            </p>
                            <h3 class="mt-2 text-2xl font-bold">
                                {stat.value()}
                            </h3>
                        </div>
                        <div
                            class="rounded-full p-3 {stat.bgColor} {stat.color}"
                        >
                            <stat.icon size={20} />
                        </div>
                    </div>
                    <div
                        class="mt-4 flex items-center text-xs {getTrendColor(stat.trend())}"
                    >
                        {@const TrendIcon = getTrendIcon(stat.trend())}
                        <TrendIcon size={14} class="mr-1" />
                        <span class="font-medium">{stat.trend()}</span>
                        <span class="ml-1 text-muted-foreground"
                            >from yesterday</span
                        >
                    </div>
                </div>
            {/each}
        </div>
    {/if}

    <!-- Main Content Grid -->
    <div class="grid gap-8 lg:grid-cols-3">
        <!-- Revenue Chart Placeholder -->
        <div class="glass-card col-span-2 rounded-xl p-6 lg:col-span-2">
            <div class="mb-6 flex items-center justify-between">
                <h3 class="text-lg font-bold">Revenue Overview</h3>
                <BarChart3 class="text-muted-foreground" size={20} />
            </div>
            <div
                class="flex h-[300px] items-center justify-center rounded-lg border border-dashed border-border bg-muted/30"
            >
                <div class="text-center">
                    <BarChart3
                        size={48}
                        class="mx-auto mb-3 text-muted-foreground opacity-50"
                    />
                    <p class="text-muted-foreground">
                        Revenue chart requires Chart.js
                    </p>
                    <p class="text-sm text-muted-foreground">
                        Install: pnpm add chart.js
                    </p>
                </div>
            </div>
        </div>

        <!-- Recent Activity -->
        <div class="glass-card rounded-xl p-6">
            <div class="mb-4 flex items-center justify-between">
                <h3 class="text-lg font-bold">Recent Activity</h3>
                <Button variant="ghost" size="sm" href="/orders"
                    >View All</Button
                >
            </div>
            <div class="space-y-4">
                {#if recentBookings.length > 0}
                    {#each recentBookings as booking}
                        <div
                            class="flex items-center gap-3 border-b border-border/50 pb-3 last:border-0 hover:bg-muted/50 p-2 rounded-lg transition-colors"
                        >
                            <div
                                class="flex size-2 shrink-0 items-center justify-center rounded-full {booking.status === 'confirmed' ? 'bg-green-500' : 'bg-yellow-500'}"
                            ></div>
                            <div class="flex-1 min-w-0">
                                <p class="text-sm font-medium truncate">
                                    New Booking #{booking.id}
                                </p>
                                <p class="text-xs text-muted-foreground truncate">
                                    {booking.route} • {booking.seats} Seats •
                                    ৳{booking.amount}
                                </p>
                            </div>
                            <span
                                class="shrink-0 text-xs text-muted-foreground"
                                >{booking.time}</span
                            >
                        </div>
                    {/each}
                {:else}
                    <div class="py-8 text-center text-muted-foreground">
                        <Calendar size={32} class="mx-auto mb-3 opacity-50" />
                        <p>No recent bookings</p>
                    </div>
                {/if}
            </div>
        </div>
    </div>

    <!-- Quick Actions -->
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <a
            href="/organization/operations/trips"
            class="glass-card rounded-xl p-6 transition-all hover:shadow-md hover:scale-[1.02] block"
        >
            <div class="flex items-center gap-3">
                <div class="rounded-full bg-blue-100 p-3 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400">
                    <Calendar size={20} />
                </div>
                <div>
                    <p class="font-semibold">Schedule Trip</p>
                    <p class="text-sm text-muted-foreground">Create new trip</p>
                </div>
            </div>
        </a>
        <a
            href="/organization/sales/counter"
            class="glass-card rounded-xl p-6 transition-all hover:shadow-md hover:scale-[1.02] block"
        >
            <div class="flex items-center gap-3">
                <div class="rounded-full bg-green-100 p-3 text-green-600 dark:bg-green-900/30 dark:text-green-400">
                    <Ticket size={20} />
                </div>
                <div>
                    <p class="font-semibold">Counter Sales</p>
                    <p class="text-sm text-muted-foreground">Walk-in booking</p>
                </div>
            </div>
        </a>
        <a
            href="/organization/operations/fleet"
            class="glass-card rounded-xl p-6 transition-all hover:shadow-md hover:scale-[1.02] block"
        >
            <div class="flex items-center gap-3">
                <div class="rounded-full bg-purple-100 p-3 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400">
                    <Bus size={20} />
                </div>
                <div>
                    <p class="font-semibold">Fleet Management</p>
                    <p class="text-sm text-muted-foreground">Manage vehicles</p>
                </div>
            </div>
        </a>
        <a
            href="/organization/finance"
            class="glass-card rounded-xl p-6 transition-all hover:shadow-md hover:scale-[1.02] block"
        >
            <div class="flex items-center gap-3">
                <div class="rounded-full bg-orange-100 p-3 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400">
                    <Wallet size={20} />
                </div>
                <div>
                    <p class="font-semibold">Finance</p>
                    <p class="text-sm text-muted-foreground">Revenue & payouts</p>
                </div>
            </div>
        </a>
    </div>
</div>
