<script lang="ts">
    import { onMount } from "svelte";
    import { eventsApi, type Venue } from "$lib/api/events";
    import { Button } from "$lib/components/ui/button";
    import { Plus, MapPin, Users, Edit } from "@lucide/svelte";
    import VenueModal from "$lib/components/events/VenueModal.svelte";
    import * as Table from "$lib/components/ui/table";
    import { toast } from "svelte-sonner";
    import { page } from "$app/stores";

    let venues: Venue[] = [];
    let loading = true;
    let showModal = false;
    let selectedVenue: Venue | null = null;

    async function loadVenues() {
        loading = true;
        try {
            const orgId = $page.data.user.organization_id;
            venues = await eventsApi.getVenues(orgId);
        } catch (e) {
            console.error(e);
            toast.error("Failed to load venues");
        } finally {
            loading = false;
        }
    }

    function openCreate() {
        selectedVenue = null;
        showModal = true;
    }

    function openEdit(v: Venue) {
        selectedVenue = v;
        showModal = true;
    }

    onMount(() => {
        loadVenues();
    });
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Venues</h1>
            <p class="text-muted-foreground mt-2">
                Manage stadiums, auditoriums, and event grounds.
            </p>
        </div>
        <Button onclick={openCreate}>
            <Plus class="mr-2 h-4 w-4" />
            Add Venue
        </Button>
    </div>

    <div class="rounded-xl border bg-card shadow-sm">
        <Table.Root>
            <Table.Header>
                <Table.Row>
                    <Table.Head>NAME</Table.Head>
                    <Table.Head>LOCATION</Table.Head>
                    <Table.Head>CAPACITY</Table.Head>
                    <Table.Head>TYPE</Table.Head>
                    <Table.Head class="text-right">ACTIONS</Table.Head>
                </Table.Row>
            </Table.Header>
            <Table.Body>
                {#if loading}
                    <Table.Row>
                        <Table.Cell colspan={5} class="h-24 text-center"
                            >Loading venues...</Table.Cell
                        >
                    </Table.Row>
                {:else if venues.length === 0}
                    <Table.Row>
                        <Table.Cell colspan={5} class="h-64 text-center">
                            <div
                                class="flex flex-col items-center justify-center text-muted-foreground"
                            >
                                <MapPin class="h-8 w-8 mb-4 opacity-50" />
                                <p class="text-lg font-medium text-foreground">
                                    No venues found
                                </p>
                                <p class="text-sm">
                                    Add your first venue to start hosting
                                    events.
                                </p>
                            </div>
                        </Table.Cell>
                    </Table.Row>
                {:else}
                    {#each venues as venue}
                        <Table.Row>
                            <Table.Cell class="font-medium"
                                >{venue.name}</Table.Cell
                            >
                            <Table.Cell>
                                <div class="flex flex-col">
                                    <span>{venue.city}</span>
                                    <span class="text-xs text-muted-foreground"
                                        >{venue.address}</span
                                    >
                                </div>
                            </Table.Cell>
                            <Table.Cell>
                                <div class="flex items-center gap-2">
                                    <Users
                                        class="h-3 w-3 text-muted-foreground"
                                    />
                                    {venue.capacity ||
                                        venue.sections?.reduce(
                                            (acc, s) => acc + s.capacity,
                                            0,
                                        ) ||
                                        0}
                                </div>
                            </Table.Cell>
                            <Table.Cell>
                                <span class="capitalize"
                                    >{venue.type
                                        ?.toString()
                                        .replace("VENUE_TYPE_", "")
                                        .toLowerCase()
                                        .replace("_", " ") || "Unknown"}</span
                                >
                            </Table.Cell>
                            <Table.Cell class="text-right">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    onclick={() => openEdit(venue)}
                                >
                                    <Edit class="h-4 w-4" />
                                </Button>
                            </Table.Cell>
                        </Table.Row>
                    {/each}
                {/if}
            </Table.Body>
        </Table.Root>
    </div>
</div>

<VenueModal
    bind:open={showModal}
    venueToEdit={selectedVenue}
    onSuccess={loadVenues}
/>
