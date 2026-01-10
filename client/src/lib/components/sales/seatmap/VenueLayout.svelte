<script lang="ts">
    import Seat from "./Seat.svelte";
    import { createEventDispatcher } from "svelte";
    import * as Card from "$lib/components/ui/card";
    import { Badge } from "$lib/components/ui/badge";

    export let zones: Array<{
        id: string;
        name: string;
        price: number;
        rows: number;
        cols: number;
        color: string;
    }> = [];

    export let seatData: Record<string, any> = {};

    const dispatch = createEventDispatcher();
    let selectedZoneId: string | null = null;
    $: activeZone = zones.find((z) => z.id === selectedZoneId) || zones[0];

    function handleSeatClick(e: CustomEvent) {
        dispatch("seatClick", { ...e.detail, zoneId: activeZone?.id });
    }
</script>

<div class="flex flex-col gap-6">
    <!-- Zone Selector -->
    <div class="flex flex-wrap gap-4 justify-center">
        {#each zones as zone}
            <button
                type="button"
                class="flex flex-col items-center p-4 border rounded-lg transition-all hover:shadow-md hover:border-primary w-32"
                class:ring-2={activeZone?.id === zone.id}
                class:ring-primary={activeZone?.id === zone.id}
                on:click={() => (selectedZoneId = zone.id)}
            >
                <span class="font-bold text-lg">{zone.name}</span>
                <Badge variant="secondary" class="mt-2">৳{zone.price}</Badge>
            </button>
        {/each}
    </div>

    <!-- Stage / Screen -->
    <div
        class="w-full h-12 bg-gray-800 text-white rounded-t-3xl flex items-center justify-center mb-0 shadow-lg mx-auto max-w-2xl transform perspective-1000 rotate-x-12 opacity-80"
    >
        <span class="uppercase tracking-[0.5em] text-sm font-light"
            >Stage / Screen</span
        >
    </div>

    <!-- Active Zone Grid -->
    {#if activeZone}
        <Card.Root class="mx-auto border-none shadow-none bg-transparent">
            <Card.Content class="pt-6">
                <h3
                    class="text-center mb-6 font-semibold text-muted-foreground"
                >
                    Select Seats in {activeZone.name}
                </h3>

                <div class="flex flex-col gap-3 items-center">
                    {#each Array(activeZone.rows) as _, r}
                        <div class="flex gap-2">
                            {#each Array(activeZone.cols) as _, c}
                                {@const label = `${activeZone.name[0]}-${String.fromCharCode(65 + r)}${c + 1}`}
                                {@const data = seatData[label] || {}}

                                <Seat
                                    id={label}
                                    label={`${String.fromCharCode(65 + r)}${c + 1}`}
                                    status={data.status || "available"}
                                    price={activeZone.price}
                                    category={activeZone.name}
                                    on:click={handleSeatClick}
                                    className="w-9 h-9 text-[10px]"
                                />
                            {/each}
                        </div>
                    {/each}
                </div>
            </Card.Content>
        </Card.Root>
    {/if}
</div>
