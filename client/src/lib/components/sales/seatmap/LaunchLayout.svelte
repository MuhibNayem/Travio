<script lang="ts">
    import Seat from "./Seat.svelte";
    import { createEventDispatcher } from "svelte";
    import { Badge } from "$lib/components/ui/badge";
    import { Ship, Users, Crown } from "@lucide/svelte";

    export let decks: Array<{
        id: string;
        name: string;
        type: "economy" | "business" | "vip_cabin";
        rows?: number;
        cols?: number;
        cabins?: Array<{ id: string; beds: number; price: number }>;
        seatPrice?: number;
    }> = [];

    export let seatData: Record<string, any> = {};

    const dispatch = createEventDispatcher();
    let selectedDeckId: string | null = null;
    $: activeDeck = decks.find((d) => d.id === selectedDeckId) || decks[0];

    function getSeatLabel(deckId: string, row: number, col: number) {
        return `${deckId}-${String.fromCharCode(65 + row)}${col + 1}`;
    }

    function getCabinLabel(deckId: string, cabinId: string, bed: number) {
        return `${deckId}-${cabinId}-B${bed}`;
    }

    function handleSeatClick(e: CustomEvent) {
        dispatch("seatClick", { ...e.detail, deckId: activeDeck?.id });
    }

    // Type icons
    const deckIcons = {
        economy: Users,
        business: Ship,
        vip_cabin: Crown,
    };
</script>

<div class="flex flex-col gap-6">
    <!-- Deck Selector -->
    <div class="flex flex-wrap gap-3 justify-center">
        {#each decks as deck}
            {@const DeckIcon = deckIcons[deck.type] || Ship}
            <button
                type="button"
                class="flex flex-col items-center p-4 border rounded-xl transition-all hover:shadow-lg hover:border-primary min-w-[120px]
                    {deck.type === 'vip_cabin'
                    ? 'bg-amber-50 border-amber-200'
                    : deck.type === 'business'
                      ? 'bg-blue-50 border-blue-200'
                      : 'bg-gray-50 border-gray-200'}"
                class:ring-2={activeDeck?.id === deck.id}
                class:ring-primary={activeDeck?.id === deck.id}
                onclick={() => (selectedDeckId = deck.id)}
            >
                <DeckIcon size={24} class="mb-1 text-primary" />
                <span class="font-bold text-sm">{deck.name}</span>
                <Badge variant="secondary" class="mt-1 text-[10px] capitalize"
                    >{deck.type.replace("_", " ")}</Badge
                >
            </button>
        {/each}
    </div>

    <!-- Deck Layout -->
    {#if activeDeck}
        <div
            class="bg-gradient-to-b from-blue-50 to-blue-100 rounded-2xl p-6 border border-blue-200 shadow-inner max-w-3xl mx-auto"
        >
            <!-- Ship Header -->
            <div class="flex items-center justify-center gap-2 mb-6">
                <div class="h-px flex-1 bg-blue-300"></div>
                <div
                    class="px-4 py-1 bg-blue-600 text-white rounded-full text-sm font-bold"
                >
                    {activeDeck.name}
                </div>
                <div class="h-px flex-1 bg-blue-300"></div>
            </div>

            {#if activeDeck.type === "vip_cabin" && activeDeck.cabins}
                <!-- Cabin Layout -->
                <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
                    {#each activeDeck.cabins as cabin}
                        <div
                            class="bg-white rounded-lg p-4 border-2 border-amber-300 shadow-sm"
                        >
                            <div
                                class="font-bold text-center mb-2 text-amber-800"
                            >
                                Cabin {cabin.id}
                            </div>
                            <div class="flex flex-wrap gap-2 justify-center">
                                {#each Array(cabin.beds) as _, b}
                                    {@const label = getCabinLabel(
                                        activeDeck.id,
                                        cabin.id,
                                        b + 1,
                                    )}
                                    {@const data = seatData[label] || {}}
                                    <Seat
                                        id={label}
                                        label={`B${b + 1}`}
                                        status={data.status || "available"}
                                        price={cabin.price}
                                        category="VIP Cabin"
                                        on:click={handleSeatClick}
                                        className="w-10 h-10"
                                    />
                                {/each}
                            </div>
                            <div class="text-center mt-2 text-xs text-gray-500">
                                ৳{cabin.price}/bed
                            </div>
                        </div>
                    {/each}
                </div>
            {:else}
                <!-- Open Seating Grid -->
                <div class="flex flex-col gap-2 items-center">
                    {#each Array(activeDeck.rows || 8) as _, r}
                        <div class="flex gap-2">
                            {#each Array(activeDeck.cols || 6) as _, c}
                                {@const label = getSeatLabel(
                                    activeDeck.id,
                                    r,
                                    c,
                                )}
                                {@const data = seatData[label] || {}}
                                <Seat
                                    id={label}
                                    label={`${String.fromCharCode(65 + r)}${c + 1}`}
                                    status={data.status || "available"}
                                    price={activeDeck.seatPrice || 500}
                                    category={activeDeck.type}
                                    on:click={handleSeatClick}
                                    className="w-9 h-9"
                                />
                            {/each}
                        </div>
                    {/each}
                </div>
            {/if}

            <!-- Water Effect Footer -->
            <div
                class="mt-6 h-4 bg-gradient-to-r from-blue-300 via-blue-400 to-blue-300 rounded-full opacity-50"
            ></div>

            <!-- Legend -->
            <div class="flex gap-4 mt-4 text-xs text-gray-600 justify-center">
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
