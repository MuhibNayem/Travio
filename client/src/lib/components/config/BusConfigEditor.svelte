<script lang="ts">
    import { Label } from "$lib/components/ui/label";
    import { Input } from "$lib/components/ui/input";
    import { Button } from "$lib/components/ui/button";
    import { Checkbox } from "$lib/components/ui/checkbox";
    import { Plus, Trash } from "@lucide/svelte";
    import type { BusConfig, SeatCategory } from "$lib/api/fleet";

    let {
        config = $bindable<BusConfig>({
            rows: 10,
            seats_per_row: 4,
            aisle_after_seat: 2,
            has_toilet: false,
            has_sleeper: false,
            categories: [],
        }),
    } = $props<{ config: BusConfig }>();

    // Ensure defaults
    $effect(() => {
        if (!config.rows) config.rows = 10;
        if (!config.seats_per_row) config.seats_per_row = 4;
        if (!config.aisle_after_seat) config.aisle_after_seat = 2;
        if (!config.categories) config.categories = [];
    });

    function addCategory() {
        config.categories = [
            ...(config.categories || []),
            { name: "", price_paisa: 0, seat_ids: [] },
        ];
    }

    function removeCategory(index: number) {
        config.categories = (config.categories || []).filter(
            (_: SeatCategory, i: number) => i !== index,
        );
    }

    // Preview grid
    let totalSeats = $derived(config.rows * config.seats_per_row);
</script>

<div class="space-y-6">
    <div class="rounded-lg border bg-muted/30 p-4">
        <h4 class="font-semibold mb-4">Bus Seat Layout</h4>

        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="space-y-2">
                <Label>Rows</Label>
                <Input
                    type="number"
                    bind:value={config.rows}
                    min="1"
                    max="20"
                />
            </div>
            <div class="space-y-2">
                <Label>Seats per Row</Label>
                <Input
                    type="number"
                    bind:value={config.seats_per_row}
                    min="2"
                    max="6"
                />
            </div>
            <div class="space-y-2">
                <Label>Aisle After Seat</Label>
                <Input
                    type="number"
                    bind:value={config.aisle_after_seat}
                    min="1"
                    max={config.seats_per_row - 1}
                />
            </div>
            <div class="space-y-2">
                <Label>Total Seats</Label>
                <div
                    class="h-10 flex items-center px-3 bg-muted rounded-md text-muted-foreground font-mono"
                >
                    {totalSeats}
                </div>
            </div>
        </div>

        <div class="flex gap-6 mt-4">
            <label class="flex items-center gap-2 cursor-pointer">
                <Checkbox bind:checked={config.has_toilet} />
                <span class="text-sm">Has Toilet</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
                <Checkbox bind:checked={config.has_sleeper} />
                <span class="text-sm">Sleeper Bus</span>
            </label>
        </div>
    </div>

    <!-- Live Preview -->
    <div class="rounded-lg border bg-gray-50 p-4">
        <h4 class="font-semibold mb-4 text-sm text-muted-foreground">
            Layout Preview
        </h4>
        <div class="flex flex-col gap-1 items-center max-w-fit mx-auto">
            {#each Array(Math.min(config.rows, 5)) as _, r}
                <div class="flex gap-1">
                    {#each Array(config.seats_per_row) as _, c}
                        {#if c === config.aisle_after_seat}
                            <div class="w-4"></div>
                        {/if}
                        <div
                            class="w-6 h-6 rounded border bg-white flex items-center justify-center text-[8px] text-gray-400"
                        >
                            {String.fromCharCode(65 + r)}{c + 1}
                        </div>
                    {/each}
                </div>
            {/each}
            {#if config.rows > 5}
                <div class="text-xs text-muted-foreground">
                    ... {config.rows - 5} more rows
                </div>
            {/if}
        </div>
    </div>

    <!-- Seat Categories -->
    <div class="rounded-lg border p-4 space-y-4">
        <div class="flex items-center justify-between">
            <h4 class="font-semibold">Seat Categories (Optional)</h4>
            <Button variant="ghost" size="sm" onclick={addCategory}>
                <Plus class="h-4 w-4 mr-1" /> Add Category
            </Button>
        </div>

        {#if config.categories && config.categories.length > 0}
            <div class="space-y-3">
                {#each config.categories as cat, i}
                    <div class="flex gap-3 items-end">
                        <div class="flex-1 space-y-1">
                            <Label class="text-xs">Name</Label>
                            <Input bind:value={cat.name} placeholder="VIP" />
                        </div>
                        <div class="w-32 space-y-1">
                            <Label class="text-xs">Price (৳)</Label>
                            <Input
                                type="number"
                                bind:value={cat.price_paisa}
                                placeholder="500"
                            />
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            class="text-destructive"
                            onclick={() => removeCategory(i)}
                        >
                            <Trash class="h-4 w-4" />
                        </Button>
                    </div>
                {/each}
            </div>
        {:else}
            <p class="text-sm text-muted-foreground text-center py-4">
                No categories defined. All seats will use default pricing.
            </p>
        {/if}
    </div>
</div>
