<script lang="ts">
    import { page } from "$app/stores";
    import TripCard from "$lib/components/blocks/TripCard.svelte";
    import { MOCK_TRIPS, STATIONS } from "$lib/mocks/data";
    import { Button } from "$lib/components/ui/button";
    import { Filter } from "@lucide/svelte";

    let fromId = $state("");
    let toId = $state("");
    let type = $state("");
    let date = $state("");

    $effect(() => {
        fromId = $page.url.searchParams.get("from") || "";
        toId = $page.url.searchParams.get("to") || "";
        type = $page.url.searchParams.get("type") || "bus";
        date = $page.url.searchParams.get("date") || "";
    });

    let filteredTrips = $derived(
        MOCK_TRIPS.filter((t) => {
            // In a real app we'd filter by from/to IDs on the backend
            // For mock, we just show all if matches type, or all if mock data is limited
            return type ? t.type === type : true;
        }),
    );

    let fromName = $derived(
        STATIONS.find((s) => s.id === fromId)?.name || "Origin",
    );
    let toName = $derived(
        STATIONS.find((s) => s.id === toId)?.name || "Destination",
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
                {filteredTrips.length} available trips found for {date ||
                    "today"}
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
                    <!-- Mock Filters -->
                    <div class="space-y-4">
                        <div>
                            <label
                                class="mb-2 block text-xs font-bold uppercase text-muted-foreground"
                                >Operators</label
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
                                    Green Line
                                </label>
                                <label
                                    class="flex items-center gap-2 text-sm text-foreground"
                                >
                                    <input
                                        type="checkbox"
                                        class="rounded border-gray-300"
                                        checked
                                    />
                                    Desh Travels
                                </label>
                            </div>
                        </div>
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
                                    AC Business
                                </label>
                                <label
                                    class="flex items-center gap-2 text-sm text-foreground"
                                >
                                    <input
                                        type="checkbox"
                                        class="rounded border-gray-300"
                                        checked
                                    />
                                    Non-AC
                                </label>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- List -->
            <div class="flex flex-1 flex-col gap-4">
                {#each filteredTrips as trip (trip.id)}
                    <TripCard {trip} />
                {/each}

                {#if filteredTrips.length === 0}
                    <div
                        class="flex h-64 flex-col items-center justify-center rounded-2xl border-2 border-dashed border-muted-foreground/20 bg-muted/10 p-10 text-center"
                    >
                        <p class="text-lg font-bold text-muted-foreground">
                            No trips found
                        </p>
                        <p class="text-sm text-muted-foreground/80">
                            Try changing your search criteria
                        </p>
                        <Button variant="link" class="mt-4" href="/dashboard"
                            >Go Back</Button
                        >
                    </div>
                {/if}
            </div>
        </div>
    </div>
</div>
