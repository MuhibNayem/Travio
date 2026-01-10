<script lang="ts">
    import { onMount } from "svelte";
    import { catalogApi, type Trip } from "$lib/api/catalog";
    import { Button } from "$lib/components/ui/button";
    import { Plus, Calendar, Clock, Bus, Train, Ship } from "@lucide/svelte";
    import TripModal from "$lib/components/operations/TripModal.svelte";
    import * as Table from "$lib/components/ui/table";
    import { toast } from "svelte-sonner";

    let trips: Trip[] = [];
    let loading = true;
    let showCreateModal = false;

    async function loadTrips() {
        loading = true;
        try {
            trips = await catalogApi.getTrips();
        } catch (e) {
            console.error(e);
            toast.error("Failed to load trips");
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        loadTrips();
    });

    const icons = {
        bus: Bus,
        train: Train,
        launch: Ship,
    };

    function formatTime(ts: number) {
        return new Date(ts * 1000).toLocaleString();
    }
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Trip Schedule</h1>
            <p class="text-muted-foreground mt-2">
                Manage daily trip schedules and inventory.
            </p>
        </div>
        <Button onclick={() => (showCreateModal = true)}>
            <Plus class="mr-2 h-4 w-4" />
            Schedule Trip
        </Button>
    </div>

    <div class="rounded-xl border bg-card shadow-sm">
        <Table.Root>
            <Table.Header>
                <Table.Row>
                    <Table.Head>VEHICLE</Table.Head>
                    <Table.Head>DEPARTURE</Table.Head>
                    <Table.Head>CLASS</Table.Head>
                    <Table.Head>SEATS</Table.Head>
                    <Table.Head>PRICE</Table.Head>
                    <Table.Head class="text-right">STATUS</Table.Head>
                </Table.Row>
            </Table.Header>
            <Table.Body>
                {#if loading}
                    <Table.Row>
                        <Table.Cell colspan={6} class="h-24 text-center"
                            >Loading trips...</Table.Cell
                        >
                    </Table.Row>
                {:else if trips.length === 0}
                    <Table.Row>
                        <Table.Cell colspan={6} class="h-64 text-center">
                            <div
                                class="flex flex-col items-center justify-center text-muted-foreground"
                            >
                                <div class="p-4 rounded-full bg-muted mb-4">
                                    <Calendar class="h-8 w-8" />
                                </div>
                                <p class="text-lg font-medium text-foreground">
                                    No trips scheduled
                                </p>
                                <p class="text-sm">
                                    Schedule your first trip to start selling
                                    tickets.
                                </p>
                            </div>
                        </Table.Cell>
                    </Table.Row>
                {:else}
                    {#each trips as trip}
                        <Table.Row>
                            <Table.Cell>
                                <div class="flex items-center gap-3">
                                    <div
                                        class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary"
                                    >
                                        <svelte:component
                                            this={icons[
                                                trip.vehicle_type as keyof typeof icons
                                            ] || Bus}
                                            size={16}
                                        />
                                    </div>
                                    <span class="font-medium"
                                        >{trip.vehicle_id}</span
                                    >
                                </div>
                            </Table.Cell>
                            <Table.Cell>
                                <div class="flex flex-col">
                                    <span class="font-medium"
                                        >{new Date(
                                            trip.departure_time * 1000,
                                        ).toLocaleDateString()}</span
                                    >
                                    <span class="text-xs text-muted-foreground"
                                        >{new Date(
                                            trip.departure_time * 1000,
                                        ).toLocaleTimeString()}</span
                                    >
                                </div>
                            </Table.Cell>
                            <Table.Cell>{trip.vehicle_class}</Table.Cell>
                            <Table.Cell
                                >{trip.available_seats} / {trip.total_seats}</Table.Cell
                            >
                            <Table.Cell>
                                {trip.pricing?.base_price_paisa
                                    ? `৳${trip.pricing.base_price_paisa / 100}`
                                    : "-"}
                            </Table.Cell>
                            <Table.Cell class="text-right">
                                <span
                                    class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400"
                                >
                                    {trip.status.replace("TRIP_STATUS_", "")}
                                </span>
                            </Table.Cell>
                        </Table.Row>
                    {/each}
                {/if}
            </Table.Body>
        </Table.Root>
    </div>
</div>

<TripModal bind:open={showCreateModal} onSuccess={loadTrips} />
