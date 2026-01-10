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

    // State
    let items: any[] = [];
    let loading = true;
    let activeTab: "trips" | "events" = "trips";

    // Selected State
    let selectedItem: any = null;
    let selectedSeats: string[] = [];
    let customerName = "";
    let customerPhone = "";

    $: subtotal = selectedSeats.length * (selectedItem?.price || 0);

    onMount(async () => {
        loadData();
    });

    async function loadData() {
        loading = true;
        try {
            if (activeTab === "trips") {
                const res = await catalogApi.getTrips();
                items = res.map((t) => ({
                    id: t.id,
                    title: `${t.vehicle_class} Trip`, // Enhance with full name in future
                    time: new Date(t.departure_time * 1000).toLocaleTimeString(
                        [],
                        { hour: "2-digit", minute: "2-digit" },
                    ),
                    type: "bus",
                    price: t.pricing?.base_price_paisa
                        ? t.pricing.base_price_paisa / 100
                        : 0,
                    operator: t.organization_id.substring(0, 8),
                    seatsAvailable: t.available_seats,
                    raw: t,
                }));
            } else {
                // Hardcode org ID for demo or use current user's org
                const orgId = "org_2rQd5zK8X7y3vM9pL4nJ1";
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
        } catch (err) {
            console.error("Failed to load data", err);
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
    function selectItem(item: any) {
        selectedItem = item;
        selectedSeats = []; // Reset selection
    }

    function handleSeatSelection(e: CustomEvent) {
        selectedSeats = e.detail;
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
                        on:click={() => selectItem(item)}
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
                <SeatMap
                    type={selectedItem.type}
                    config={selectedItem.type === "bus"
                        ? { rows: 10, columns: 4, aisleIndex: 1 }
                        : {
                              zones: [
                                  {
                                      id: "vip",
                                      name: "VIP",
                                      price: 5000,
                                      rows: 4,
                                      cols: 8,
                                  },
                                  {
                                      id: "reg",
                                      name: "Regular",
                                      price: 2500,
                                      rows: 8,
                                      cols: 12,
                                  },
                              ],
                          }}
                    on:selectionChange={handleSeatSelection}
                />
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
                disabled={selectedSeats.length === 0}
            >
                <CreditCard class="w-4 h-4" /> Book & Print
            </Button>
        </div>
    </div>
</div>
