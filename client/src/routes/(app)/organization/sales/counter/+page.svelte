<script lang="ts">
    import { onMount } from "svelte";
    import { Input } from "$lib/components/ui/input";
    import { Button } from "$lib/components/ui/button";
    import * as Tabs from "$lib/components/ui/tabs";
    import {
        Search,
        MonitorPlay,
        Bus,
        Calendar,
        MapPin,
        Armchair,
        ShoppingCart,
        User,
        CreditCard,
        Loader2,
    } from "@lucide/svelte";
    import SeatMap from "$lib/components/sales/seatmap/SeatMap.svelte";
    import { Badge } from "$lib/components/ui/badge";
    import { Separator } from "$lib/components/ui/separator";
    import { eventsApi } from "$lib/api/events";
    import { catalogApi } from "$lib/api/catalog";
    import { fleetApi, type AssetConfig, AssetType } from "$lib/api/fleet";
    import * as Dialog from "$lib/components/ui/dialog";
    import TicketPrint from "$lib/components/sales/TicketPrint.svelte";
    import { ordersApi, type CreateOrderRequest } from "$lib/api/orders";
    import { paymentApi, type PaymentMethod } from "$lib/api/payment";
    import { auth } from "$lib/runes/auth.svelte";
    import { toast } from "svelte-sonner";

    // State
    let items: any[] = [];
    let paymentMethods: PaymentMethod[] = [];
    let loading = true;
    let processing = false;
    let activeTab: "trips" | "events" = "trips";
    let showTicket = false;
    let createdOrder: any = null;

    // Selected State
    let selectedItem: any = null;
    let selectedSeats: string[] = [];
    let customerName = "";
    let customerPhone = "";
    let selectedPaymentId = "cash";

    // Asset config for SeatMap
    let assetConfig: AssetConfig | null = null;
    let loadingAsset = false;

    $: subtotal = selectedSeats.length * (selectedItem?.price || 0);

    onMount(async () => {
        loadData();
    });

    async function loadData() {
        loading = true;
        try {
            const orgId = auth.user?.organizationId;
            if (!orgId) return;

            if (activeTab === "trips") {
                const res = await catalogApi.listTripInstances();
                items = res.map((r) => {
                    const t = r.trip;
                    return {
                        id: t.id,
                        title: r.route?.name || `${t.vehicle_class} Trip`,
                        time: new Date(t.departure_time * 1000).toLocaleTimeString(
                            [],
                            { hour: "2-digit", minute: "2-digit" },
                        ),
                        type: "bus",
                        price: t.pricing?.base_price_paisa
                            ? t.pricing.base_price_paisa / 100
                            : 0,
                        operator: r.operator_name || t.organization_id.substring(0, 8),
                        seatsAvailable: t.available_seats,
                        raw: t,
                    };
                });
            } else {
                const res = await eventsApi.searchEvents("");
                items = res.results.map((r) => ({
                    id: r.event.id,
                    title: r.event.title,
                    time: new Date(r.event.start_time).toLocaleDateString([], {
                        weekday: "short",
                        day: "numeric",
                        month: "short",
                    }),
                    type: "event",
                    price: 0,
                    venue: r.venue.name,
                    ticketTypes: [],
                    raw: r,
                }));
            }

            // Load Payment Methods
            if (paymentMethods.length === 0) {
                try {
                    const pRes = await paymentApi.getPaymentMethods();
                    paymentMethods = pRes.methods || [];
                } catch (e) {
                    console.error("Failed to load payments", e);
                }
            }
        } catch (err) {
            console.error("Failed to load data", err);
            toast.error("Failed to load data");
        } finally {
            loading = false;
        }
    }

    function handleTabChange(tab: string) {
        activeTab = tab as any;
        selectedItem = null;
        loadData();
    }

    // Handlers
    async function selectItem(item: any) {
        selectedItem = item;
        selectedSeats = []; // Reset selection
        assetConfig = null;

        // Fetch asset config for trips
        if (activeTab === "trips" && item.raw?.vehicle_id) {
            loadingAsset = true;
            try {
                const asset = await fleetApi.getAsset(item.raw.vehicle_id);
                assetConfig = asset.config || null;
            } catch (e) {
                console.error("Failed to load asset config", e);
                // Fall back to defaults
            } finally {
                loadingAsset = false;
            }
        }
    }

    // Helper: Determine SeatMap type from item
    function getSeatMapType(item: any): "bus" | "train" | "launch" | "event" {
        if (activeTab === "events") return "event";

        // Map vehicle_type to SeatMap type
        const vehicleType = item.raw?.vehicle_type || "bus";
        if (vehicleType.includes("train") || vehicleType.includes("TRAIN")) return "train";
        if (vehicleType.includes("launch") || vehicleType.includes("LAUNCH") || vehicleType.includes("ferry")) return "launch";
        return "bus";
    }

    // Helper: Build SeatMap config from asset config
    function buildSeatMapConfig(item: any, config: AssetConfig | null): any {
        const type = getSeatMapType(item);

        // Use asset config if available
        if (config) {
            if (type === "bus" && config.bus) {
                return {
                    rows: config.bus.rows || 10,
                    columns: config.bus.seats_per_row || 4,
                    aisleIndex: config.bus.aisle_after_seat || 2,
                };
            }
            if (type === "train" && config.train) {
                return { coaches: config.train.coaches || [] };
            }
            if (type === "launch" && config.launch) {
                return { decks: config.launch.decks || [] };
            }
        }

        // Default fallbacks
        if (type === "bus") {
            return { rows: 10, columns: 4, aisleIndex: 2 };
        }
        if (type === "train") {
            return {
                coaches: [
                    {
                        id: "S1",
                        name: "S1",
                        class: "S_Chair",
                        rows: 15,
                        seatsPerRow: 6,
                        hasBerths: false,
                    },
                ],
            };
        }
        if (type === "launch") {
            return {
                decks: [
                    {
                        id: "D1",
                        name: "Deck 1",
                        type: "economy",
                        rows: 8,
                        cols: 6,
                        seatPrice: 500,
                    },
                ],
            };
        }
        // Event fallback
        return { zones: [] };
    }

    function handleSeatSelection(e: CustomEvent) {
        selectedSeats = e.detail;
    }

    async function handleBooking() {
        if (!selectedItem || selectedSeats.length === 0) return;
        if (!customerName || !customerPhone) {
            toast.error("Please enter customer name and phone");
            return;
        }

        processing = true;
        try {
            const orgId = auth.user?.organizationId;
            if (!orgId) throw new Error("Organization ID missing");

            const payload: CreateOrderRequest = {
                organization_id: orgId,
                user_id: "counter_agent",
                trip_id: activeTab === "trips" ? selectedItem.id : "",
                from_station_id: "counter",
                to_station_id: "counter",
                passengers: selectedSeats.map((seat) => ({
                    name: customerName,
                    seat_id: seat,
                })),
                payment_method: {
                    type: selectedPaymentId,
                },
                contact_email: "",
                contact_phone: customerPhone,
            };

            const res = await ordersApi.createOrder(payload);

            if (res.payment_redirect_url && selectedPaymentId !== "cash") {
                toast.success("Order Created. Proceed to payment.");
                createdOrder = res.order;
                showTicket = true;
            } else {
                toast.success("Booking confirmed! Printing ticket...");
                createdOrder = res.order;
                showTicket = true;
            }

            // Reset
            selectedSeats = [];
            customerName = "";
            customerPhone = "";
            selectedPaymentId = "cash";
        } catch (err) {
            console.error(err);
            toast.error("Booking failed");
        } finally {
            processing = false;
        }
    }
</script>

<div class="flex h-[calc(100vh-4rem)] overflow-hidden bg-gray-50/50">
    <!-- LEFT PANEL: Search & Results -->
    <div class="w-96 flex flex-col border-r bg-white">
        <div class="p-4 border-b">
            <h2 class="font-bold text-lg mb-4 flex items-center gap-2">
                <MonitorPlay class="w-5 h-5 text-primary" /> Counter Sales
            </h2>
            <Tabs.Root
                value={activeTab}
                onValueChange={handleTabChange}
                class="w-full"
            >
                <Tabs.List class="grid w-full grid-cols-2">
                    <Tabs.Trigger value="trips">
                        <Bus class="w-4 h-4 mr-2" /> Trips
                    </Tabs.Trigger>
                    <Tabs.Trigger value="events">
                        <Calendar class="w-4 h-4 mr-2" /> Events
                    </Tabs.Trigger>
                </Tabs.List>
            </Tabs.Root>
            <div class="mt-4 relative">
                <Search
                    class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground"
                />
                <Input
                    placeholder="Search destination or event..."
                    class="pl-9"
                />
            </div>
        </div>

        <div class="flex-1 overflow-y-auto p-2 space-y-2">
            {#if loading}
                <div class="flex justify-center p-8 text-primary animate-spin">
                    <Loader2 />
                </div>
            {:else}
                {#each items as item}
                    <button
                        class="w-full text-left p-3 rounded-lg border hover:border-primary transition-all group flex flex-col gap-2 relative bg-white"
                        class:border-primary={selectedItem?.id === item.id}
                        class:ring-1={selectedItem?.id === item.id}
                        class:ring-primary={selectedItem?.id === item.id}
                        onclick={() => selectItem(item)}
                    >
                        <div class="flex justify-between items-start">
                            <span
                                class="font-semibold text-sm group-hover:text-primary transition-colors"
                                >{item.title}</span
                            >
                            <Badge variant="outline" class="text-[10px]"
                                >{item.type}</Badge
                            >
                        </div>

                        <div
                            class="flex justify-between items-center text-xs text-muted-foreground"
                        >
                            <div class="flex items-center gap-1">
                                {#if item.type === "event"}
                                    <MapPin class="w-3 h-3" /> {item.venue}
                                {:else}
                                    <Bus class="w-3 h-3" /> {item.operator}
                                {/if}
                            </div>
                            <div
                                class="font-mono bg-gray-100 px-1.5 py-0.5 rounded text-gray-700"
                            >
                                {item.time}
                            </div>
                        </div>
                    </button>
                {/each}
            {/if}
        </div>
    </div>

    <!-- CENTER PANEL: Seat Map -->
    <div class="flex-1 flex flex-col bg-gray-100/50 overflow-hidden relative">
        {#if selectedItem}
            <div
                class="p-4 bg-white border-b flex justify-between items-center shadow-sm z-10"
            >
                <div>
                    <h3 class="font-bold text-lg">{selectedItem.title}</h3>
                    <p
                        class="text-sm text-muted-foreground flex items-center gap-2"
                    >
                        <span
                            class="font-mono text-xs bg-primary/10 text-primary px-2 py-0.5 rounded"
                            >Tickets: ৳{selectedItem.price}</span
                        >
                        {#if selectedItem.venue}• {selectedItem.venue}{/if}
                    </p>
                </div>
                <div class="flex items-center gap-4 text-sm">
                    <div class="flex items-center gap-2">
                        <div
                            class="w-3 h-3 rounded-full bg-white border border-gray-300"
                        ></div>
                        Available
                    </div>
                    <div class="flex items-center gap-2">
                        <div class="w-3 h-3 rounded-full bg-primary"></div>
                        Selected
                    </div>
                    <div class="flex items-center gap-2">
                        <div class="w-3 h-3 rounded-full bg-gray-300"></div>
                        Sold
                    </div>
                </div>
            </div>

            <div
                class="flex-1 overflow-auto p-8 flex items-start justify-center"
            >
                {#if loadingAsset}
                    <div class="flex items-center justify-center h-64">
                        <Loader2 class="w-8 h-8 animate-spin text-primary" />
                    </div>
                {:else}
                    <SeatMap
                        type={getSeatMapType(selectedItem)}
                        config={buildSeatMapConfig(selectedItem, assetConfig)}
                        on:selectionChange={handleSeatSelection}
                    />
                {/if}
            </div>
        {:else}
            <div
                class="flex-1 flex flex-col items-center justify-center text-muted-foreground"
            >
                <Armchair class="w-16 h-16 mb-4 opacity-20" />
                <p>Select a trip or event to view availability</p>
            </div>
        {/if}
    </div>

    <!-- RIGHT PANEL: Cart & Checkout -->
    <div class="w-80 border-l bg-white flex flex-col">
        <div class="p-4 border-b bg-gray-50/30">
            <h2 class="font-bold mb-1 flex items-center gap-2">
                <ShoppingCart class="w-4 h-4" /> Booking Cart
            </h2>
            <p class="text-xs text-muted-foreground">Currently processing...</p>
        </div>

        <div class="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
            <!-- Selected Seats List -->
            {#if selectedSeats.length > 0}
                <div
                    class="bg-primary/5 border border-primary/20 rounded-lg p-3"
                >
                    <h4
                        class="text-xs font-semibold text-primary mb-2 uppercase tracking-wider"
                    >
                        Seats Selected ({selectedSeats.length})
                    </h4>
                    <div class="flex flex-wrap gap-2">
                        {#each selectedSeats as seat}
                            <Badge
                                class="bg-white hover:bg-white text-primary border-primary shadow-sm"
                                >{seat}</Badge
                            >
                        {/each}
                    </div>
                </div>
            {:else}
                <div
                    class="border rounded-lg border-dashed p-6 flex flex-col items-center justify-center text-center text-muted-foreground text-sm"
                >
                    <Armchair class="w-8 h-8 mb-2 opacity-30" />
                    No seats selected
                </div>
            {/if}

            <Separator />

            <!-- Customer Details -->
            <div class="space-y-3">
                <h4
                    class="text-xs font-semibold uppercase tracking-wider text-muted-foreground"
                >
                    Customer Info
                </h4>
                <div class="space-y-2">
                    <div class="relative">
                        <User
                            class="absolute left-2.5 top-2.5 w-4 h-4 text-gray-400"
                        />
                        <Input
                            class="pl-9 h-9"
                            placeholder="Customer Name"
                            bind:value={customerName}
                        />
                    </div>
                    <div class="relative">
                        <div
                            class="absolute left-2.5 top-2.5 w-4 h-4 text-gray-400 font-bold text-xs flex items-center justify-center"
                        >
                            +88
                        </div>
                        <Input
                            class="pl-9 h-9"
                            placeholder="017..."
                            bind:value={customerPhone}
                        />
                    </div>
                </div>
            </div>
        </div>

        <div class="px-4 pb-4">
            <Separator class="my-4" />
            <div class="space-y-3">
                <h4
                    class="text-xs font-semibold uppercase tracking-wider text-muted-foreground"
                >
                    Payment Method
                </h4>
                <div class="grid grid-cols-2 gap-2">
                    <button
                        class="border rounded-md p-2 text-xs font-medium transition-colors {selectedPaymentId ===
                        'cash'
                            ? 'bg-primary text-white border-primary'
                            : 'hover:border-primary'}"
                        onclick={() => (selectedPaymentId = "cash")}
                    >
                        Cash
                    </button>
                    {#each paymentMethods as pm}
                        <button
                            class="border rounded-md p-2 text-xs font-medium transition-colors {selectedPaymentId ===
                            pm.id
                                ? 'bg-primary text-white border-primary'
                                : 'hover:border-primary'}"
                            onclick={() => (selectedPaymentId = pm.id)}
                        >
                            {pm.name}
                        </button>
                    {/each}
                </div>
            </div>
        </div>

        <!-- Footer / Checkout -->
        <div class="p-4 border-t bg-gray-50">
            <div class="space-y-1 mb-4">
                <div class="flex justify-between text-sm">
                    <span class="text-muted-foreground">Subtotal</span>
                    <span>৳{subtotal}</span>
                </div>
                <div class="flex justify-between text-sm">
                    <span class="text-muted-foreground">Service Chg.</span>
                    <span>৳0</span>
                </div>
                <div
                    class="flex justify-between font-bold text-lg pt-2 border-t mt-2"
                >
                    <span>Total</span>
                    <span class="text-primary">৳{subtotal}</span>
                </div>
            </div>

            <Button
                class="w-full gap-2 font-bold shadow-lg shadow-primary/20"
                size="lg"
                disabled={selectedSeats.length === 0 || processing}
                onclick={handleBooking}
            >
                {#if processing}
                    <Loader2 class="w-4 h-4 animate-spin" /> Processing...
                {:else}
                    <CreditCard class="w-4 h-4" /> Book & Print
                {/if}
            </Button>
        </div>
    </div>
</div>

<Dialog.Root bind:open={showTicket}>
    <Dialog.Content class="max-w-[400px]">
        <Dialog.Header>
            <Dialog.Title>Booking Confirmed</Dialog.Title>
            <Dialog.Description>
                Ticket generated successfully. Please print it for the customer.
            </Dialog.Description>
        </Dialog.Header>
        {#if createdOrder}
            <div
                class="flex justify-center py-4 bg-gray-50 rounded-lg max-h-[60vh] overflow-y-auto print:hidden"
            >
                <TicketPrint order={createdOrder} />
            </div>
            <!-- Hidden Print Container -->
            <div class="hidden print:block fixed inset-0 bg-white z-[9999]">
                <TicketPrint order={createdOrder} />
            </div>

            <Dialog.Footer class="sm:justify-end gap-2 print:hidden">
                <Button variant="outline" onclick={() => (showTicket = false)}>
                    Close
                </Button>
                <Button onclick={() => window.print()}>Print Ticket</Button>
            </Dialog.Footer>
        {/if}
    </Dialog.Content>
</Dialog.Root>
