<script lang="ts">
    import { inventoryApi } from "$lib/api/inventory";
    import { catalogApi } from "$lib/api/catalog";
    import { orderApi } from "$lib/api/order";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import {
        Search,
        User,
        Phone,
        Mail,
        Ticket,
        Printer,
        Loader2,
        Check,
    } from "@lucide/svelte";
    import { toast } from "svelte-sonner";

    // Search
    let searchQuery = $state("");
    let searchResults = $state<any[]>([]);
    let isSearching = $state(false);

    // Booking
    let selectedTrip = $state<any>(null);
    let passengers = $state<any[]>([]);
    let contactName = $state("");
    let contactPhone = $state("");
    let isBooking = $state(false);
    let bookingComplete = $state(false);
    let lastOrder = $state<any>(null);

    async function searchTrips() {
        if (!searchQuery.trim()) return;
        isSearching = true;
        try {
            // In production, call search API with route/date
            searchResults = [
                {
                    id: "TRIP-001",
                    operator: "Green Line",
                    from: "Dhaka",
                    to: "Chittagong",
                    departure: "2026-04-15T08:00:00",
                    available: 12,
                    price: 800,
                    type: "bus",
                },
                {
                    id: "TRIP-002",
                    operator: "Shohagh",
                    from: "Dhaka",
                    to: "Chittagong",
                    departure: "2026-04-15T10:00:00",
                    available: 8,
                    price: 750,
                    type: "bus",
                },
            ];
        } catch (error) {
            toast.error("Search failed");
        } finally {
            isSearching = false;
        }
    }

    function selectTrip(trip: any) {
        selectedTrip = trip;
        passengers = [{ name: "", nid: "", phone: "" }];
        searchResults = [];
    }

    function addPassenger() {
        if (passengers.length < selectedTrip?.available) {
            passengers = [...passengers, { name: "", nid: "", phone: "" }];
        }
    }

    function removePassenger(index: number) {
        if (passengers.length > 1) {
            passengers = passengers.filter((_, i) => i !== index);
        }
    }

    async function completeBooking() {
        // Validate
        for (let i = 0; i < passengers.length; i++) {
            if (!passengers[i].name?.trim()) {
                toast.error(`Passenger ${i + 1} name required`);
                return;
            }
        }
        if (!contactPhone.trim()) {
            toast.error("Contact phone required");
            return;
        }

        isBooking = true;
        try {
            // Create order for counter sale (cash payment)
            const order = await orderApi.createOrder({
                trip_id: selectedTrip.id,
                from_station_id: selectedTrip.from,
                to_station_id: selectedTrip.to,
                hold_id: "", // Counter booking doesn't need hold
                passengers: passengers.map((p, i) => ({
                    nid: p.nid || "N/A",
                    name: p.name,
                    seat_id: `C${i + 1}`, // Counter assigned seats
                    date_of_birth: "",
                    gender: "male",
                })),
                payment_method: { type: "cash" },
                contact_email: "",
                contact_phone: contactPhone,
                idempotency_key: crypto.randomUUID(),
            });

            lastOrder = order;
            bookingComplete = true;
            toast.success("Booking completed!", {
                description: `Order #${order.id.slice(0, 8)}`,
            });
        } catch (error: any) {
            toast.error("Booking failed", {
                description: error.message,
            });
        } finally {
            isBooking = false;
        }
    }

    function resetForm() {
        selectedTrip = null;
        passengers = [];
        contactName = "";
        contactPhone = "";
        bookingComplete = false;
        lastOrder = null;
    }

    function printTicket() {
        toast.info("Printing ticket...");
        // In production, open print dialog with TicketPrint component
    }
</script>

<div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">
                Counter Booking
            </h1>
            <p class="mt-2 text-muted-foreground">
                Create walk-in bookings for customers
            </p>
        </div>
        {#if bookingComplete}
            <Button onclick={resetForm}>New Booking</Button>
        {/if}
    </div>

    {#if !selectedTrip && !bookingComplete}
        <!-- Search Step -->
        <div class="glass-card rounded-xl p-6">
            <h3 class="mb-4 text-lg font-bold">Step 1: Find Trip</h3>
            <div class="flex gap-3">
                <div class="relative flex-1">
                    <Search
                        class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
                        size={18}
                    />
                    <Input
                        type="text"
                        bind:value={searchQuery}
                        placeholder="Search by route (e.g., Dhaka to Chittagong)"
                        class="pl-10"
                        onkeydown={(e) => {
                            if (e.key === "Enter") searchTrips();
                        }}
                    />
                </div>
                <Button onclick={searchTrips} disabled={isSearching}>
                    {#if isSearching}
                        <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                    {:else}
                        <Search class="mr-2 h-4 w-4" />
                    {/if}
                    Search
                </Button>
            </div>

            {#if searchResults.length > 0}
                <div class="mt-4 space-y-2">
                    {#each searchResults as trip}
                        <button
                            class="flex w-full items-center justify-between rounded-lg border border-border bg-white/50 p-4 text-left transition-all hover:bg-white hover:shadow-md"
                            onclick={() => selectTrip(trip)}
                        >
                            <div>
                                <p class="font-semibold">
                                    {trip.operator}
                                </p>
                                <p class="text-sm text-muted-foreground">
                                    {trip.from} → {trip.to} • {new Date(
                                        trip.departure,
                                    ).toLocaleTimeString([], {
                                        hour: "2-digit",
                                        minute: "2-digit",
                                    })}
                                </p>
                            </div>
                            <div class="text-right">
                                <p class="text-lg font-bold text-primary">
                                    ৳{trip.price}
                                </p>
                                <p class="text-xs text-muted-foreground">
                                    {trip.available} seats
                                </p>
                            </div>
                        </button>
                    {/each}
                </div>
            {/if}
        </div>
    {:else if selectedTrip && !bookingComplete}
        <!-- Passenger Details -->
        <div class="glass-card rounded-xl p-6">
            <div class="mb-4 flex items-center justify-between">
                <h3 class="text-lg font-bold">Step 2: Passenger Details</h3>
                <div class="text-right text-sm">
                    <p class="font-medium">{selectedTrip.operator}</p>
                    <p class="text-muted-foreground">
                        {selectedTrip.from} → {selectedTrip.to}
                    </p>
                </div>
            </div>

            <div class="space-y-4">
                {#each passengers as passenger, index}
                    <div class="rounded-lg border border-border p-4">
                        <div class="mb-3 flex items-center justify-between">
                            <h4 class="font-medium">Passenger {index + 1}</h4>
                            {#if passengers.length > 1}
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onclick={() => removePassenger(index)}
                                >
                                    Remove
                                </Button>
                            {/if}
                        </div>
                        <div class="grid gap-3 sm:grid-cols-3">
                            <Input
                                type="text"
                                bind:value={passenger.name}
                                placeholder="Full Name"
                            />
                            <Input
                                type="text"
                                bind:value={passenger.nid}
                                placeholder="NID (Optional)"
                            />
                            <Input
                                type="tel"
                                bind:value={passenger.phone}
                                placeholder="Phone"
                            />
                        </div>
                    </div>
                {/each}

                <Button
                    variant="outline"
                    onclick={addPassenger}
                    disabled={passengers.length >= selectedTrip.available}
                >
                    + Add Passenger
                </Button>

                <Separator />

                <!-- Contact Info -->
                <div class="grid gap-3 sm:grid-cols-2">
                    <Input
                        type="text"
                        bind:value={contactName}
                        placeholder="Contact Name"
                    />
                    <Input
                        type="tel"
                        bind:value={contactPhone}
                        placeholder="Contact Phone *"
                    />
                </div>

                <!-- Summary -->
                <div class="rounded-lg bg-muted/50 p-4">
                    <div class="flex justify-between">
                        <span>Total Amount</span>
                        <span class="text-xl font-bold text-primary">
                            ৳{(selectedTrip.price * passengers.length).toFixed(2)}
                        </span>
                    </div>
                    <p class="text-sm text-muted-foreground">
                        Payment: Cash at Counter
                    </p>
                </div>

                <Button
                    size="lg"
                    class="w-full"
                    onclick={completeBooking}
                    disabled={isBooking}
                >
                    {#if isBooking}
                        <Loader2 class="mr-2 h-5 w-5 animate-spin" />
                        Processing...
                    {:else}
                        <Check class="mr-2 h-5 w-5" />
                        Complete Booking
                    {/if}
                </Button>
            </div>
        </div>
    {:else if bookingComplete && lastOrder}
        <!-- Success -->
        <div class="glass-card rounded-xl p-8 text-center">
            <div
                class="mx-auto mb-6 flex size-20 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30"
            >
                <Check size={40} class="text-green-600 dark:text-green-400" />
            </div>
            <h2 class="mb-2 text-2xl font-bold">Booking Confirmed!</h2>
            <p class="mb-6 text-muted-foreground">
                Order #{lastOrder.id.slice(0, 8)}
            </p>

            <div class="mx-auto mb-6 max-w-sm rounded-lg bg-muted/50 p-4 text-left">
                <p><strong>Passengers:</strong> {passengers.length}</p>
                <p>
                    <strong>Total:</strong> ৳
                    {(lastOrder.total_paisa / 100).toFixed(2)}
                </p>
                <p><strong>Payment:</strong> Cash</p>
            </div>

            <div class="flex justify-center gap-3">
                <Button onclick={printTicket}>
                    <Printer class="mr-2 h-4 w-4" />
                    Print Ticket
                </Button>
                <Button variant="outline" onclick={resetForm}>
                    New Booking
                </Button>
            </div>
        </div>
    {/if}
</div>
