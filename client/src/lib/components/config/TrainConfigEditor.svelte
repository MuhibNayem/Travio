<script lang="ts">
    import { Label } from "$lib/components/ui/label";
    import { Input } from "$lib/components/ui/input";
    import { Button } from "$lib/components/ui/button";
    import { Checkbox } from "$lib/components/ui/checkbox";
    import { Badge } from "$lib/components/ui/badge";
    import { Plus, Trash, Train } from "@lucide/svelte";
    import {
        type TrainConfig,
        type TrainCoach,
        TrainCoachClass,
        BerthType,
    } from "$lib/api/fleet";

    let {
        config = $bindable<TrainConfig>({
            coaches: [],
        }),
    } = $props<{ config: TrainConfig }>();

    // Ensure defaults
    $effect(() => {
        if (!config.coaches) config.coaches = [];
    });

    const coachClasses = [
        { value: TrainCoachClass.AC_FIRST, label: "AC First Class" },
        { value: TrainCoachClass.AC_SECOND, label: "AC Second Class" },
        { value: TrainCoachClass.AC_CHAIR, label: "AC Chair Car" },
        { value: TrainCoachClass.SLEEPER, label: "Sleeper" },
        { value: TrainCoachClass.S_CHAIR, label: "Shovan Chair" },
        { value: TrainCoachClass.GENERAL, label: "General" },
    ];

    const berthTypes = [
        { value: BerthType.CHAIR, label: "Chair Seating" },
        { value: BerthType.TWO_TIER, label: "2-Tier Berth" },
        { value: BerthType.THREE_TIER, label: "3-Tier Berth" },
    ];

    function addCoach() {
        const nextId = `C${(config.coaches?.length || 0) + 1}`;
        config.coaches = [
            ...(config.coaches || []),
            {
                id: nextId,
                name: `Coach ${nextId}`,
                class: TrainCoachClass.S_CHAIR,
                rows: 15,
                seats_per_row: 6,
                has_berths: false,
                berth_config: {
                    type: BerthType.CHAIR,
                    berths_per_compartment: 0,
                    has_side_berths: false,
                },
                price_paisa: 50000,
            },
        ];
    }

    function removeCoach(index: number) {
        config.coaches = (config.coaches || []).filter(
            (_: TrainCoach, i: number) => i !== index,
        );
    }

    function getClassColor(cls: TrainCoachClass): string {
        switch (cls) {
            case TrainCoachClass.AC_FIRST:
                return "bg-purple-100 text-purple-800";
            case TrainCoachClass.AC_SECOND:
                return "bg-blue-100 text-blue-800";
            case TrainCoachClass.AC_CHAIR:
                return "bg-sky-100 text-sky-800";
            case TrainCoachClass.SLEEPER:
                return "bg-amber-100 text-amber-800";
            case TrainCoachClass.S_CHAIR:
                return "bg-green-100 text-green-800";
            default:
                return "bg-gray-100 text-gray-800";
        }
    }

    // Total capacity
    let totalCapacity = $derived(
        (config.coaches || []).reduce(
            (sum: number, c: TrainCoach) => sum + c.rows * c.seats_per_row,
            0,
        ),
    );
</script>

<div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
            <Train class="h-5 w-5 text-primary" />
            <h4 class="font-semibold">Train Coach Configuration</h4>
            <Badge variant="secondary"
                >{config.coaches?.length || 0} coaches</Badge
            >
            <Badge variant="outline">{totalCapacity} seats</Badge>
        </div>
        <Button onclick={addCoach}>
            <Plus class="h-4 w-4 mr-1" /> Add Coach
        </Button>
    </div>

    <!-- Coach List -->
    {#if config.coaches && config.coaches.length > 0}
        <div class="space-y-4">
            {#each config.coaches as coach, i}
                <div class="rounded-lg border bg-card p-4 space-y-4">
                    <div class="flex items-center justify-between">
                        <div class="flex items-center gap-3">
                            <Input
                                class="w-20 font-mono"
                                bind:value={coach.id}
                                placeholder="S1"
                            />
                            <Input
                                class="w-48"
                                bind:value={coach.name}
                                placeholder="Coach Name"
                            />
                            <Badge class={getClassColor(coach.class)}>
                                {coachClasses.find(
                                    (c) => c.value === coach.class,
                                )?.label || "Unknown"}
                            </Badge>
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            class="text-destructive"
                            onclick={() => removeCoach(i)}
                        >
                            <Trash class="h-4 w-4" />
                        </Button>
                    </div>

                    <div class="grid grid-cols-2 md:grid-cols-5 gap-4">
                        <div class="space-y-1">
                            <Label class="text-xs">Class</Label>
                            <select
                                class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm"
                                bind:value={coach.class}
                            >
                                {#each coachClasses as cls}
                                    <option value={cls.value}
                                        >{cls.label}</option
                                    >
                                {/each}
                            </select>
                        </div>
                        <div class="space-y-1">
                            <Label class="text-xs">Rows</Label>
                            <Input
                                type="number"
                                bind:value={coach.rows}
                                min="1"
                                max="30"
                            />
                        </div>
                        <div class="space-y-1">
                            <Label class="text-xs">Seats/Row</Label>
                            <Input
                                type="number"
                                bind:value={coach.seats_per_row}
                                min="2"
                                max="8"
                            />
                        </div>
                        <div class="space-y-1">
                            <Label class="text-xs">Price (৳)</Label>
                            <Input
                                type="number"
                                bind:value={coach.price_paisa}
                            />
                        </div>
                        <div class="space-y-1">
                            <Label class="text-xs">Capacity</Label>
                            <div
                                class="h-10 flex items-center px-3 bg-muted rounded-md text-muted-foreground font-mono text-sm"
                            >
                                {coach.rows * coach.seats_per_row}
                            </div>
                        </div>
                    </div>

                    <!-- Berth Configuration -->
                    <div class="flex items-center gap-4 pt-2 border-t">
                        <label class="flex items-center gap-2 cursor-pointer">
                            <Checkbox bind:checked={coach.has_berths} />
                            <span class="text-sm">Has Berths</span>
                        </label>

                        {#if coach.has_berths}
                            <select
                                class="h-8 rounded-md border border-input bg-background px-2 text-sm"
                                bind:value={coach.berth_config!.type}
                            >
                                {#each berthTypes as bt}
                                    <option value={bt.value}>{bt.label}</option>
                                {/each}
                            </select>
                            <label
                                class="flex items-center gap-2 cursor-pointer"
                            >
                                <Checkbox
                                    bind:checked={
                                        coach.berth_config!.has_side_berths
                                    }
                                />
                                <span class="text-sm">Side Berths</span>
                            </label>
                        {/if}
                    </div>
                </div>
            {/each}
        </div>
    {:else}
        <div class="rounded-lg border border-dashed p-8 text-center">
            <Train class="h-8 w-8 mx-auto mb-3 text-muted-foreground" />
            <p class="text-muted-foreground">No coaches configured.</p>
            <p class="text-sm text-muted-foreground">
                Add coaches to define the train layout.
            </p>
        </div>
    {/if}
</div>
