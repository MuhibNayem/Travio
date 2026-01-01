<script lang="ts">
    import { cn } from "$lib/utils";
    import type { Seat, SeatStatus } from "$lib/types/transport";
    import * as Tooltip from "$lib/components/ui/tooltip";

    let { layout, onSelectionChange } = $props<{
        layout: Seat[][];
        onSelectionChange: (seats: Seat[]) => void;
    }>();

    let selectedSeats = $state<string[]>([]);

    function toggleSeat(seat: Seat) {
        if (seat.status === "booked" || seat.status === "held") return;

        if (selectedSeats.includes(seat.id)) {
            selectedSeats = selectedSeats.filter((id) => id !== seat.id);
        } else {
            // max 4 seats
            if (selectedSeats.length >= 4) {
                alert("You can only select up to 4 seats.");
                return;
            }
            selectedSeats = [...selectedSeats, seat.id];
        }

        // Notify parent
        // Flatten layout to find objects
        const flat = layout.flat();
        const selectedObjects = flat.filter((s) =>
            selectedSeats.includes(s.id),
        );
        onSelectionChange(selectedObjects);
    }

    function getSeatColor(status: SeatStatus, id: string) {
        if (status === "booked")
            return "bg-red-500/20 text-red-700 cursor-not-allowed border-red-200";
        if (status === "held")
            return "bg-yellow-500/20 text-yellow-700 cursor-not-allowed border-yellow-200";
        if (selectedSeats.includes(id))
            return "bg-primary text-primary-foreground shadow-lg scale-105 border-primary";
        return "bg-white dark:bg-white/5 hover:bg-blue-50 hover:border-blue-300 border-border cursor-pointer text-muted-foreground";
    }
</script>

<div
    class="flex flex-col items-center gap-8 rounded-3xl bg-white/50 p-8 shadow-inner backdrop-blur-xl dark:bg-black/20"
>
    <!-- Driver / Front -->
    <div class="flex w-full justify-between px-10 opacity-50">
        <div
            class="flex size-12 items-center justify-center rounded-full border border-black/10"
        >
            <span class="text-[10px] uppercase font-bold">Door</span>
        </div>
        <div
            class="flex size-12 items-center justify-center rounded-full border border-black/10 bg-black/5"
        >
            <span class="text-[10px] uppercase font-bold">Driver</span>
        </div>
    </div>

    <!-- Grid -->
    <div class="flex flex-col gap-4">
        {#each layout as row, i}
            <div class="flex items-center gap-8">
                <!-- Left Side (A, B) -->
                <div class="flex gap-4">
                    {#each row.slice(0, 2) as seat}
                        <button
                            class={cn(
                                "flex size-12 items-center justify-center rounded-xl border text-sm font-bold transition-all duration-200",
                                getSeatColor(seat.status, seat.id),
                            )}
                            onclick={() => toggleSeat(seat)}
                            disabled={seat.status === "booked" ||
                                seat.status === "held"}
                        >
                            {seat.label}
                        </button>
                    {/each}
                </div>

                <!-- Aisle -->
                <div class="w-8"></div>

                <!-- Right Side (C, D) -->
                <div class="flex gap-4">
                    {#each row.slice(2, 4) as seat}
                        <button
                            class={cn(
                                "flex size-12 items-center justify-center rounded-xl border text-sm font-bold transition-all duration-200",
                                getSeatColor(seat.status, seat.id),
                            )}
                            onclick={() => toggleSeat(seat)}
                            disabled={seat.status === "booked" ||
                                seat.status === "held"}
                        >
                            {seat.label}
                        </button>
                    {/each}
                </div>
            </div>
        {/each}
    </div>

    <!-- Legend -->
    <div class="flex gap-6 mt-4">
        <div class="flex items-center gap-2">
            <div class="size-4 rounded bg-white border border-border"></div>
            <span class="text-xs font-medium text-muted-foreground"
                >Available</span
            >
        </div>
        <div class="flex items-center gap-2">
            <div class="size-4 rounded bg-primary"></div>
            <span class="text-xs font-medium text-muted-foreground"
                >Selected</span
            >
        </div>
        <div class="flex items-center gap-2">
            <div
                class="size-4 rounded bg-red-500/20 border border-red-200"
            ></div>
            <span class="text-xs font-medium text-muted-foreground">Booked</span
            >
        </div>
    </div>
</div>
