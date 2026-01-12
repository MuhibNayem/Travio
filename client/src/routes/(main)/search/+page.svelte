<script lang="ts">
    import TripCard from "$lib/components/search/TripCard.svelte";
    import SearchHero from "$lib/components/blocks/SearchHero.svelte";
    import { Button } from "$lib/components/ui/button";
    import { Filter } from "@lucide/svelte";
    import { browser } from "$app/environment";
    import { stationsStore } from "$lib/stores/stations.svelte";
    import type { PageData } from "./$types";

    export let data: PageData;

    $: results = data.results;
    $: total = data.total;
    $: params = data.searchParams;
    let fromName = "";
    let toName = "";
    let lastFrom = "";
    let lastTo = "";

    $: fromLabel =
        stationsStore.stations.find((station) => station.id === params.from)
            ?.name ||
        fromName ||
        params.from;
    $: toLabel =
        stationsStore.stations.find((station) => station.id === params.to)?.name ||
        toName ||
        params.to;

    async function loadStationNames(fromId: string, toId: string) {
        if (!fromId && !toId) return;
        const [fromResult, toResult] = await Promise.allSettled([
            fromId ? stationsStore.getStationById(fromId) : Promise.resolve(null),
            toId ? stationsStore.getStationById(toId) : Promise.resolve(null),
        ]);

        if (fromResult.status === "fulfilled" && fromResult.value) {
            fromName = fromResult.value.name || "";
        }
        if (toResult.status === "fulfilled" && toResult.value) {
            toName = toResult.value.name || "";
        }
    }

    $: if (browser && (params.from !== lastFrom || params.to !== lastTo)) {
        lastFrom = params.from;
        lastTo = params.to;
        fromName = "";
        toName = "";
        void loadStationNames(params.from, params.to);
    }
</script>

<div class="min-h-screen bg-muted/30 pb-20">
    <!-- Compact Search Hero for Results Page -->
    <div class="bg-white border-b sticky top-0 z-40 shadow-sm">
        <!-- Reuse SearchHero but maybe controlled via props to be compact? 
              For now, just render it. It might be large but acceptable. -->
        <div class="scale-90 origin-top -mb-32">
            <SearchHero />
        </div>
    </div>

    <div class="container mx-auto px-4 pt-40">
        <div class="flex flex-col gap-6 lg:flex-row">
            <!-- Filters Sidebar (Desktop) -->
            <div class="hidden w-64 flex-col gap-6 lg:flex">
                <div class="rounded-xl border bg-white p-6 shadow-sm">
                    <div class="flex items-center gap-2 mb-4">
                        <Filter size={20} />
                        <h3 class="font-bold">Filters</h3>
                    </div>
                    <!-- Add filters here later -->
                    <p class="text-sm text-muted-foreground">
                        Price, Operator, Type filters coming soon.
                    </p>
                </div>
            </div>

            <!-- Results List -->
            <div class="flex-1">
                <div class="mb-6 flex items-center justify-between">
                    <h2 class="text-2xl font-bold">
                        {total}
                        {total === 1 ? "Trip" : "Trips"} Found
                        <span
                            class="ml-2 text-base font-medium text-muted-foreground"
                        >
                            for {fromLabel} to {toLabel}
                        </span>
                    </h2>
                    <Button variant="outline" size="sm" class="lg:hidden">
                        <Filter size={16} class="mr-2" /> Filters
                    </Button>
                </div>

                {#if results.length > 0}
                    <div class="flex flex-col gap-4">
                        {#each results as trip (trip.id)}
                            <TripCard {trip} />
                        {/each}
                    </div>
                {:else}
                    <div
                        class="flex flex-col items-center justify-center rounded-2xl border bg-white p-12 text-center shadow-sm"
                    >
                        <div class="mb-4 rounded-full bg-muted p-4">
                            <Filter size={32} class="text-muted-foreground" />
                        </div>
                        <h3 class="text-xl font-bold">No trips found</h3>
                        <p class="text-muted-foreground max-w-md mt-2">
                            We couldn't find any trips for this route on the
                            selected date. Try changing the date or search for a
                            different route.
                        </p>
                    </div>
                {/if}
            </div>
        </div>
    </div>
</div>
