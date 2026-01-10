<script lang="ts">
    import { Label } from "$lib/components/ui/label";
    import { Input } from "$lib/components/ui/input";
    import { Button } from "$lib/components/ui/button";
    import { Badge } from "$lib/components/ui/badge";
    import { Plus, Trash, Ship, Bed } from "@lucide/svelte";
    import {
        type LaunchConfig,
        type LaunchDeck,
        type LaunchCabin,
        LaunchDeckType,
    } from "$lib/api/fleet";

    let {
        config = $bindable<LaunchConfig>({
            decks: [],
        }),
    } = $props<{ config: LaunchConfig }>();

    // Ensure defaults
    $effect(() => {
        if (!config.decks) config.decks = [];
    });

    const deckTypes = [
        { value: LaunchDeckType.ECONOMY, label: "Economy" },
        { value: LaunchDeckType.BUSINESS, label: "Business" },
        { value: LaunchDeckType.VIP_CABIN, label: "VIP Cabin" },
    ];

    function addDeck() {
        const nextId = `D${(config.decks?.length || 0) + 1}`;
        config.decks = [
            ...(config.decks || []),
            {
                id: nextId,
                name: `Deck ${config.decks?.length + 1}`,
                type: LaunchDeckType.ECONOMY,
                rows: 8,
                cols: 6,
                seat_price_paisa: 50000,
                cabins: [],
            },
        ];
    }

    function removeDeck(index: number) {
        config.decks = (config.decks || []).filter(
            (_: LaunchDeck, i: number) => i !== index,
        );
    }

    function addCabin(deckIndex: number) {
        const deck = config.decks![deckIndex];
        if (!deck.cabins) deck.cabins = [];
        const nextId = `C${deck.cabins.length + 1}`;
        deck.cabins = [
            ...deck.cabins,
            {
                id: nextId,
                name: `Cabin ${deck.cabins.length + 1}`,
                beds: 2,
                price_paisa: 300000,
                is_suite: false,
            },
        ];
        config.decks = [...config.decks!];
    }

    function removeCabin(deckIndex: number, cabinIndex: number) {
        const deck = config.decks![deckIndex];
        deck.cabins = (deck.cabins || []).filter(
            (_: LaunchCabin, i: number) => i !== cabinIndex,
        );
        config.decks = [...config.decks!];
    }

    function getDeckTypeColor(type: LaunchDeckType): string {
        switch (type) {
            case LaunchDeckType.VIP_CABIN:
                return "bg-amber-100 text-amber-800";
            case LaunchDeckType.BUSINESS:
                return "bg-blue-100 text-blue-800";
            default:
                return "bg-green-100 text-green-800";
        }
    }

    // Total capacity
    let totalCapacity = $derived(
        (config.decks || []).reduce((sum: number, d: LaunchDeck) => {
            if (d.type === LaunchDeckType.VIP_CABIN) {
                return (
                    sum +
                    (d.cabins || []).reduce(
                        (cs: number, c: LaunchCabin) => cs + c.beds,
                        0,
                    )
                );
            }
            return sum + (d.rows || 0) * (d.cols || 0);
        }, 0),
    );
</script>

<div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
            <Ship class="h-5 w-5 text-primary" />
            <h4 class="font-semibold">Launch Deck Configuration</h4>
            <Badge variant="secondary">{config.decks?.length || 0} decks</Badge>
            <Badge variant="outline">{totalCapacity} capacity</Badge>
        </div>
        <Button onclick={addDeck}>
            <Plus class="h-4 w-4 mr-1" /> Add Deck
        </Button>
    </div>

    <!-- Deck List -->
    {#if config.decks && config.decks.length > 0}
        <div class="space-y-4">
            {#each config.decks as deck, i}
                <div
                    class="rounded-lg border bg-gradient-to-r from-blue-50/50 to-transparent p-4 space-y-4"
                >
                    <div class="flex items-center justify-between">
                        <div class="flex items-center gap-3">
                            <Input
                                class="w-16 font-mono"
                                bind:value={deck.id}
                                placeholder="D1"
                            />
                            <Input
                                class="w-40"
                                bind:value={deck.name}
                                placeholder="Deck Name"
                            />
                            <Badge class={getDeckTypeColor(deck.type)}>
                                {deckTypes.find((d) => d.value === deck.type)
                                    ?.label || "Unknown"}
                            </Badge>
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            class="text-destructive"
                            onclick={() => removeDeck(i)}
                        >
                            <Trash class="h-4 w-4" />
                        </Button>
                    </div>

                    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <div class="space-y-1">
                            <Label class="text-xs">Type</Label>
                            <select
                                class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm"
                                bind:value={deck.type}
                            >
                                {#each deckTypes as dt}
                                    <option value={dt.value}>{dt.label}</option>
                                {/each}
                            </select>
                        </div>

                        {#if deck.type !== LaunchDeckType.VIP_CABIN}
                            <div class="space-y-1">
                                <Label class="text-xs">Rows</Label>
                                <Input
                                    type="number"
                                    bind:value={deck.rows}
                                    min="1"
                                    max="20"
                                />
                            </div>
                            <div class="space-y-1">
                                <Label class="text-xs">Columns</Label>
                                <Input
                                    type="number"
                                    bind:value={deck.cols}
                                    min="1"
                                    max="10"
                                />
                            </div>
                            <div class="space-y-1">
                                <Label class="text-xs">Price (৳)</Label>
                                <Input
                                    type="number"
                                    bind:value={deck.seat_price_paisa}
                                />
                            </div>
                        {/if}
                    </div>

                    <!-- VIP Cabin Configuration -->
                    {#if deck.type === LaunchDeckType.VIP_CABIN}
                        <div class="border-t pt-4 mt-4 space-y-3">
                            <div class="flex items-center justify-between">
                                <div class="flex items-center gap-2">
                                    <Bed
                                        class="h-4 w-4 text-muted-foreground"
                                    />
                                    <span class="text-sm font-medium"
                                        >Cabins</span
                                    >
                                </div>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onclick={() => addCabin(i)}
                                >
                                    <Plus class="h-3 w-3 mr-1" /> Add Cabin
                                </Button>
                            </div>

                            {#if deck.cabins && deck.cabins.length > 0}
                                <div
                                    class="grid grid-cols-1 md:grid-cols-2 gap-3"
                                >
                                    {#each deck.cabins as cabin, ci}
                                        <div
                                            class="flex items-center gap-2 p-3 rounded-lg bg-white border"
                                        >
                                            <Input
                                                class="w-16"
                                                bind:value={cabin.id}
                                                placeholder="C1"
                                            />
                                            <Input
                                                class="flex-1"
                                                bind:value={cabin.name}
                                                placeholder="Cabin Name"
                                            />
                                            <div class="w-16">
                                                <Input
                                                    type="number"
                                                    bind:value={cabin.beds}
                                                    min="1"
                                                    max="6"
                                                    title="Beds"
                                                />
                                            </div>
                                            <div class="w-24">
                                                <Input
                                                    type="number"
                                                    bind:value={
                                                        cabin.price_paisa
                                                    }
                                                    title="Price"
                                                />
                                            </div>
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                class="text-destructive h-8 w-8"
                                                onclick={() =>
                                                    removeCabin(i, ci)}
                                            >
                                                <Trash class="h-3 w-3" />
                                            </Button>
                                        </div>
                                    {/each}
                                </div>
                            {:else}
                                <p
                                    class="text-sm text-muted-foreground text-center py-2"
                                >
                                    No cabins configured.
                                </p>
                            {/if}
                        </div>
                    {/if}
                </div>
            {/each}
        </div>
    {:else}
        <div class="rounded-lg border border-dashed p-8 text-center">
            <Ship class="h-8 w-8 mx-auto mb-3 text-muted-foreground" />
            <p class="text-muted-foreground">No decks configured.</p>
            <p class="text-sm text-muted-foreground">
                Add decks to define the launch layout.
            </p>
        </div>
    {/if}
</div>
