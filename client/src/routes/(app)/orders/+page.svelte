<script lang="ts">
    import { orderApi, type Order } from "$lib/api/order";
    import { Button } from "$lib/components/ui/button";
    import {
        Plus,
        Calendar,
        MapPin,
        Clock,
        Loader2,
        Search,
        Ticket,
        Download,
        X,
        Eye,
    } from "@lucide/svelte";
    import { toast } from "svelte-sonner";
    import { goto } from "$app/navigation";

    let orders = $state<Order[]>([]);
    let loading = $state(true);
    let selectedStatus = $state<string>("");
    let showCancelModal = $state(false);
    let cancellingOrderId = $state("");
    let cancelReason = $state("");
    let isCancelling = $state(false);

    async function loadOrders() {
        loading = true;
        try {
            const response = await orderApi.listOrders();
            orders = response.orders || [];
        } catch (e) {
            console.error(e);
            toast.error("Failed to load bookings");
        } finally {
            loading = false;
        }
    }

    function getStatusColor(status: string) {
        switch (status) {
            case "confirmed":
                return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
            case "pending":
                return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
            case "cancelled":
                return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
            case "refunded":
                return "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400";
            default:
                return "bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400";
        }
    }

    function getPaymentStatusColor(status: string) {
        switch (status) {
            case "captured":
                return "text-green-600";
            case "pending":
                return "text-yellow-600";
            case "failed":
                return "text-red-600";
            case "refunded":
                return "text-blue-600";
            default:
                return "text-gray-600";
        }
    }

    async function cancelOrder(orderId: string) {
        if (!cancelReason.trim()) {
            toast.error("Please provide a reason");
            return;
        }

        isCancelling = true;
        try {
            await orderApi.cancelOrder(orderId, cancelReason);
            toast.success("Booking cancelled", {
                description: "Refund will be processed within 5-7 business days",
            });
            showCancelModal = false;
            cancelReason = "";
            loadOrders();
        } catch (e) {
            toast.error("Cancellation failed");
        } finally {
            isCancelling = false;
        }
    }

    function filterByStatus(status: string) {
        selectedStatus = status;
    }

    $effect(() => {
        loadOrders();
    });

    $: filteredOrders = selectedStatus
        ? orders.filter((o) => o.status === selectedStatus)
        : orders;
</script>

<div class="min-h-screen bg-muted/30 py-20">
    <div class="container mx-auto max-w-6xl px-4">
        <!-- Header -->
        <div class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
                <h1 class="text-3xl font-bold">My Bookings</h1>
                <p class="mt-1 text-muted-foreground">
                    Manage your travel bookings
                </p>
            </div>
            <Button href="/search">
                <Plus class="mr-2 h-4 w-4" />
                Book New Trip
            </Button>
        </div>

        <!-- Filters -->
        <div class="mb-6 flex flex-wrap gap-2">
            <button
                class="rounded-full px-4 py-2 text-sm font-medium transition-all {!selectedStatus ? 'bg-primary text-primary-foreground' : 'bg-white text-muted-foreground hover:bg-muted'}"
                onclick={() => filterByStatus("")}
            >
                All
            </button>
            {#each ["pending", "confirmed", "cancelled", "refunded"] as status}
                <button
                    class="rounded-full px-4 py-2 text-sm font-medium capitalize transition-all {selectedStatus === status ? 'bg-primary text-primary-foreground' : 'bg-white text-muted-foreground hover:bg-muted'}"
                    onclick={() => filterByStatus(status)}
                >
                    {status}
                </button>
            {/each}
        </div>

        <!-- Orders List -->
        {#if loading}
            <div
                class="flex h-64 flex-col items-center justify-center gap-4"
            >
                <Loader2 class="animate-spin text-primary" size={32} />
                <p class="text-muted-foreground">Loading bookings...</p>
            </div>
        {:else if filteredOrders.length > 0}
            <div class="space-y-4">
                {#each filteredOrders as order}
                    <div
                        class="glass-card rounded-xl p-6 transition-all hover:shadow-lg"
                    >
                        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                            <!-- Order Info -->
                            <div class="flex-1">
                                <div class="mb-2 flex items-center gap-3">
                                    <h3 class="text-lg font-bold">
                                        Order #{order.id.slice(0, 8)}
                                    </h3>
                                    <span
                                        class="rounded-full px-2.5 py-0.5 text-xs font-semibold {getStatusColor(order.status)}"
                                    >
                                        {order.status}
                                    </span>
                                </div>

                                <div class="grid gap-2 sm:grid-cols-2 md:grid-cols-3">
                                    <div class="flex items-center gap-2 text-sm">
                                        <MapPin
                                            size={16}
                                            class="text-muted-foreground"
                                        />
                                        <span
                                            >{order.from_station_id} → {order.to_station_id}</span
                                        >
                                    </div>
                                    <div class="flex items-center gap-2 text-sm">
                                        <Calendar
                                            size={16}
                                            class="text-muted-foreground"
                                        />
                                        <span
                                            >{new Date(
                                                order.created_at,
                                            ).toLocaleDateString()}</span
                                        >
                                    </div>
                                    <div class="flex items-center gap-2 text-sm">
                                        <Ticket
                                            size={16}
                                            class="text-muted-foreground"
                                        />
                                        <span
                                            >{(order.passengers || []).length} Passenger(s)</span
                                        >
                                    </div>
                                </div>

                                <div class="mt-3 flex items-center gap-4">
                                    <span class="text-lg font-bold text-primary">
                                        ৳{(order.total_paisa / 100).toFixed(2)}
                                    </span>
                                    <span
                                        class="text-sm {getPaymentStatusColor(order.payment_status)}"
                                    >
                                        Payment: {order.payment_status}
                                    </span>
                                </div>
                            </div>

                            <!-- Actions -->
                            <div class="flex gap-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onclick={() =>
                                        goto(`/confirmation/${order.id}`)}
                                >
                                    <Eye class="mr-1 h-4 w-4" />
                                    View
                                </Button>
                                {#if order.status === "confirmed"}
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onclick={() => {
                                            cancellingOrderId = order.id;
                                            showCancelModal = true;
                                        }}
                                    >
                                        <X class="mr-1 h-4 w-4" />
                                        Cancel
                                    </Button>
                                {/if}
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        {:else}
            <div
                class="flex flex-col items-center justify-center rounded-2xl border bg-white p-12 text-center shadow-sm"
            >
                <div class="mb-4 rounded-full bg-muted p-4">
                    <Search
                        size={32}
                        class="text-muted-foreground"
                    />
                </div>
                <h3 class="text-xl font-bold">No bookings found</h3>
                <p class="mt-2 max-w-md text-muted-foreground">
                    {selectedStatus
                        ? `No bookings with status "${selectedStatus}"`
                        : "You haven't made any bookings yet. Book your first trip!"}
                </p>
                <Button class="mt-4" href="/search">
                    Search Trips
                </Button>
            </div>
        {/if}
    </div>

    <!-- Cancel Modal -->
    {#if showCancelModal}
        <div
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
            onclick={() => (showCancelModal = false)}
        >
            <div
                class="mx-4 w-full max-w-md rounded-xl bg-white p-6 shadow-2xl"
                onclick={(e) => e.stopPropagation()}
            >
                <h3 class="mb-4 text-lg font-bold">Cancel Booking</h3>
                <p class="mb-4 text-sm text-muted-foreground">
                    Are you sure you want to cancel this booking? A refund will
                    be processed within 5-7 business days.
                </p>
                <div class="mb-4">
                    <label class="mb-2 block text-sm font-medium"
                        >Reason for cancellation</label
                    >
                    <textarea
                        bind:value={cancelReason}
                        class="w-full rounded-lg border border-border p-3 text-sm focus:ring-2 focus:ring-primary focus:ring-offset-2"
                        rows="3"
                        placeholder="Please provide a reason..."
                    ></textarea>
                </div>
                <div class="flex gap-3">
                    <Button
                        variant="outline"
                        class="flex-1"
                        onclick={() => (showCancelModal = false)}
                    >
                        Keep Booking
                    </Button>
                    <Button
                        variant="destructive"
                        class="flex-1"
                        onclick={() => cancelOrder(cancellingOrderId)}
                        disabled={isCancelling || !cancelReason}
                    >
                        {#if isCancelling}
                            <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                        {/if}
                        Cancel Booking
                    </Button>
                </div>
            </div>
        </div>
    {/if}
</div>
