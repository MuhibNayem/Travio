<script lang="ts">
    import { onMount } from "svelte";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import * as Table from "$lib/components/ui/table";
    import { Badge } from "$lib/components/ui/badge";
    import { Combobox } from "$lib/components/ui/combobox";
    import {
        Search,
        Download,
        FileSpreadsheet,
        Users,
        Loader2,
        Filter,
    } from "@lucide/svelte";
    import {
        ordersApi,
        type Order,
        OrderStatus,
        PaymentStatus,
    } from "$lib/api/orders";
    import { catalogApi } from "$lib/api/catalog";
    import { toast } from "svelte-sonner";
    import { auth } from "$lib/runes/auth.svelte";

    // State
    let orders = $state<Order[]>([]);
    let trips = $state<Array<{ id: string; label: string }>>([]);
    let loading = $state(true);
    let selectedTripId = $state("");
    let searchQuery = $state("");

    // Derived - filtered passengers
    let filteredOrders = $derived(() => {
        if (!searchQuery) return orders;
        const q = searchQuery.toLowerCase();
        return orders.filter(
            (o) =>
                o.passengers.some((p) => p.name.toLowerCase().includes(q)) ||
                o.contact_phone.includes(q) ||
                o.booking_id?.toLowerCase().includes(q),
        );
    });

    // Stats
    let totalPassengers = $derived(
        orders.reduce((acc, o) => acc + o.passengers.length, 0),
    );
    let confirmedCount = $derived(
        orders.filter((o) => o.status === OrderStatus.ORDER_STATUS_CONFIRMED)
            .length,
    );

    onMount(async () => {
        await loadTrips();
    });

    async function loadTrips() {
        try {
            const res = await catalogApi.getTrips();
            trips = res.map((t) => ({
                id: t.id,
                label: `${t.vehicle_class || "Trip"} - ${new Date(t.departure_time * 1000).toLocaleDateString()}`,
            }));
        } catch (e) {
            console.error(e);
        }
    }

    async function loadOrders() {
        if (!selectedTripId) return;
        loading = true;
        try {
            // For now, we'll fetch by user and filter by trip
            // In production, add a backend endpoint for trip-specific orders
            const userId = auth.user?.id || "";
            const res = await ordersApi.listOrders(userId);
            orders = res.orders.filter((o) => o.trip_id === selectedTripId);
        } catch (e) {
            console.error(e);
            toast.error("Failed to load manifest");
        } finally {
            loading = false;
        }
    }

    $effect(() => {
        if (selectedTripId) {
            loadOrders();
        }
    });

    function exportCSV() {
        if (orders.length === 0) {
            toast.error("No data to export");
            return;
        }

        const headers = [
            "PNR",
            "Passenger",
            "Seat",
            "Phone",
            "Status",
            "Amount",
        ];
        const rows = orders.flatMap((o) =>
            o.passengers.map((p) => [
                o.booking_id || o.id.substring(0, 8),
                p.name,
                p.seat_id,
                o.contact_phone,
                OrderStatus[o.status]?.replace("ORDER_STATUS_", "") ||
                    "Unknown",
                (o.total_paisa / 100).toFixed(2),
            ]),
        );

        const csv = [headers.join(","), ...rows.map((r) => r.join(","))].join(
            "\n",
        );
        const blob = new Blob([csv], { type: "text/csv" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `manifest-${selectedTripId.substring(0, 8)}.csv`;
        a.click();
        URL.revokeObjectURL(url);
        toast.success("CSV exported successfully");
    }

    function getStatusColor(status: OrderStatus) {
        switch (status) {
            case OrderStatus.ORDER_STATUS_CONFIRMED:
                return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
            case OrderStatus.ORDER_STATUS_PENDING:
                return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
            case OrderStatus.ORDER_STATUS_CANCELLED:
            case OrderStatus.ORDER_STATUS_FAILED:
                return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
            default:
                return "bg-gray-100 text-gray-800";
        }
    }
</script>

<div class="space-y-6">
    <!-- Header -->
    <div
        class="flex flex-col md:flex-row md:items-center justify-between gap-4"
    >
        <div>
            <h1 class="text-3xl font-bold tracking-tight">
                Passenger Manifest
            </h1>
            <p class="text-muted-foreground mt-2">
                View and export passenger lists for trips.
            </p>
        </div>
        <Button onclick={exportCSV} disabled={orders.length === 0}>
            <Download class="mr-2 h-4 w-4" />
            Export CSV
        </Button>
    </div>

    <!-- Filters -->
    <div class="flex flex-col md:flex-row gap-4">
        <div class="w-full md:w-80">
            <Combobox
                items={trips.map((t) => ({ value: t.id, label: t.label }))}
                bind:value={selectedTripId}
                placeholder="Select a trip..."
            />
        </div>
        <div class="relative flex-1">
            <Search
                class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"
            />
            <Input
                class="pl-10"
                placeholder="Search by name, phone, or PNR..."
                bind:value={searchQuery}
            />
        </div>
    </div>

    <!-- Stats Cards -->
    {#if selectedTripId && !loading}
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="rounded-xl border bg-card p-4 shadow-sm">
                <div class="flex items-center gap-3">
                    <div class="rounded-lg bg-primary/10 p-2">
                        <FileSpreadsheet class="h-5 w-5 text-primary" />
                    </div>
                    <div>
                        <p class="text-sm text-muted-foreground">Bookings</p>
                        <p class="text-2xl font-bold">{orders.length}</p>
                    </div>
                </div>
            </div>
            <div class="rounded-xl border bg-card p-4 shadow-sm">
                <div class="flex items-center gap-3">
                    <div class="rounded-lg bg-blue-500/10 p-2">
                        <Users class="h-5 w-5 text-blue-500" />
                    </div>
                    <div>
                        <p class="text-sm text-muted-foreground">Passengers</p>
                        <p class="text-2xl font-bold">{totalPassengers}</p>
                    </div>
                </div>
            </div>
            <div class="rounded-xl border bg-card p-4 shadow-sm">
                <div class="flex items-center gap-3">
                    <div class="rounded-lg bg-green-500/10 p-2">
                        <Users class="h-5 w-5 text-green-500" />
                    </div>
                    <div>
                        <p class="text-sm text-muted-foreground">Confirmed</p>
                        <p class="text-2xl font-bold">{confirmedCount}</p>
                    </div>
                </div>
            </div>
            <div class="rounded-xl border bg-card p-4 shadow-sm">
                <div class="flex items-center gap-3">
                    <div class="rounded-lg bg-amber-500/10 p-2">
                        <Filter class="h-5 w-5 text-amber-500" />
                    </div>
                    <div>
                        <p class="text-sm text-muted-foreground">Pending</p>
                        <p class="text-2xl font-bold">
                            {orders.length - confirmedCount}
                        </p>
                    </div>
                </div>
            </div>
        </div>
    {/if}

    <!-- Table -->
    <div class="rounded-xl border bg-card shadow-sm">
        <Table.Root>
            <Table.Header>
                <Table.Row>
                    <Table.Head>PNR</Table.Head>
                    <Table.Head>PASSENGER</Table.Head>
                    <Table.Head>SEAT</Table.Head>
                    <Table.Head>PHONE</Table.Head>
                    <Table.Head>STATUS</Table.Head>
                    <Table.Head class="text-right">AMOUNT</Table.Head>
                </Table.Row>
            </Table.Header>
            <Table.Body>
                {#if !selectedTripId}
                    <Table.Row>
                        <Table.Cell colspan={6} class="h-64 text-center">
                            <div
                                class="flex flex-col items-center justify-center text-muted-foreground"
                            >
                                <FileSpreadsheet
                                    class="h-8 w-8 mb-4 opacity-50"
                                />
                                <p class="text-lg font-medium text-foreground">
                                    Select a Trip
                                </p>
                                <p class="text-sm">
                                    Choose a trip from the dropdown to view its
                                    manifest.
                                </p>
                            </div>
                        </Table.Cell>
                    </Table.Row>
                {:else if loading}
                    <Table.Row>
                        <Table.Cell colspan={6} class="h-24 text-center">
                            <Loader2
                                class="h-6 w-6 animate-spin mx-auto text-primary"
                            />
                        </Table.Cell>
                    </Table.Row>
                {:else if filteredOrders().length === 0}
                    <Table.Row>
                        <Table.Cell colspan={6} class="h-64 text-center">
                            <div
                                class="flex flex-col items-center justify-center text-muted-foreground"
                            >
                                <Users class="h-8 w-8 mb-4 opacity-50" />
                                <p class="text-lg font-medium text-foreground">
                                    No passengers found
                                </p>
                                <p class="text-sm">
                                    This trip has no bookings yet.
                                </p>
                            </div>
                        </Table.Cell>
                    </Table.Row>
                {:else}
                    {#each filteredOrders() as order}
                        {#each order.passengers as passenger, i}
                            <Table.Row>
                                <Table.Cell>
                                    {#if i === 0}
                                        <span
                                            class="font-mono text-xs bg-muted px-2 py-1 rounded"
                                        >
                                            {order.booking_id ||
                                                order.id.substring(0, 8)}
                                        </span>
                                    {/if}
                                </Table.Cell>
                                <Table.Cell class="font-medium"
                                    >{passenger.name}</Table.Cell
                                >
                                <Table.Cell>
                                    <Badge variant="outline"
                                        >{passenger.seat_id}</Badge
                                    >
                                </Table.Cell>
                                <Table.Cell class="text-muted-foreground">
                                    {i === 0 ? order.contact_phone : ""}
                                </Table.Cell>
                                <Table.Cell>
                                    {#if i === 0}
                                        <span
                                            class={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${getStatusColor(order.status)}`}
                                        >
                                            {OrderStatus[order.status]?.replace(
                                                "ORDER_STATUS_",
                                                "",
                                            ) || "Unknown"}
                                        </span>
                                    {/if}
                                </Table.Cell>
                                <Table.Cell class="text-right">
                                    {#if i === 0}
                                        ৳{(order.total_paisa / 100).toFixed(2)}
                                    {/if}
                                </Table.Cell>
                            </Table.Row>
                        {/each}
                    {/each}
                {/if}
            </Table.Body>
        </Table.Root>
    </div>
</div>
