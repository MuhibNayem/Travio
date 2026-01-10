<script lang="ts">
    import Seat from "./Seat.svelte";
    import { createEventDispatcher } from "svelte";

    export let rows: number = 10;
    export let columns: number = 4; // Total columns excluding aisle (e.g. 4 for 2+2)
    export let aisleIndex: number = 2; // Index after which aisle is placed (0-indexed)
    export let seatData: Record<string, any> = {}; // id -> { status, price, category }

    const dispatch = createEventDispatcher();

    // Generate seat grid
    // Rows: A, B, C...
    // Cols: 1, 2, 3...

    function getSeatLabel(row: number, col: number) {
        const rowChar = String.fromCharCode(65 + row); // A, B, C
        return `${rowChar}${col + 1}`;
    }

    function handleSeatClick(e: CustomEvent) {
        dispatch("seatClick", e.detail);
    }
</script>

<div
    class="flex flex-col items-center gap-4 p-6 bg-gray-50 rounded-lg border shadow-inner max-w-fit mx-auto"
>
    <!-- Driver Section -->
    <div class="w-full flex justify-end mb-4 pr-2">
        <div
            class="w-10 h-10 rounded-full border-2 border-gray-400 flex items-center justify-center bg-white shadow-sm"
        >
            <span class="text-xs text-gray-500 font-bold">Driver</span>
        </div>
    </div>

    <!-- Seats Container -->
    <div class="flex flex-col gap-3">
        {#each Array(rows) as _, r}
            <div class="flex gap-3">
                {#each Array(columns) as _, c}
                    <!-- Aisle Spacer -->
                    {#if c === aisleIndex}
                        <div class="w-8"></div>
                    {/if}

                    {@const label = getSeatLabel(r, c)}
                    {@const data = seatData[label] || {}}

                    <Seat
                        id={label}
                        {label}
                        status={data.status || "available"}
                        price={data.price || 500}
                        category={data.category || "Economy"}
                        on:click={handleSeatClick}
                    />
                {/each}
            </div>
        {/each}
    </div>

    <!-- Legend -->
    <div class="flex gap-4 mt-6 text-xs text-gray-600">
        <div class="flex items-center gap-1">
            <div
                class="w-3 h-3 border border-gray-300 rounded-sm bg-white"
            ></div>
             Available
        </div>
        <div class="flex items-center gap-1">
            <div
                class="w-3 h-3 border border-primary rounded-sm bg-primary"
            ></div>
             Selected
        </div>
        <div class="flex items-center gap-1">
            <div class="w-3 h-3 bg-gray-200 rounded-sm"></div>
             Sold
        </div>
    </div>
</div>
