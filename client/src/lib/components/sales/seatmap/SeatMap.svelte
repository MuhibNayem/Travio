<script lang="ts">
    import BusLayout from "./BusLayout.svelte";
    import TrainLayout from "./TrainLayout.svelte";
    import LaunchLayout from "./LaunchLayout.svelte";
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
            aisleIndex={config.aisleIndex ?? 2}
            {seatData}
            on:seatClick={handleSeatClick}
        />
    {:else if type === "train"}
        <TrainLayout
            coaches={config.coaches || [
                {
                    id: "S1",
                    name: "S1",
                    class: "S_Chair",
                    rows: 15,
                    seatsPerRow: 6,
                    hasBerths: false,
                },
                {
                    id: "AC1",
                    name: "AC1",
                    class: "AC",
                    rows: 12,
                    seatsPerRow: 4,
                    hasBerths: false,
                },
            ]}
            {seatData}
            on:seatClick={handleSeatClick}
        />
    {:else if type === "launch"}
        <LaunchLayout
            decks={config.decks || [
                {
                    id: "D1",
                    name: "Deck 1",
                    type: "economy",
                    rows: 8,
                    cols: 6,
                    seatPrice: 500,
                },
                {
                    id: "VIP",
                    name: "VIP Cabin",
                    type: "vip_cabin",
                    cabins: [
                        { id: "C1", beds: 4, price: 2000 },
                        { id: "C2", beds: 2, price: 3000 },
                    ],
                },
            ]}
            {seatData}
            on:seatClick={handleSeatClick}
        />
    {:else if type === "event"}
        <VenueLayout
            zones={config.zones || []}
            {seatData}
            on:seatClick={handleSeatClick}
        />
    {/if}
</div>
