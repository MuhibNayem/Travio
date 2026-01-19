<script lang="ts">
    import { onMount } from "svelte";
    import {
        catalogApi,
        type Schedule,
        type TripInstanceResult,
        type Route,
    } from "$lib/api/catalog";
    import { fleetApi, type Asset } from "$lib/api/fleet";
    import { Button } from "$lib/components/ui/button";
    import {
        Plus,
        Calendar,
        Clock,
        Bus,
        Train,
        Ship,
        RefreshCw,
    } from "@lucide/svelte";
    import TripModal from "$lib/components/operations/TripModal.svelte";
    import * as Table from "$lib/components/ui/table";
    import { toast } from "svelte-sonner";

    let schedules: Schedule[] = [];
    let tripInstances: TripInstanceResult[] = [];
    let routes: Route[] = [];
    let assets: Asset[] = [];
    let loading = true;
    let showCreateModal = false;
    let regeneratingId: string | null = null;

    const icons = {
        bus: Bus,
        train: Train,
        launch: Ship,
        ferry: Ship,
    };

    function minutesToTime(minutes: number) {
        const h = Math.floor(minutes / 60)
            .toString()
            .padStart(2, "0");
        const m = (minutes % 60).toString().padStart(2, "0");
        return `${h}:${m}`;
    }

    function daysMaskToText(mask: number) {
        const days = [
            { label: "Mon", bit: 1 },
            { label: "Tue", bit: 2 },
            { label: "Wed", bit: 4 },
            { label: "Thu", bit: 8 },
            { label: "Fri", bit: 16 },
            { label: "Sat", bit: 32 },
            { label: "Sun", bit: 64 },
        ];
        return days
            .filter((d) => (mask & d.bit) !== 0)
            .map((d) => d.label)
            .join(", ");
    }

    function getRouteName(routeId: string) {
        return routes.find((r) => r.id === routeId)?.name || routeId;
    }

    function formatDateRange(days: number) {
        const start = new Date();
        const end = new Date();
        end.setDate(start.getDate() + days);
        return {
            start: start.toISOString().slice(0, 10),
            end: end.toISOString().slice(0, 10),
        };
    }

    async function loadData() {
        loading = true;
        try {
            const [routesRes, schedulesRes, assetsRes] = await Promise.all([
                catalogApi.getRoutes(),
                catalogApi.listSchedules(),
                fleetApi.getAssets(),
            ]);
            routes = routesRes;
            schedules = schedulesRes;
            assets = assetsRes;

            const range = formatDateRange(7);
            tripInstances = await catalogApi.listTripInstances({
                start_date: range.start,
                end_date: range.end,
            });
        } catch (e) {
            console.error(e);
            toast.error("Failed to load schedules");
        } finally {
            loading = false;
        }
    }

    async function handleRegenerate(scheduleId: string) {
        regeneratingId = scheduleId;
        try {
            const range = formatDateRange(30); // Generate for next 30 days
            await catalogApi.generateTripInstances(
                scheduleId,
                range.start,
                range.end,
            );
            toast.success("Trip instances regenerated successfully");
            await loadData();
        } catch (e) {
            console.error(e);
            toast.error("Failed to generate trips");
        } finally {
            regeneratingId = null;
        }
    }

    function getVehicleName(id: string) {
        const asset = assets.find((a) => a.id === id);
        if (asset) return `${asset.name} (${asset.license_plate})`;
        return id;
    }

    onMount(() => {
        loadData();
    });
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Trip Schedules</h1>
            <p class="text-muted-foreground mt-2">
                Create recurring schedules and review upcoming trips.
            </p>
        </div>
        <Button onclick={() => (showCreateModal = true)}>
            <Plus class="mr-2 h-4 w-4" />
            Create Schedule
        </Button>
    </div>

    <div class="space-y-4">
        <div class="flex items-center justify-between">
            <h2 class="text-xl font-semibold tracking-tight">
                Recurring Schedules (Templates)
            </h2>
            <p class="text-sm text-muted-foreground">
                Base templates that generate actual daily trips.
            </p>
        </div>
        <div class="rounded-xl border bg-card shadow-sm">
            <Table.Root>
                <Table.Header>
                    <Table.Row>
                        <Table.Head>ROUTE</Table.Head>
                        <Table.Head>DEPARTURE</Table.Head>
                        <Table.Head>DAYS</Table.Head>
                        <Table.Head>SEATS</Table.Head>
                        <Table.Head>STATUS</Table.Head>
                        <Table.Head class="text-right">ACTIONS</Table.Head>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {#if loading}
                        <Table.Row>
                            <Table.Cell colspan={6} class="h-24 text-center">
                                Loading schedules...
                            </Table.Cell>
                        </Table.Row>
                    {:else if schedules.length === 0}
                        <Table.Row>
                            <Table.Cell colspan={6} class="h-40 text-center">
                                <div
                                    class="flex flex-col items-center justify-center text-muted-foreground"
                                >
                                    <div class="p-4 rounded-full bg-muted mb-4">
                                        <Calendar class="h-8 w-8" />
                                    </div>
                                    <p
                                        class="text-lg font-medium text-foreground"
                                    >
                                        No schedules created
                                    </p>
                                    <p class="text-sm">
                                        Create a schedule to generate trip
                                        instances.
                                    </p>
                                </div>
                            </Table.Cell>
                        </Table.Row>
                    {:else}
                        {#each schedules as schedule}
                            <Table.Row>
                                <Table.Cell
                                    >{getRouteName(
                                        schedule.route_id,
                                    )}</Table.Cell
                                >
                                <Table.Cell>
                                    <div class="flex items-center gap-2">
                                        <Clock
                                            class="h-4 w-4 text-muted-foreground"
                                        />
                                        {minutesToTime(
                                            schedule.departure_minutes,
                                        )}
                                    </div>
                                </Table.Cell>
                                <Table.Cell
                                    >{daysMaskToText(
                                        schedule.days_of_week,
                                    )}</Table.Cell
                                >
                                <Table.Cell>{schedule.total_seats}</Table.Cell>
                                <Table.Cell>
                                    <span
                                        class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400"
                                    >
                                        {schedule.status}
                                    </span>
                                </Table.Cell>
                                <Table.Cell class="text-right">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        disabled={regeneratingId ===
                                            schedule.id}
                                        onclick={() =>
                                            handleRegenerate(schedule.id)}
                                    >
                                        <RefreshCw
                                            class="mr-2 h-3.5 w-3.5 {regeneratingId ===
                                            schedule.id
                                                ? 'animate-spin'
                                                : ''}"
                                        />
                                        Regenerate
                                    </Button>
                                </Table.Cell>
                            </Table.Row>
                        {/each}
                    {/if}
                </Table.Body>
            </Table.Root>
        </div>
    </div>

    <div class="space-y-4">
        <div class="flex items-center justify-between">
            <h2 class="text-xl font-semibold tracking-tight">
                Upcoming Daily Trips (Live)
            </h2>
            <p class="text-sm text-muted-foreground">
                Actual bookable trips generated for the next 7 days.
            </p>
        </div>
        <div class="rounded-xl border bg-card shadow-sm">
            <Table.Root>
                <Table.Header>
                    <Table.Row>
                        <Table.Head>VEHICLE</Table.Head>
                        <Table.Head>ROUTE</Table.Head>
                        <Table.Head>DEPARTURE</Table.Head>
                        <Table.Head>SEATS</Table.Head>
                        <Table.Head class="text-right">STATUS</Table.Head>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {#if loading}
                        <Table.Row>
                            <Table.Cell colspan={5} class="h-24 text-center">
                                Loading trips...
                            </Table.Cell>
                        </Table.Row>
                    {:else if tripInstances.length === 0}
                        <Table.Row>
                            <Table.Cell colspan={5} class="h-40 text-center">
                                <div
                                    class="flex flex-col items-center justify-center text-muted-foreground"
                                >
                                    <div class="p-4 rounded-full bg-muted mb-4">
                                        <Calendar class="h-8 w-8" />
                                    </div>
                                    <p
                                        class="text-lg font-medium text-foreground"
                                    >
                                        No upcoming trips
                                    </p>
                                    <p class="text-sm">
                                        Generate trip instances from schedules.
                                    </p>
                                </div>
                            </Table.Cell>
                        </Table.Row>
                    {:else}
                        {#each tripInstances as result}
                            {#if result.trip}
                                <Table.Row>
                                    <Table.Cell>
                                        <div class="flex items-center gap-3">
                                            <div
                                                class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary"
                                            >
                                                <svelte:component
                                                    this={icons[
                                                        result.trip
                                                            .vehicle_type as keyof typeof icons
                                                    ] || Bus}
                                                    size={16}
                                                />
                                            </div>
                                            <span class="font-medium">
                                                {getVehicleName(
                                                    result.trip.vehicle_id,
                                                )}
                                            </span>
                                        </div>
                                    </Table.Cell>
                                    <Table.Cell
                                        >{result.route?.name || "-"}</Table.Cell
                                    >
                                    <Table.Cell>
                                        <div class="flex flex-col">
                                            <span class="font-medium">
                                                {new Date(
                                                    result.trip.departure_time *
                                                        1000,
                                                ).toLocaleDateString()}
                                            </span>
                                            <span
                                                class="text-xs text-muted-foreground"
                                            >
                                                {new Date(
                                                    result.trip.departure_time *
                                                        1000,
                                                ).toLocaleTimeString()}
                                            </span>
                                        </div>
                                    </Table.Cell>
                                    <Table.Cell>
                                        {result.trip.available_seats} / {result
                                            .trip.total_seats}
                                    </Table.Cell>
                                    <Table.Cell class="text-right">
                                        <span
                                            class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400"
                                        >
                                            {result.trip.status}
                                        </span>
                                    </Table.Cell>
                                </Table.Row>
                            {/if}
                        {/each}
                    {/if}
                </Table.Body>
            </Table.Root>
        </div>
    </div>
</div>

<TripModal bind:open={showCreateModal} onSuccess={loadData} />
