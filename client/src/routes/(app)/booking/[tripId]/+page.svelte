<script lang="ts" module>
    import { Armchair as ArmchairIcon } from "@lucide/svelte";
</script>

<script lang="ts">
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import SeatMap from "$lib/components/blocks/SeatMap.svelte";
    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";
    import type { Seat, SeatStatus, Trip } from "$lib/types/transport";
    import { CreditCard, ShieldCheck, Loader } from "@lucide/svelte";
    import { searchApi } from "$lib/api/search";
    import { inventoryApi } from "$lib/api/inventory";
    import { catalogApi } from "$lib/api/catalog";
    import { toast } from "svelte-sonner";

    let tripId = $derived($page.params.tripId);
    let fromId = $derived($page.url.searchParams.get("from") || "");
    let toId = $derived($page.url.searchParams.get("to") || "");

    let trip = $state<Trip | null>(null);
    let fromStationName = $state<string>("Unknown Origin");
    let toStationName = $state<string>("Unknown Destination");
    let layout = $state<Seat[][]>([]);
    let isLoading = $state(true);
    let isHolding = $state(false);

    let selectedSeats = $state<Seat[]>([]);

    async function getStationName(id: string): Promise<string> {
        if (!id) return "";
        try {
            const s = await catalogApi.getStation(id);
            return s.name; // assuming optional chaining if needed, but s defined
        } catch {
            return id;
        }
    }

    async function fetchData() {
        if (!tripId) return;
        isLoading = true;
        try {
            // Parallel fetch for details
            const [tripData, originName, destName] = await Promise.all([
                searchApi.getTrip(tripId),
                getStationName(fromId),
                getStationName(toId),
            ]);

            fromStationName = originName || "Unknown";
            toStationName = destName || "Unknown";

            // Map tripData (snake_case) to Trip (camelCase)
            const t = tripData as any;
            trip = {
                id: t.id,
                routeId: t.route_id,
                type: (t.vehicle_type || "bus") as any,
                operator: t.operator_name || "Travio Partner",
                vehicleName: t.vehicle_class || "Standard",
                departureTime: t.departure_time,
                arrivalTime: t.arrival_time || t.departure_time,
                price: t.pricing?.base_price_paisa
                    ? t.pricing.base_price_paisa / 100
                    : 0,
                class: t.vehicle_class,
                availableSeats: t.total_seats,
                totalSeats: t.total_seats,
            };

            // Fetch SeatMap
            if (fromId && toId) {
                const mapResp = await inventoryApi.getSeatMap(
                    tripId,
                    fromId,
                    toId,
                );
                layout = mapResp.rows.map((r) =>
                    r.seats.map((s) => ({
                        id: s.seat_id,
                        label: s.seat_number,
                        status: s.status.toLowerCase() as SeatStatus,
                        price: s.price_paisa / 100,
                    })),
                );
            }
        } catch (error) {
            console.error("Failed to load booking data", error);
            toast.error("Failed to load trip details");
        } finally {
            isLoading = false;
        }
    }

    $effect(() => {
        if (tripId) fetchData();
    });

    let total = $derived(selectedSeats.reduce((acc, s) => acc + s.price, 0));
    let tax = $derived(total * 0.05); // 5% tax
    let grandTotal = $derived(total + tax);

    async function handleCheckout() {
        if (selectedSeats.length === 0 || !tripId) return;
        isHolding = true;
        try {
            const sessionId = crypto.randomUUID();
            const holdResp = await inventoryApi.holdSeats({
                trip_id: tripId,
                from_station_id: fromId,
                to_station_id: toId,
                seat_ids: selectedSeats.map((s) => s.id),
                session_id: sessionId,
            });

            if (holdResp.success) {
                toast.success("Seats held successfully!");
                goto(`/checkout/${holdResp.hold_id}`);
            } else {
                toast.error("Failed to hold seats: " + holdResp.failure_reason);
                fetchData(); // Refresh map
            }
        } catch (error) {
            console.error("Checkout failed", error);
            toast.error("System error during checkout");
        } finally {
            isHolding = false;
        }
    }
</script>

<div class="min-h-screen bg-muted/30 pb-32 pt-20">
    <div class="container mx-auto max-w-6xl px-4">
        {#if isLoading}
            <div
                class="flex h-[50vh] flex-col items-center justify-center gap-4"
            >
                <Loader class="animate-spin text-primary" size={48} />
                <p class="text-muted-foreground">Loading trip details...</p>
            </div>
        {:else if trip}
            <div class="mb-8">
                <Button
                    variant="ghost"
                    href="/search"
                    class="-ml-4 text-muted-foreground hover:text-foreground"
                >
                    ← Back to Search
                </Button>
                <div class="flex items-center justify-between mt-4">
                    <div>
                        <h1 class="text-3xl font-bold">
                            {trip.operator} - {trip.vehicleName}
                        </h1>
                        <p
                            class="text-xl text-muted-foreground mt-1 capitalize"
                        >
                            {trip.type} ({trip.class}) • {fromStationName} to {toStationName}
                        </p>
                    </div>
                    <div class="text-right">
                        <p class="text-lg font-bold text-primary">
                            {new Date(trip.departureTime).toLocaleString()}
                        </p>
                    </div>
                </div>
            </div>

            <div class="flex flex-col gap-8 lg:flex-row">
                <!-- Left: Seat Map -->
                <div class="flex-1">
                    <div
                        class="flex items-center justify-center p-8 glass-panel min-h-[600px]"
                    >
                        <SeatMap
                            {layout}
                            onSelectionChange={(seats) =>
                                (selectedSeats = seats)}
                        />
                    </div>
                </div>

                <!-- Right: Summary -->
                <div class="w-full lg:w-[400px]">
                    <div class="glass-panel sticky top-24 p-6">
                        <h2 class="text-xl font-bold mb-6">Booking Summary</h2>

                        {#if selectedSeats.length === 0}
                            <div
                                class="flex flex-col items-center justify-center py-10 text-center text-muted-foreground"
                            >
                                <ArmchairIcon class="size-12 mb-3 opacity-20" />
                                <p>Select seats to proceed</p>
                            </div>
                        {:else}
                            <div class="space-y-4 mb-6">
                                {#each selectedSeats as seat}
                                    <div
                                        class="flex justify-between items-center text-sm"
                                    >
                                        <span class="font-bold"
                                            >Seat {seat.label}</span
                                        >
                                        <span>৳{seat.price}</span>
                                    </div>
                                {/each}
                                <Separator />
                                <div
                                    class="flex justify-between items-center text-sm text-muted-foreground"
                                >
                                    <span>Subtotal</span>
                                    <span>৳{total}</span>
                                </div>
                                <div
                                    class="flex justify-between items-center text-sm text-muted-foreground"
                                >
                                    <span>Service Charge & Tax (5%)</span>
                                    <span>৳{tax}</span>
                                </div>
                                <Separator />
                                <div
                                    class="flex justify-between items-center text-lg font-black"
                                >
                                    <span>Total</span>
                                    <span>৳{grandTotal}</span>
                                </div>
                            </div>
                        {/if}

                        <Button
                            size="lg"
                            class="w-full font-bold h-14 text-lg rounded-xl shadow-xl shadow-primary/20 gap-2"
                            disabled={selectedSeats.length === 0 || isHolding}
                            onclick={handleCheckout}
                        >
                            {#if isHolding}
                                <Loader class="animate-spin" size={20} />
                                Holding Seats...
                            {:else}
                                <CreditCard size={20} />
                                Proceed to Pay
                            {/if}
                        </Button>

                        <div
                            class="mt-6 flex items-center justify-center gap-2 text-xs text-muted-foreground"
                        >
                            <ShieldCheck size={14} class="text-green-600" />
                            <span>Payments are secure and encrypted</span>
                        </div>
                    </div>
                </div>
            </div>
        {:else}
            <div class="flex h-[50vh] items-center justify-center">
                <p>Trip not found</p>
            </div>
        {/if}
    </div>
</div>
