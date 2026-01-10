<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Badge } from "$lib/components/ui/badge";
    import type { Trip, TransportType } from "$lib/types/transport";
    import {
        ArrowRight,
        Clock,
        Armchair,
        Bus,
        Train,
        Ship,
    } from "@lucide/svelte";

    let { trip, fromId, toId } = $props<{
        trip: Trip;
        fromId: string;
        toId: string;
    }>();

    const icons: Record<TransportType, typeof Bus> = {
        bus: Bus,
        train: Train,
        launch: Ship,
    };

    const Icon = $derived(icons[trip.type as TransportType] || Bus);

    function formatTime(iso: string) {
        return new Date(iso).toLocaleTimeString("en-US", {
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    function getDuration(start: string, end: string) {
        const diff = new Date(end).getTime() - new Date(start).getTime();
        const hrs = Math.floor(diff / (1000 * 60 * 60));
        const mins = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
        return `${hrs}h ${mins}m`;
    }
</script>

<div
    class="group relative overflow-hidden rounded-2xl border border-white/10 bg-white/70 shadow-lg backdrop-blur-xl transition-all hover:scale-[1.01] hover:shadow-2xl dark:bg-[#1a2333]/70"
>
    <div class="flex flex-col gap-6 p-6 md:flex-row md:items-center">
        <!-- Carrier Info -->
        <div class="flex items-center gap-4 md:w-1/4">
            <div
                class="flex size-14 items-center justify-center rounded-xl bg-gradient-to-br from-blue-100 to-indigo-100 text-blue-600 shadow-sm dark:from-blue-900/40 dark:to-indigo-900/40 dark:text-blue-400"
            >
                <Icon size={28} />
            </div>
            <div>
                <h4 class="font-bold text-foreground">{trip.operator}</h4>
                <p class="text-sm text-muted-foreground">{trip.vehicleName}</p>
                <Badge variant="secondary" class="mt-1 text-xs">
                    {trip.class}
                </Badge>
            </div>
        </div>

        <!-- Schedule -->
        <div class="flex flex-1 items-center justify-between gap-4 md:px-8">
            <div class="text-center">
                <p class="text-2xl font-black text-foreground">
                    {formatTime(trip.departureTime)}
                </p>
                <p
                    class="text-xs uppercase tracking-wide text-muted-foreground"
                >
                    Departure
                </p>
            </div>

            <div class="flex flex-col items-center gap-1">
                <span class="text-xs font-medium text-muted-foreground"
                    >{getDuration(trip.departureTime, trip.arrivalTime)}</span
                >
                <div class="relative h-[2px] w-24 bg-border">
                    <div
                        class="absolute right-0 top-1/2 h-2 w-2 -translate-y-1/2 rounded-full bg-border"
                    ></div>
                    <div
                        class="absolute left-0 top-1/2 h-2 w-2 -translate-y-1/2 rounded-full bg-border"
                    ></div>
                </div>
                <div
                    class="flex items-center gap-1 text-xs text-muted-foreground"
                >
                    <Clock size={12} /> Direct
                </div>
            </div>

            <div class="text-center">
                <p class="text-2xl font-black text-foreground">
                    {formatTime(trip.arrivalTime)}
                </p>
                <p
                    class="text-xs uppercase tracking-wide text-muted-foreground"
                >
                    Arrival
                </p>
            </div>
        </div>

        <!-- Price & Action -->
        <div
            class="flex items-center justify-between border-t border-black/5 pt-4 md:w-1/4 md:flex-col md:border-l md:border-t-0 md:pl-6 md:pt-0 dark:border-white/5"
        >
            <div class="text-right md:mb-3 md:text-center">
                <p
                    class="text-xs font-bold uppercase tracking-wider text-muted-foreground"
                >
                    Starting From
                </p>
                <p class="text-2xl font-black text-primary">
                    ৳{trip.price}
                </p>
            </div>
            <Button
                size="lg"
                class="w-full gap-2 rounded-xl font-bold shadow-lg shadow-blue-500/20"
                href={`/booking/${trip.id}?from=${fromId}&to=${toId}`}
            >
                Book Seats <ArrowRight size={16} />
            </Button>
            <div
                class="hidden items-center justify-center gap-1.5 text-xs font-medium text-green-600 md:flex"
            >
                <Armchair size={14} />
                {trip.availableSeats} seats left
            </div>
        </div>
    </div>
</div>
