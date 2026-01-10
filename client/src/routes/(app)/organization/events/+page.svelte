<script lang="ts">
    import { onMount } from "svelte";
    import { eventsApi, type Event } from "$lib/api/events";
    import { Button } from "$lib/components/ui/button";
    import { Plus, Calendar, MapPin, Ticket, MoreHorizontal, CheckCircle } from "@lucide/svelte";
    import * as Table from "$lib/components/ui/table";
    import * as DropdownMenu from "$lib/components/ui/dropdown-menu";
    import { toast } from "svelte-sonner";
    import { page } from "$app/stores";
    import TicketTypeModal from "$lib/components/events/TicketTypeModal.svelte";

    let events: Event[] = [];
    let loading = true;
    
    // Ticket Modal State
    let showTicketModal = false;
    let selectedEvent: Event | null = null;

    async function loadEvents() {
        loading = true;
        try {
            const orgId = $page.data.user.organization_id;
            events = await eventsApi.getEvents(orgId);
        } catch (e) {
            console.error(e);
            toast.error("Failed to load events");
        } finally {
            loading = false;
        }
    }

    async function publishEvent(id: string) {
        try {
            await eventsApi.publishEvent(id);
            toast.success("Event published!");
            loadEvents();
        } catch (e) {
            toast.error("Failed to publish event");
        }
    }

    function openTickets(e: Event) {
        selectedEvent = e;
        showTicketModal = true;
    }

    onMount(() => {
        loadEvents();
    });
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Events</h1>
            <p class="text-muted-foreground mt-2">Manage your concerts, games, and shows.</p>
        </div>
        <div class="flex gap-2">
            <Button variant="outline" href="/organization/events/venues">
                <MapPin class="mr-2 h-4 w-4" />
                Manage Venues
            </Button>
            <Button href="/organization/events/create">
                <Plus class="mr-2 h-4 w-4" />
                Create Event
            </Button>
        </div>
    </div>

     <div class="rounded-xl border bg-card shadow-sm">
        <Table.Root>
            <Table.Header>
                <Table.Row>
                    <Table.Head>EVENT</Table.Head>
                    <Table.Head>SCHEDULE</Table.Head>
                    <Table.Head>STATUS</Table.Head>
                    <Table.Head>CATEGORY</Table.Head>
                    <Table.Head class="text-right">ACTIONS</Table.Head>
                </Table.Row>
            </Table.Header>
            <Table.Body>
                {#if loading}
                     <Table.Row>
                        <Table.Cell colspan={5} class="h-24 text-center">Loading events...</Table.Cell>
                    </Table.Row>
                {:else if events.length === 0}
                    <Table.Row>
                        <Table.Cell colspan={5} class="h-64 text-center">
                            <div class="flex flex-col items-center justify-center text-muted-foreground">
                                <Calendar class="h-8 w-8 mb-4 opacity-50" />
                                <p class="text-lg font-medium text-foreground">No events found</p>
                                <p class="text-sm">Create your first event to start selling tickets.</p>
                            </div>
                        </Table.Cell>
                    </Table.Row>
                {:else}
                    {#each events as event}
                        <Table.Row>
                            <Table.Cell>
                                <div class="font-medium">{event.title}</div>
                                <div class="text-xs text-muted-foreground truncate max-w-[200px]">{event.description}</div>
                            </Table.Cell>
                            <Table.Cell>
                                <div class="flex flex-col">
                                    <span>{new Date(event.start_time).toLocaleDateString()}</span>
                                    <span class="text-xs text-muted-foreground">{new Date(event.start_time).toLocaleTimeString()}</span>
                                </div>
                            </Table.Cell>
                             <Table.Cell>
                                <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold
                                    {event.status.toString().includes('PUBLISHED') ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'}">
                                    {event.status.toString().replace("EVENT_STATUS_", "")}
                                </span>
                            </Table.Cell>
                            <Table.Cell>{event.category}</Table.Cell>
                            <Table.Cell class="text-right">
                                <DropdownMenu.Root>
                                    <DropdownMenu.Trigger asChild let:builder>
                                        <Buttonbuilders={[builder]} variant="ghost" size="icon">
                                            <MoreHorizontal class="h-4 w-4" />
                                        </Buttonbuilders>
                                    </DropdownMenu.Trigger>
                                    <DropdownMenu.Content align="end">
                                        <DropdownMenu.Label>Actions</DropdownMenu.Label>
                                        <DropdownMenu.Item onclick={() => openTickets(event)}>
                                            <Ticket class="mr-2 h-4 w-4" /> Manage Tickets
                                        </DropdownMenu.Item>
                                        {#if !event.status.toString().includes('PUBLISHED')}
                                            <DropdownMenu.Item onclick={() => publishEvent(event.id)}>
                                                <CheckCircle class="mr-2 h-4 w-4" /> Publish
                                            </DropdownMenu.Item>
                                        {/if}
                                        <!-- Edit Action (Future: Link to edit page) -->
                                    </DropdownMenu.Content>
                                </DropdownMenu.Root>
                            </Table.Cell>
                        </Table.Row>
                    {/each}
                {/if}
            </Table.Body>
        </Table.Root>
    </div>
</div>

<TicketTypeModal
    bind:open={showTicketModal}
    event={selectedEvent}
    onSuccess={() => {}}
/>
