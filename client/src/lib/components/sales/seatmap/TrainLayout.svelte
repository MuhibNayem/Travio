<script lang="ts">
    import Seat from "./Seat.svelte";
    import { createEventDispatcher } from "svelte";
    import { Badge } from "$lib/components/ui/badge";

    export let coaches: Array<{
        id: string;
        name: string;
        class: string; // AC, Sleeper, S_Chair
        rows: number;
        seatsPerRow: number; // 4 for 2+2, 6 for 3+3
        hasBerths: boolean;
    }> = [];

    export let seatData: Record<string, any> = {};

    const dispatch = createEventDispatcher();
    let selectedCoachId: string | null = null;
    $: activeCoach =
        coaches.find((c) => c.id === selectedCoachId) || coaches[0];

    function getSeatLabel(coachId: string, row: number, col: number) {
        return `${coachId}-${String.fromCharCode(65 + row)}${col + 1}`;
    }

    function handleSeatClick(e: CustomEvent) {
        dispatch("seatClick", { ...e.detail, coachId: activeCoach?.id });
    }

    // Color coding by class
    const classColors: Record<string, string> = {
        AC: "bg-blue-50 border-blue-200",
        Sleeper: "bg-purple-50 border-purple-200",
        S_Chair: "bg-green-50 border-green-200",
    };
</script>

<div class="flex flex-col gap-6">
    <!-- Coach Selector -->
    <div class="flex flex-wrap gap-3 justify-center">
        {#each coaches as coach}
            <button
                type="button"
                class="flex flex-col items-center p-3 border rounded-lg transition-all hover:shadow-md hover:border-primary min-w-[100px] {classColors[
                    coach.class
                ] || 'bg-gray-50'}"
                class:ring-2={activeCoach?.id === coach.id}
                class:ring-primary={activeCoach?.id === coach.id}
                onclick={() => (selectedCoachId = coach.id)}
            >
                <span class="font-bold text-sm">{coach.name}</span>
                <Badge variant="secondary" class="mt-1 text-[10px]"
                    >{coach.class}</Badge
                >
            </button>
        {/each}
    </div>

    <!-- Train Car Visual -->
    {#if activeCoach}
        <div
            class="bg-gray-100 rounded-xl p-6 border shadow-inner max-w-3xl mx-auto"
        >
            <!-- Train Header -->
            <div
                class="flex items-center justify-between mb-4 pb-2 border-b border-gray-300"
            >
                <span class="font-bold text-lg">{activeCoach.name}</span>
                <Badge variant="outline">{activeCoach.class}</Badge>
            </div>

            <!-- Toilet/Entry Indicator -->
            <div class="flex justify-between text-xs text-gray-500 mb-4">
                <div
                    class="flex items-center gap-1 bg-gray-200 px-2 py-1 rounded"
                >
                    <span>🚻</span> Entry
                </div>
                <div
                    class="flex items-center gap-1 bg-gray-200 px-2 py-1 rounded"
                >
                    Exit <span>🚪</span>
                </div>
            </div>

            <!-- Seat Grid -->
            <div class="flex flex-col gap-2">
                {#each Array(activeCoach.rows) as _, r}
                    <div class="flex gap-2 items-center">
                        <!-- Row Number -->
                        <span class="w-6 text-xs text-gray-400 text-right"
                            >{r + 1}</span
                        >

                        <!-- Left Seats -->
                        <div class="flex gap-1">
                            {#each Array(Math.floor(activeCoach.seatsPerRow / 2)) as _, c}
                                {@const label = getSeatLabel(
                                    activeCoach.id,
                                    r,
                                    c,
                                )}
                                {@const data = seatData[label] || {}}
                                <Seat
                                    id={label}
                                    label={`${c + 1}`}
                                    status={data.status || "available"}
                                    price={data.price || 0}
                                    category={activeCoach.class}
                                    on:click={handleSeatClick}
                                    className="w-9 h-9"
                                />
                            {/each}
                        </div>

                        <!-- Aisle -->
                        <div
                            class="w-8 flex items-center justify-center text-gray-300 text-xs"
                        >
                            │
                        </div>

                        <!-- Right Seats -->
                        <div class="flex gap-1">
                            {#each Array(Math.ceil(activeCoach.seatsPerRow / 2)) as _, c}
                                {@const actualCol =
                                    Math.floor(activeCoach.seatsPerRow / 2) + c}
                                {@const label = getSeatLabel(
                                    activeCoach.id,
                                    r,
                                    actualCol,
                                )}
                                {@const data = seatData[label] || {}}
                                <Seat
                                    id={label}
                                    label={`${actualCol + 1}`}
                                    status={data.status || "available"}
                                    price={data.price || 0}
                                    category={activeCoach.class}
                                    on:click={handleSeatClick}
                                    className="w-9 h-9"
                                />
                            {/each}
                        </div>

                        <!-- Berth Indicator (for sleeper) -->
                        {#if activeCoach.hasBerths}
                            <div class="ml-2 text-xs text-gray-400">L/M/U</div>
                        {/if}
                    </div>
                {/each}
            </div>

            <!-- Legend -->
            <div
                class="flex gap-4 mt-6 pt-4 border-t text-xs text-gray-600 justify-center"
            >
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
    {/if}
</div>
