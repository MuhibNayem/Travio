<script lang="ts">
    import BusLayout from "./BusLayout.svelte";
    import VenueLayout from "./VenueLayout.svelte";
    import { createEventDispatcher } from "svelte";

    export let type: "bus" | "train" | "launch" | "event" = "bus";
    export let config: any = {}; // layout config (rows, cols, etc)
    export let bookedSeats: string[] = [];
    export let blockedSeats: string[] = [];

    const dispatch = createEventDispatcher();

    // Local state for current selection (to show 'selected' status visually)
    let selectedSeats: string[] = [];

    // Combine all data for child components
    $: seatData = [
        ...bookedSeats.map((id) => [id, { status: "sold" }]),
        ...blockedSeats.map((id) => [id, { status: "blocked" }]),
        ...selectedSeats.map((id) => [id, { status: "selected" }]),
    ].reduce((acc: any, [id, data]) => {
        acc[id as string] = data;
        return acc;
    }, {});

    function handleSeatClick(e: CustomEvent) {
        const { id, status } = e.detail;

        if (status === "available") {
            selectedSeats = [...selectedSeats, id];
        } else if (status === "selected") {
            selectedSeats = selectedSeats.filter((s) => s !== id);
        }

        dispatch("selectionChange", selectedSeats);
    }
</script>

<div class="w-full overflow-x-auto py-4">
    {#if type === "bus"}
        <BusLayout
            rows={config.rows || 10}
            columns={config.columns || 4}
            aisleIndex={config.aisleIndex ?? 1}
            {seatData}
            on:seatClick={handleSeatClick}
        />
    {:else if type === "event"}
        <VenueLayout
            zones={config.zones || []}
            {seatData}
            on:seatClick={handleSeatClick}
        />
    {:else}
        <!-- Fallback / TODO for Train/Launch -->
        <div class="p-8 text-center text-gray-500">
            {type} layout coming soon
        </div>
    {/if}
</div>
