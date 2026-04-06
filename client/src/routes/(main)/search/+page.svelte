<script lang="ts">
    import { page } from "$app/stores";
    import TripCard from "$lib/components/search/TripCard.svelte";
    import SearchHero from "$lib/components/blocks/SearchHero.svelte";
    import { Button } from "$lib/components/ui/button";
    import { Slider } from "$lib/components/ui/slider";
    import { Checkbox } from "$lib/components/ui/checkbox";
    import { Separator } from "$lib/components/ui/separator";
    import { Filter, Loader2, ArrowUpDown, X } from "@lucide/svelte";
    import type { PageData } from "./$types";

    export let data: PageData;

    $: results = data.results || [];
    $: total = data.total || 0;
    $: params = data.searchParams || {};
    $: isLoading = data.loading || false;

    // Filters
    let priceRange = $state([0, 2000]);
    let selectedOperators = $state<string[]>([]);
    let selectedTypes = $state<string[]>([]);
    let selectedDeparture = $state<string>("");
    let sortBy = $state<"price" | "duration" | "departure">("price");
    let showFiltersMobile = $state(false);

    const operators = ["Green Line", "Shohagh", "Ena Transport", "Hanif"];
    const vehicleTypes = ["AC", "Non-AC", "Sleeper", "Premium"];
    const departureTimes = [
        { value: "morning", label: "Morning (6AM-12PM)" },
        { value: "afternoon", label: "Afternoon (12PM-5PM)" },
        { value: "evening", label: "Evening (5PM-9PM)" },
        { value: "night", label: "Night (9PM-6AM)" },
    ];

    function toggleFilter(list: string[], item: string): string[] {
        if (list.includes(item)) {
            return list.filter((i) => i !== item);
        }
        return [...list, item];
    }

    function clearAllFilters() {
        priceRange = [0, 2000];
        selectedOperators = [];
        selectedTypes = [];
        selectedDeparture = "";
    }

    $: filteredResults = results.filter((trip: any) => {
        const price = trip.price || trip.pricing?.base_price_paisa / 100 || 0;
        if (price < priceRange[0] || price > priceRange[1]) return false;
        if (
            selectedOperators.length > 0 &&
            !selectedOperators.includes(trip.operator)
        )
            return false;
        if (
            selectedTypes.length > 0 &&
            !selectedTypes.includes(trip.class)
        )
            return false;
        return true;
    });

    $: sortedResults = [...filteredResults].sort((a: any, b: any) => {
        if (sortBy === "price") {
            return (
                (a.price || a.pricing?.base_price_paisa / 100 || 0) -
                (b.price || b.pricing?.base_price_paisa / 100 || 0)
            );
        }
        if (sortBy === "departure") {
            return (
                new Date(a.departure_time).getTime() -
                new Date(b.departure_time).getTime()
            );
        }
        return 0;
    });

    const hasActiveFilters =
        priceRange[0] > 0 ||
        priceRange[1] < 2000 ||
        selectedOperators.length > 0 ||
        selectedTypes.length > 0 ||
        selectedDeparture !== "";
</script>

<div class="min-h-screen bg-muted/30 pb-20">
    <!-- Sticky Search Header -->
    <div class="sticky top-0 z-40 border-b bg-white/80 shadow-sm backdrop-blur-md dark:bg-[#101922]/80">
        <div class="container mx-auto px-4 py-3">
            <div class="flex items-center justify-between">
                <div class="flex items-center gap-4">
                    <h2 class="text-lg font-bold">
                        {isLoading
                            ? "Searching..."
                            : `${sortedResults.length} Trips Found`}
                    </h2>
                    <span class="hidden text-sm text-muted-foreground sm:block">
                        {params.from} to {params.to}
                    </span>
                </div>
                <div class="flex items-center gap-2">
                    <div class="flex items-center gap-2">
                        <ArrowUpDown size={16} class="text-muted-foreground" />
                        <select
                            bind:value={sortBy}
                            class="rounded-lg border border-border bg-transparent px-2 py-1 text-sm"
                        >
                            <option value="price">Cheapest</option>
                            <option value="departure">Earliest</option>
                        </select>
                    </div>
                    <Button
                        variant="outline"
                        size="sm"
                        class="lg:hidden"
                        onclick={() =>
                            (showFiltersMobile = !showFiltersMobile)}
                    >
                        <Filter size={16} class="mr-2" />
                        Filters
                    </Button>
                </div>
            </div>
        </div>
    </div>

    <div class="container mx-auto px-4 pt-4">
        <div class="flex flex-col gap-6 lg:flex-row">
            <!-- Filters Sidebar -->
            <div
                class="{showFiltersMobile ? 'fixed inset-0 z-50 bg-black/50 lg:relative lg:bg-transparent' : 'hidden lg:block'} lg:w-64"
            >
                <div
                    class="{showFiltersMobile ? 'absolute right-0 top-0 h-full w-80 overflow-y-auto bg-white shadow-2xl dark:bg-[#101922] p-6 animate-in slide-in-from-right' : ''} lg:sticky lg:top-20 lg:flex lg:flex-col lg:gap-6"
                >
                    {#if showFiltersMobile}
                        <div class="mb-4 flex items-center justify-between">
                            <h3 class="text-lg font-bold">Filters</h3>
                            <Button
                                variant="ghost"
                                size="sm"
                                onclick={() => (showFiltersMobile = false)}
                            >
                                <X size={20} />
                            </Button>
                        </div>
                    {/if}

                    <!-- Price Range -->
                    <div class="rounded-xl border bg-card p-4 shadow-sm">
                        <h4 class="mb-3 font-semibold">Price Range</h4>
                        <div class="space-y-3">
                            <div class="flex justify-between text-sm">
                                <span>৳{priceRange[0]}</span>
                                <span>৳{priceRange[1]}</span>
                            </div>
                            <input
                                type="range"
                                min="0"
                                max="2000"
                                step="50"
                                bind:value={priceRange[1]}
                                class="w-full"
                            />
                        </div>
                    </div>

                    <Separator />

                    <!-- Vehicle Type -->
                    <div class="rounded-xl border bg-card p-4 shadow-sm">
                        <h4 class="mb-3 font-semibold">Vehicle Type</h4>
                        <div class="space-y-2">
                            {#each vehicleTypes as type}
                                <label class="flex items-center gap-2">
                                    <Checkbox
                                        checked={selectedTypes.includes(type)}
                                        on:change={() =>
                                            (selectedTypes = toggleFilter(
                                                selectedTypes,
                                                type,
                                            ))}
                                    />
                                    <span class="text-sm">{type}</span>
                                </label>
                            {/each}
                        </div>
                    </div>

                    <Separator />

                    <!-- Operator -->
                    <div class="rounded-xl border bg-card p-4 shadow-sm">
                        <h4 class="mb-3 font-semibold">Operator</h4>
                        <div class="space-y-2">
                            {#each operators as operator}
                                <label class="flex items-center gap-2">
                                    <Checkbox
                                        checked={selectedOperators.includes(
                                            operator,
                                        )}
                                        on:change={() =>
                                            (selectedOperators = toggleFilter(
                                                selectedOperators,
                                                operator,
                                            ))}
                                    />
                                    <span class="text-sm">{operator}</span>
                                </label>
                            {/each}
                        </div>
                    </div>

                    {#if hasActiveFilters}
                        <Button
                            variant="outline"
                            class="w-full"
                            onclick={clearAllFilters}
                        >
                            Clear All Filters
                        </Button>
                    {/if}
                </div>
            </div>

            <!-- Results List -->
            <div class="flex-1">
                {#if isLoading}
                    <!-- Skeleton Loading -->
                    <div class="flex flex-col gap-4">
                        {#each Array(5) as _}
                            <div
                                class="animate-pulse rounded-xl border bg-card p-6"
                            >
                                <div class="flex items-center justify-between">
                                    <div class="space-y-2">
                                        <div
                                            class="h-6 w-40 rounded bg-muted"
                                        ></div>
                                        <div
                                            class="h-4 w-60 rounded bg-muted"
                                        ></div>
                                    </div>
                                    <div class="text-right">
                                        <div
                                            class="ml-auto h-8 w-24 rounded bg-muted"
                                        ></div>
                                        <div
                                            class="mt-2 h-4 w-16 rounded bg-muted"
                                        ></div>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                {:else if sortedResults.length > 0}
                    <div class="flex flex-col gap-4">
                        {#each sortedResults as trip (trip.id)}
                            <TripCard {trip} />
                        {/each}
                    </div>
                {:else}
                    <div
                        class="flex flex-col items-center justify-center rounded-2xl border bg-white p-12 text-center shadow-sm"
                    >
                        <div class="mb-4 rounded-full bg-muted p-4">
                            <Filter
                                size={32}
                                class="text-muted-foreground"
                            />
                        </div>
                        <h3 class="text-xl font-bold">No trips found</h3>
                        <p class="mt-2 max-w-md text-muted-foreground">
                            We couldn't find any trips for this route on the
                            selected date. Try changing the date or search for a
                            different route.
                        </p>
                        {#if hasActiveFilters}
                            <Button
                                class="mt-4"
                                variant="outline"
                                onclick={clearAllFilters}
                            >
                                Clear Filters
                            </Button>
                        {/if}
                    </div>
                {/if}
            </div>
        </div>
    </div>
</div>
