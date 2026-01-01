<script lang="ts" module>
    import { Armchair as ArmchairIcon } from "@lucide/svelte";
</script>

<script lang="ts">
    import { page } from "$app/stores";
    import { MOCK_TRIPS } from "$lib/mocks/data";
    import SeatMap from "$lib/components/blocks/SeatMap.svelte";
    import { Badge } from "$lib/components/ui/badge";
    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator"; // Need to make sure this exists or just use hr
    import type { Seat, SeatStatus } from "$lib/types/transport";
    import { CreditCard, ShieldCheck } from "@lucide/svelte";

    let tripId = $derived($page.params.tripId);
    let trip = $derived(MOCK_TRIPS.find((t) => t.id === tripId));

    let selectedSeats = $state<Seat[]>([]);

    // Mock Layout Generator (10 rows, 4 seats)
    function generateLayout(): Seat[][] {
        const rows = 10;
        const layout: Seat[][] = [];
        const chars = ["A", "B", "C", "D"];

        for (let i = 1; i <= rows; i++) {
            const row: Seat[] = [];
            for (let j = 0; j < 4; j++) {
                // Randomly book some seats
                const isBooked = Math.random() < 0.3;
                row.push({
                    id: `${i}${chars[j]}`,
                    label: `${chars[j]}${i}`,
                    status: isBooked ? "booked" : "available",
                    price: trip?.price || 500,
                });
            }
            layout.push(row);
        }
        return layout;
    }

    let layout = $state(generateLayout());
    let total = $derived(selectedSeats.reduce((acc, s) => acc + s.price, 0));
    let tax = $derived(total * 0.05); // 5% tax
    let grandTotal = $derived(total + tax);

    function handleCheckout() {
        alert("Proceeding to payment gateway...\nAmount: ৳" + grandTotal);
    }
</script>

<div class="min-h-screen bg-muted/30 pb-32 pt-20">
    <div class="container mx-auto max-w-6xl px-4">
        {#if trip}
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
                        <p class="text-xl text-muted-foreground mt-1">
                            Bus (AC Business) • Dhaka to Chittagong
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
                            class="w-full font-bold h-14 text-lg rounded-xl shadow-xl shadow-primary/20"
                            disabled={selectedSeats.length === 0}
                            onclick={handleCheckout}
                        >
                            Proceed to Pay
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
