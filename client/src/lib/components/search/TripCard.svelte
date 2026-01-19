<script lang="ts">
    import {
        MapPin,
        ArrowRight,
        Clock,
        Armchair,
        Bus,
        Train,
        Ship,
    } from "@lucide/svelte";
    import { Button } from "$lib/components/ui/button";
    import type { TripSearchResult } from "$lib/api/search";

    let { trip } = $props<{ trip: TripSearchResult }>();

    function formatTime(timestamp: number) {
        if (!timestamp) return "--:--";
        const date = new Date(timestamp * 1000);
        return date.toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    function formatPrice(paisa: number) {
        if (!paisa) return "0";
        return Math.floor(paisa / 100).toLocaleString();
    }

    const icons = {
        bus: Bus,
        train: Train,
        launch: Ship,
    } as const;

    const vehicleType = $derived(
        (trip.vehicle_type || "bus").toLowerCase() as keyof typeof icons,
    );

    const VehicleIcon = $derived(icons[vehicleType] || Bus);

    const orgParam = $derived(
        trip.organization_id ? `&org_id=${trip.organization_id}` : "",
    );
</script>

<div
    class="group relative overflow-hidden rounded-2xl border bg-white p-6 shadow-sm transition-all hover:shadow-md dark:bg-white/5 dark:border-white/10"
>
    <div
        class="flex flex-col gap-6 md:flex-row md:items-center md:justify-between"
    >
        <!-- Carrier Info -->
        <div class="flex items-start gap-4">
            <div
                class="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary dark:bg-primary/20"
            >
                <VehicleIcon size={24} />
            </div>
            <div>
                <h3 class="text-lg font-bold text-foreground">
                    {trip.operator_name || "Operator"}
                </h3>
                <p class="text-sm font-medium text-muted-foreground">
                    {trip.route_name || trip.vehicle_type || "Trip"}
                </p>
            </div>
        </div>

        <!-- Journey Info -->
        <div class="flex flex-1 items-center justify-center gap-6 px-4">
            <div class="text-center group-hover:text-primary transition-colors">
                <p class="text-lg font-bold text-foreground">
                    {formatTime(trip.departure_time)}
                </p>
                <div
                    class="flex items-center justify-center gap-1 text-xs font-semibold uppercase text-muted-foreground"
                >
                    <MapPin size={12} class="text-muted-foreground/70" />
                    {trip.from_station_name || "Origin"}
                </div>
            </div>

            <div class="flex flex-col items-center gap-1">
                <div class="relative flex w-24 items-center">
                    <div
                        class="h-[2px] w-full bg-border group-hover:bg-primary/50 transition-colors"
                    ></div>
                    <ArrowRight
                        class="absolute right-0 text-muted-foreground group-hover:text-primary transition-colors"
                        size={14}
                    />
                </div>
            </div>

            <div class="text-center group-hover:text-primary transition-colors">
                <p class="text-lg font-bold text-foreground">
                    {formatTime(trip.arrival_time)}
                </p>
                <div
                    class="flex items-center justify-center gap-1 text-xs font-semibold uppercase text-muted-foreground"
                >
                    <MapPin size={12} class="text-muted-foreground/70" />
                    {trip.to_station_name || "Destination"}
                </div>
            </div>
        </div>

        <!-- Action -->
        <div class="flex flex-col items-end gap-3 md:w-32">
            <div class="text-right">
                <p class="text-xl font-black text-primary">
                    ৳{formatPrice(trip.price_paisa)}
                </p>
                <div
                    class="flex items-center justify-end gap-1.5 text-xs font-medium text-muted-foreground"
                >
                    <Armchair size={14} class="text-emerald-500" />
                    <span>{trip.available_seats || 0} seats left</span>
                </div>
            </div>
            <Button
                href={`/booking/${trip.trip_id}?from=${trip.from_station_id}&to=${trip.to_station_id}${orgParam}`}
                class="w-full rounded-lg font-bold shadow-lg shadow-primary/20 hover:shadow-primary/30 transition-all"
                size="sm"
            >
                Book Seats
            </Button>
        </div>
    </div>
</div>
