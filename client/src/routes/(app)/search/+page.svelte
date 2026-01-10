<script lang="ts">
    import { page } from "$app/stores";
    import TripCard from "$lib/components/blocks/TripCard.svelte";
    import { Button } from "$lib/components/ui/button";
    import { Filter, Loader } from "@lucide/svelte";
    import { onMount } from "svelte";
    import { fade } from "svelte/transition";
    import { searchApi } from "$lib/api/search";
    import { catalogApi, type Station } from "$lib/api/catalog";
    import type { Trip, TransportType } from "$lib/types/transport";

    let trips = $state<Trip[]>([]);
    let isLoading = $state(true);
    let stations = $state<Station[]>([]);

    // Search Params
    let fromId = $state("");
    let toId = $state("");
    let type = $state("");
    let date = $state("");

    // Initialize & Fetch Stations
    onMount(async () => {
        try {
            stations = await catalogApi.getStations();
        } catch (error) {
            console.error("Failed to load stations", error);
        }
    });

    // React to URL changes
    $effect(() => {
        fromId = $page.url.searchParams.get("from") || "";
        toId = $page.url.searchParams.get("to") || "";
        type = $page.url.searchParams.get("type") || "";
        date = $page.url.searchParams.get("date") || "";

        // Trigger fetch if we have somewhat valid params (at least date)
        if (date) {
            fetchTripsWithResolution();
        } else {
            isLoading = false;
        }
    });

    // Helper to get city name from ID
    function getCity(id: string): string {
        const s = stations.find((s) => s.id === id);
        return s ? s.city : id;
    }

    async function fetchTripsWithResolution() {
        isLoading = true;

        // Wait for stations if needed? If stations are loading, this might run before they are ready.
        // But since onMount fires async, and $effect fires on navigation/mount.
        // We might want to wait. But let's rely on basic resolution or ID fallback.

        const fromStation = stations.find((s) => s.id === fromId);
        const toStation = stations.find((s) => s.id === toId);

        const fromCity = fromStation ? fromStation.city : fromId;
        const toCity = toStation ? toStation.city : toId;

        // Ensure we don't search with empty cities if IDs are empty (e.g. initial load without params)
        if (!fromCity && !toCity) {
            isLoading = false;
            return;
        }

        try {
            const resp = await searchApi.searchTrips({
                from: fromCity,
                to: toCity,
                date: date,
                type: type,
            });

            // Map API (snake_case) to Client (camelCase)
            trips = resp.results.map((r) => ({
                id: r.id,
                routeId: r.route_id,
                type: r.type as TransportType,
                operator: r.operator,
                vehicleName: r.vehicle_name,
                departureTime: r.departure_time,
                arrivalTime: r.arrival_time,
                price: r.price,
                class: r.class,
                availableSeats: r.available_seats,
                totalSeats: r.total_seats,
            }));
        } catch (error) {
            console.error("Search failed", error);
            trips = [];
        } finally {
            isLoading = false;
        }
    }

    // React to stations change -> Re-fetch if needed (to resolve city names correctly)
    $effect(() => {
        if (stations.length > 0 && date) {
            fetchTripsWithResolution();
        }
    });

    let fromName = $derived(
        stations.find((s) => s.id === fromId)?.name || fromId || "Origin",
    );
    let toName = $derived(
        stations.find((s) => s.id === toId)?.name || toId || "Destination",
    );
</script>

<div class="min-h-screen bg-muted/30 pb-20">
    <!-- Header -->
    <div class="bg-[#101922] pb-24 pt-32 text-white">
        <div class="container mx-auto max-w-5xl px-4">
            <h1 class="text-3xl font-bold md:text-5xl">
                {fromName} to {toName}
            </h1>
            <p class="mt-4 text-blue-200">
                {#if isLoading}
                    Searching...
                {:else}
                    {trips.length} available trips found for {date || "today"}
                {/if}
            </p>
        </div>
    </div>

    <!-- Content -->
    <div class="container mx-auto -mt-16 max-w-5xl px-4">
        <div class="flex flex-col gap-6 lg:flex-row">
            <!-- Sidebar Filters (Desktop) -->
            <div class="hidden w-64 flex-col gap-6 lg:flex">
                <div class="glass-panel p-6">
                    <div
                        class="mb-4 flex items-center gap-2 font-bold text-foreground"
                    >
                        <Filter size={18} /> Filters
                    </div>
                    <!-- Mock Filters for now -->
                    <div class="space-y-4">
                        <div>
                            <label
                                class="mb-2 block text-xs font-bold uppercase text-muted-foreground"
                                >Class</label
                            >
                            <div class="flex flex-col gap-2">
                                <label
                                    class="flex items-center gap-2 text-sm text-foreground"
                                >
                                    <input
                                        type="checkbox"
                                        class="rounded border-gray-300"
                                        checked
                                    />
                                    All Classes
                                </label>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- List -->
            <div class="flex flex-1 flex-col gap-4">
                {#if isLoading}
                    <div class="flex h-64 items-center justify-center">
                        <Loader class="h-8 w-8 animate-spin text-primary" />
                    </div>
                {:else}
                    {#each trips as trip (trip.id)}
                        <div transition:fade={{ duration: 300 }}>
                            <TripCard {trip} {fromId} {toId} />
                        </div>
                    {/each}

                    {#if trips.length === 0}
                        <div
                            transition:fade
                            class="flex h-64 flex-col items-center justify-center rounded-2xl border-2 border-dashed border-muted-foreground/20 bg-muted/10 p-10 text-center"
                        >
                            <p class="text-lg font-bold text-muted-foreground">
                                No trips found
                            </p>
                            <p class="text-sm text-muted-foreground/80">
                                Try changing your search criteria
                            </p>
                            <Button
                                variant="link"
                                class="mt-4"
                                href="/dashboard">Go Back</Button
                            >
                        </div>
                    {/if}
                {/if}
            </div>
        </div>
    </div>
</div>
