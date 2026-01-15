<script lang="ts">
    import { onMount } from "svelte";
    import {
        reportingApi,
        type OrganizationMetrics,
        type RevenueData,
        type TopRouteData,
    } from "$lib/api/reporting";
    import { Button } from "$lib/components/ui/button";
    import * as Card from "$lib/components/ui/card";
    import * as Table from "$lib/components/ui/table";
    import {
        ArrowUpRight,
        CreditCard,
        DollarSign,
        Download,
        TrendingUp,
        Users,
        Activity,
        AlertCircle,
        ShoppingBag,
    } from "@lucide/svelte";
    import { toast } from "svelte-sonner";
    import { page } from "$app/stores";

    let metrics = $state<OrganizationMetrics | null>(null);
    let revenueData = $state<RevenueData[]>([]);
    let topRoutes = $state<TopRouteData[]>([]);
    let loading = $state(true);

    async function loadData() {
        loading = true;
        try {
            // Parallel fetch
            const [metricsRes, revenueRes, routesRes] = await Promise.all([
                reportingApi.getOrganizationMetrics({}),
                reportingApi.getRevenueReport({ limit: 10 }), // Last 10 days
                reportingApi.getTopRoutes({ limit: 5 }),
            ]);

            metrics = metricsRes;
            revenueData = revenueRes.data || [];
            topRoutes = routesRes.data || [];
        } catch (e) {
            console.error("Failed to load finance data", e);
            toast.error("Failed to refresh reports. Backend might be empty.");
        } finally {
            loading = false;
        }
    }

    async function exportReport() {
        try {
            const orgId = $page.data.user?.organizationId;
            const url = `/api/reports/export?organization_id=${orgId}&type=revenue&format=csv`;
            // Trigger download
            window.open(url, "_blank");
            toast.success("Export started");
        } catch (e) {
            toast.error("Export failed");
        }
    }

    onMount(() => {
        loadData();
    });

    function formatCurrency(amount: string | number) {
        const val = typeof amount === "string" ? parseInt(amount) : amount;
        return new Intl.NumberFormat("en-BD", {
            style: "currency",
            currency: "BDT",
            minimumFractionDigits: 0,
        }).format(val / 100);
    }
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Finance</h1>
            <p class="text-muted-foreground mt-2">
                Revenue analysis and financial performance.
            </p>
        </div>
        <div class="flex gap-2">
            <Button variant="outline" onclick={loadData} disabled={loading}>
                Refresh
            </Button>
            <Button onclick={exportReport}>
                <Download class="mr-2 h-4 w-4" />
                Export CSV
            </Button>
        </div>
    </div>

    <!-- Summary Cards -->
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card.Root>
            <Card.Header
                class="flex flex-row items-center justify-between space-y-0 pb-2"
            >
                <Card.Title class="text-sm font-medium">
                    Total Revenue
                </Card.Title>
                <DollarSign class="h-4 w-4 text-muted-foreground" />
            </Card.Header>
            <Card.Content>
                <div class="text-2xl font-bold">
                    {metrics ? formatCurrency(metrics.total_revenue) : "..."}
                </div>
                <p class="text-xs text-muted-foreground">
                    Lifetime gross revenue
                </p>
            </Card.Content>
        </Card.Root>
        <Card.Root>
            <Card.Header
                class="flex flex-row items-center justify-between space-y-0 pb-2"
            >
                <Card.Title class="text-sm font-medium">Total Orders</Card.Title
                >
                <ShoppingBag class="h-4 w-4 text-muted-foreground" />
            </Card.Header>
            <Card.Content>
                <div class="text-2xl font-bold">
                    {metrics ? metrics.total_orders : "..."}
                </div>
                <p class="text-xs text-muted-foreground">Confirmed bookings</p>
            </Card.Content>
        </Card.Root>
        <Card.Root>
            <Card.Header
                class="flex flex-row items-center justify-between space-y-0 pb-2"
            >
                <Card.Title class="text-sm font-medium">
                    Avg. Order Value
                </Card.Title>
                <Activity class="h-4 w-4 text-muted-foreground" />
            </Card.Header>
            <Card.Content>
                <div class="text-2xl font-bold">
                    {metrics
                        ? formatCurrency(metrics.avg_order_value * 100)
                        : "..."}
                </div>
                <p class="text-xs text-muted-foreground">
                    Per transaction average
                </p>
            </Card.Content>
        </Card.Root>
        <Card.Root>
            <Card.Header
                class="flex flex-row items-center justify-between space-y-0 pb-2"
            >
                <Card.Title class="text-sm font-medium">
                    Cancellation Rate
                </Card.Title>
                <AlertCircle class="h-4 w-4 text-muted-foreground" />
            </Card.Header>
            <Card.Content>
                <div class="text-2xl font-bold">
                    {metrics
                        ? (metrics.cancellation_rate * 100).toFixed(1)
                        : "0"}%
                </div>
                <p class="text-xs text-muted-foreground">
                    Orders cancelled by users
                </p>
            </Card.Content>
        </Card.Root>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <!-- Revenue Report Table -->
        <Card.Root class="col-span-4">
            <Card.Header>
                <Card.Title>Recent Revenue</Card.Title>
                <Card.Description>
                    Daily revenue breakdown for the last 30 days.
                </Card.Description>
            </Card.Header>
            <Card.Content>
                <Table.Root>
                    <Table.Header>
                        <Table.Row>
                            <Table.Head>Date</Table.Head>
                            <Table.Head>Orders</Table.Head>
                            <Table.Head class="text-right">Revenue</Table.Head>
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        {#if loading}
                            <Table.Row>
                                <Table.Cell colspan={3} class="h-24 text-center"
                                    >Loading...</Table.Cell
                                >
                            </Table.Row>
                        {:else if revenueData.length === 0}
                            <Table.Row>
                                <Table.Cell colspan={3} class="h-24 text-center"
                                    >No data available</Table.Cell
                                >
                            </Table.Row>
                        {:else}
                            {#each revenueData as row}
                                <Table.Row>
                                    <Table.Cell>
                                        {new Date(
                                            row.date,
                                        ).toLocaleDateString()}
                                    </Table.Cell>
                                    <Table.Cell>{row.order_count}</Table.Cell>
                                    <Table.Cell class="text-right">
                                        {formatCurrency(
                                            row.total_revenue_paisa,
                                        )}
                                    </Table.Cell>
                                </Table.Row>
                            {/each}
                        {/if}
                    </Table.Body>
                </Table.Root>
            </Card.Content>
        </Card.Root>

        <!-- Top Routes -->
        <Card.Root class="col-span-3">
            <Card.Header>
                <Card.Title>Top Routes</Card.Title>
                <Card.Description>
                    Best performing routes by revenue.
                </Card.Description>
            </Card.Header>
            <Card.Content>
                <div class="space-y-8">
                    {#if loading}
                        <div class="text-center py-4 text-muted-foreground">
                            Loading...
                        </div>
                    {:else if topRoutes.length === 0}
                        <div class="text-center py-4 text-muted-foreground">
                            No route performance data
                        </div>
                    {:else}
                        {#each topRoutes as route}
                            <div class="flex items-center">
                                <div class="ml-4 space-y-1">
                                    <p
                                        class="text-sm font-medium leading-none truncate max-w-[150px]"
                                    >
                                        {route.route_name}
                                    </p>
                                    <p class="text-xs text-muted-foreground">
                                        {(route.avg_occupancy * 100).toFixed(
                                            0,
                                        )}% Occupancy
                                    </p>
                                </div>
                                <div class="ml-auto font-medium">
                                    {formatCurrency(route.revenue)}
                                </div>
                            </div>
                        {/each}
                    {/if}
                </div>
            </Card.Content>
        </Card.Root>
    </div>
</div>
