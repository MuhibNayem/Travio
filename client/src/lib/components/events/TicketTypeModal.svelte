<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Dialog from "$lib/components/ui/dialog";
    import { eventsApi, type TicketType, type Event } from "$lib/api/events";
    import { toast } from "svelte-sonner";
    import { page } from "$app/stores";
    import { Loader2, Plus, Ticket } from "@lucide/svelte";

    let {
        open = $bindable(false),
        event,
        onSuccess,
    } = $props<{
        open: boolean;
        event: Event | null;
        onSuccess: () => void;
    }>();

    let loading = false;
    let ticketTypes: TicketType[] = [];
    let loadingList = false;

    // Form State
    let name = $state("");
    let price = $state(0);
    let quantity = $state(100);
    let salesStart = $state("");
    let salesEnd = $state("");

    $effect(() => {
        if (open && event) {
            loadTickets();
        }
    });

    async function loadTickets() {
        if (!event) return;
        loadingList = true;
        try {
            ticketTypes = await eventsApi.getTicketTypes(event.id);
        } catch (e) {
            console.error(e);
        } finally {
            loadingList = false;
        }
    }

    async function handleSubmit() {
        if (!name || !price || !quantity || !event) return;

        loading = true;
        try {
            await eventsApi.createTicketType({
                event_id: event.id,
                organization_id: $page.data.user.organization_id,
                name,
                price_paisa: Number(price) * 100, // Convert to paisa
                total_quantity: Number(quantity),
                sales_start_time: salesStart
                    ? new Date(salesStart).toISOString()
                    : new Date().toISOString(),
                sales_end_time: salesEnd
                    ? new Date(salesEnd).toISOString()
                    : new Date(
                          new Date(event.start_time).getTime() - 3600000,
                      ).toISOString(), // Default 1hr before event
            });
            toast.success("Ticket type added");
            loadTickets();
            resetForm();
        } catch (e) {
            console.error(e);
            toast.error("Failed to add ticket type");
        } finally {
            loading = false;
        }
    }

    function resetForm() {
        name = "";
        price = 0;
        quantity = 100;
        salesStart = "";
        salesEnd = "";
    }
</script>

<Dialog.Root bind:open>
    <Dialog.Content class="sm:max-w-[600px]">
        <Dialog.Header>
            <Dialog.Title>Manage Tickets</Dialog.Title>
            <Dialog.Description>
                Configure ticket types for <span class="font-semibold"
                    >{event?.title}</span
                >
            </Dialog.Description>
        </Dialog.Header>

        <div class="space-y-6">
            <!-- Add New Ticket Form -->
            <div class="p-4 border rounded-lg bg-muted/30 space-y-4">
                <h4 class="font-medium text-sm flex items-center gap-2">
                    <Plus class="h-4 w-4" /> Add Ticket Type
                </h4>
                <div class="grid grid-cols-2 gap-4">
                    <div class="space-y-1">
                        <Label class="text-xs">Name</Label>
                        <Input
                            bind:value={name}
                            placeholder="e.g. VIP, Early Bird"
                            class="h-8"
                        />
                    </div>
                    <div class="space-y-1">
                        <Label class="text-xs">Price (BDT)</Label>
                        <Input type="number" bind:value={price} class="h-8" />
                    </div>
                    <div class="space-y-1">
                        <Label class="text-xs">Quantity</Label>
                        <Input
                            type="number"
                            bind:value={quantity}
                            class="h-8"
                        />
                    </div>
                    <div class="space-y-1 flex items-end">
                        <Button
                            size="sm"
                            class="w-full"
                            onclick={handleSubmit}
                            disabled={loading || !name}
                        >
                            {#if loading}
                                <Loader2 class="h-3 w-3 animate-spin mr-2" />
                            {/if}
                            Add Ticket
                        </Button>
                    </div>
                </div>
            </div>

            <!-- Existing Tickets List -->
            <div class="space-y-2">
                <Label>Existing Ticket Types</Label>
                <div
                    class="border rounded-md divide-y max-h-[200px] overflow-y-auto"
                >
                    {#if loadingList}
                        <div
                            class="p-4 text-center text-xs text-muted-foreground"
                        >
                            Loading...
                        </div>
                    {:else if ticketTypes.length === 0}
                        <div
                            class="p-4 text-center text-xs text-muted-foreground"
                        >
                            No tickets configured yet.
                        </div>
                    {:else}
                        {#each ticketTypes as t}
                            <div
                                class="flex items-center justify-between p-3 text-sm"
                            >
                                <div>
                                    <div class="font-medium">{t.name}</div>
                                    <div class="text-xs text-muted-foreground">
                                        {t.available_quantity} / {t.total_quantity}
                                        avail
                                    </div>
                                </div>
                                <div class="font-semibold">
                                    ৳{t.price_paisa / 100}
                                </div>
                            </div>
                        {/each}
                    {/if}
                </div>
            </div>
        </div>

        <Dialog.Footer>
            <Button variant="outline" onclick={() => (open = false)}
                >Close</Button
            >
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
