<script lang="ts">
    import { onMount } from "svelte";
    import { catalogApi, type Route } from "$lib/api/catalog";
    import { Button } from "$lib/components/ui/button";
    import { Plus, Map, ArrowRight } from "@lucide/svelte";
    import RouteModal from "$lib/components/operations/RouteModal.svelte";
    import * as Table from "$lib/components/ui/table";
    import { toast } from "svelte-sonner";

    let routes: Route[] = [];
    let loading = true;
    let showCreateModal = false;

    async function loadRoutes() {
        loading = true;
        try {
            routes = await catalogApi.getRoutes();
        } catch (e) {
            console.error(e);
            toast.error("Failed to load routes");
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        loadRoutes();
    });
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Routes</h1>
            <p class="text-muted-foreground mt-2">
                Manage your transport network and connections.
            </p>
        </div>
        <Button onclick={() => (showCreateModal = true)}>
            <Plus class="mr-2 h-4 w-4" />
            Create Route
        </Button>
    </div>

    <!-- Stats or Summary could go here -->

    <div class="rounded-xl border bg-card shadow-sm">
        <Table.Root>
            <Table.Header>
                <Table.Row>
                    <Table.Head>CODE</Table.Head>
                    <Table.Head>NAME</Table.Head>
                    <Table.Head>ORIGIN</Table.Head>
                    <Table.Head>DESTINATION</Table.Head>
                    <Table.Head>DISTANCE</Table.Head>
                    <Table.Head>DURATION</Table.Head>
                    <Table.Head class="text-right">STATUS</Table.Head>
                </Table.Row>
            </Table.Header>
            <Table.Body>
                {#if loading}
                    <Table.Row>
                        <Table.Cell colspan={7} class="h-24 text-center"
                            >Loading routes...</Table.Cell
                        >
                    </Table.Row>
                {:else if routes.length === 0}
                    <Table.Row>
                        <Table.Cell colspan={7} class="h-64 text-center">
                            <div
                                class="flex flex-col items-center justify-center text-muted-foreground"
                            >
                                <div class="p-4 rounded-full bg-muted mb-4">
                                    <Map class="h-8 w-8" />
                                </div>
                                <p class="text-lg font-medium text-foreground">
                                    No routes found
                                </p>
                                <p class="text-sm">
                                    Create your first route to get started.
                                </p>
                            </div>
                        </Table.Cell>
                    </Table.Row>
                {:else}
                    {#each routes as route}
                        <Table.Row>
                            <Table.Cell class="font-medium"
                                >{route.code}</Table.Cell
                            >
                            <Table.Cell>{route.name}</Table.Cell>
                            <Table.Cell class="text-muted-foreground"
                                >{route.origin_station_id}</Table.Cell
                            >
                            <Table.Cell class="text-muted-foreground"
                                >{route.destination_station_id}</Table.Cell
                            >
                            <Table.Cell>{route.distance_km} km</Table.Cell>
                            <Table.Cell
                                >{Math.floor(
                                    route.estimated_duration_minutes / 60,
                                )}h {route.estimated_duration_minutes %
                                    60}m</Table.Cell
                            >
                            <Table.Cell class="text-right">
                                <span
                                    class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
                                >
                                    Active
                                </span>
                            </Table.Cell>
                        </Table.Row>
                    {/each}
                {/if}
            </Table.Body>
        </Table.Root>
    </div>
</div>

<RouteModal bind:open={showCreateModal} onSuccess={loadRoutes} />
