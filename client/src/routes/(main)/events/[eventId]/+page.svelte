<script lang="ts">
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";
    import {
        Calendar,
        MapPin,
        Clock,
        Ticket,
        Users,
        Share2,
        Loader2,
        Check,
        Armchair,
    } from "@lucide/svelte";
    import { toast } from "svelte-sonner";

    let eventId = $derived($page.params.eventId);
    let event = $state<any>(null);
    let isLoading = $state(true);
    let selectedTickets = $state<Record<string, number>>({});

    async function fetchEvent() {
        try {
            // Mock event data - in production, call events API
            event = {
                id: eventId,
                title: "Live Concert: Artist Name World Tour 2026",
                description: "Experience an unforgettable night of music with performances from top artists. This spectacular event features multiple stages, food stalls, and amazing performances.",
                category: "Concert",
                start_time: "2026-05-15T18:00:00Z",
                end_time: "2026-05-15T23:00:00Z",
                status: "published",
                images: [
                    "https://images.unsplash.com/photo-1476514525535-07fb3b4ae5f1?w=1200",
                ],
                venue: {
                    name: "Bangladesh Army Stadium",
                    address: "Matikata, Dhaka-1212",
                    city: "Dhaka",
                    capacity: 5000,
                },
                ticket_types: [
                    {
                        id: "VIP-001",
                        name: "VIP Front Row",
                        description: "Premium seating with backstage access",
                        price_paisa: 1500000,
                        available_quantity: 50,
                        total_quantity: 100,
                    },
                    {
                        id: "GA-001",
                        name: "General Admission",
                        description: "Standard entry to the concert",
                        price_paisa: 500000,
                        available_quantity: 1500,
                        total_quantity: 3000,
                    },
                    {
                        id: "EB-001",
                        name: "Early Bird",
                        description: "Limited discounted tickets",
                        price_paisa: 350000,
                        available_quantity: 0,
                        total_quantity: 500,
                    },
                ],
            };
        } catch (error) {
            toast.error("Failed to load event details");
        } finally {
            isLoading = false;
        }
    }

    function incrementTicket(typeId: string, max: number) {
        const current = selectedTickets[typeId] || 0;
        if (current < max && current < 5) {
            selectedTickets = { ...selectedTickets, [typeId]: current + 1 };
        }
    }

    function decrementTicket(typeId: string) {
        const current = selectedTickets[typeId] || 0;
        if (current > 0) {
            selectedTickets = { ...selectedTickets, [typeId]: current - 1 };
        }
    }

    $: totalTickets = Object.values(selectedTickets).reduce((a, b) => a + b, 0);
    $: totalPrice = event?.ticket_types.reduce((sum: number, t: any) => {
        return sum + (t.price_paisa * (selectedTickets[t.id] || 0));
    }, 0) || 0;

    async function proceedToCheckout() {
        if (totalTickets === 0) {
            toast.error("Select at least one ticket");
            return;
        }
        toast.success("Tickets added to cart", {
            description: "Redirecting to checkout...",
        });
        goto(`/checkout/events/${eventId}`);
    }

    $effect(() => {
        if (eventId) fetchEvent();
    });
</script>

<div class="min-h-screen bg-muted/30 pb-32 pt-20">
    {#if isLoading}
        <div class="flex h-[50vh] flex-col items-center justify-center gap-4">
            <Loader2 class="animate-spin text-primary" size={48} />
            <p class="text-muted-foreground">Loading event details...</p>
        </div>
    {:else if event}
        <div class="container mx-auto max-w-6xl px-4">
            <!-- Event Banner -->
            <div class="mb-8 overflow-hidden rounded-2xl">
                <img
                    src={event.images[0]}
                    alt={event.title}
                    class="h-64 w-full object-cover md:h-96"
                />
            </div>

            <div class="grid gap-8 lg:grid-cols-3">
                <!-- Left: Event Details -->
                <div class="lg:col-span-2">
                    <!-- Event Info -->
                    <div class="glass-card mb-6 rounded-xl p-6">
                        <div class="mb-4 flex items-center gap-2">
                            <span class="rounded-full bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
                                {event.category}
                            </span>
                            <span class="rounded-full bg-green-100 px-3 py-1 text-sm font-medium text-green-800 dark:bg-green-900/30 dark:text-green-400">
                                On Sale
                            </span>
                        </div>

                        <h1 class="mb-4 text-3xl font-black">{event.title}</h1>

                        <div class="mb-6 grid gap-4 sm:grid-cols-2">
                            <div class="flex items-center gap-3">
                                <Calendar size={20} class="text-primary" />
                                <div>
                                    <p class="text-sm text-muted-foreground">Date</p>
                                    <p class="font-medium">
                                        {new Date(event.start_time).toLocaleDateString("en-US", {
                                            weekday: "long",
                                            year: "numeric",
                                            month: "long",
                                            day: "numeric",
                                        })}
                                    </p>
                                </div>
                            </div>
                            <div class="flex items-center gap-3">
                                <Clock size={20} class="text-primary" />
                                <div>
                                    <p class="text-sm text-muted-foreground">Time</p>
                                    <p class="font-medium">
                                        {new Date(event.start_time).toLocaleTimeString([], {
                                            hour: "2-digit",
                                            minute: "2-digit",
                                        })} - {new Date(event.end_time).toLocaleTimeString([], {
                                            hour: "2-digit",
                                            minute: "2-digit",
                                        })}
                                    </p>
                                </div>
                            </div>
                            <div class="flex items-center gap-3 sm:col-span-2">
                                <MapPin size={20} class="text-primary" />
                                <div>
                                    <p class="text-sm text-muted-foreground">Venue</p>
                                    <p class="font-medium">{event.venue.name}</p>
                                    <p class="text-sm text-muted-foreground">{event.venue.address}</p>
                                </div>
                            </div>
                        </div>

                        <Separator class="my-6" />

                        <div class="prose max-w-none">
                            <h3 class="mb-2 text-lg font-bold">About This Event</h3>
                            <p class="text-muted-foreground">{event.description}</p>
                        </div>
                    </div>

                    <!-- Venue Info -->
                    <div class="glass-card rounded-xl p-6">
                        <h3 class="mb-4 text-lg font-bold">Venue Information</h3>
                        <div class="grid gap-4 sm:grid-cols-2">
                            <div class="rounded-lg bg-muted/50 p-4">
                                <div class="flex items-center gap-2">
                                    <Users size={18} class="text-primary" />
                                    <span class="font-medium">Capacity</span>
                                </div>
                                <p class="mt-1 text-2xl font-bold">{event.venue.capacity.toLocaleString()}</p>
                            </div>
                            <div class="rounded-lg bg-muted/50 p-4">
                                <div class="flex items-center gap-2">
                                    <MapPin size={18} class="text-primary" />
                                    <span class="font-medium">Location</span>
                                </div>
                                <p class="mt-1 text-lg font-medium">{event.venue.city}</p>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Right: Ticket Selection -->
                <div>
                    <div class="glass-card sticky top-24 rounded-xl p-6">
                        <h3 class="mb-4 flex items-center gap-2 text-lg font-bold">
                            <Ticket size={20} class="text-primary" />
                            Select Tickets
                        </h3>

                        <div class="space-y-4">
                            {#each event.ticket_types as ticket}
                                <div class="rounded-lg border border-border p-4">
                                    <div class="mb-2 flex items-start justify-between">
                                        <div>
                                            <p class="font-semibold">{ticket.name}</p>
                                            <p class="text-sm text-muted-foreground">{ticket.description}</p>
                                        </div>
                                        <p class="text-lg font-bold text-primary">
                                            ৳{(ticket.price_paisa / 100).toFixed(0)}
                                        </p>
                                    </div>

                                    <div class="mb-2 text-xs text-muted-foreground">
                                        {ticket.available_quantity > 0
                                            ? `${ticket.available_quantity} of ${ticket.total_quantity} available`
                                            : "Sold Out"}
                                    </div>

                                    <div class="flex items-center justify-between">
                                        <div class="flex items-center gap-2">
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                class="size-8 p-0"
                                                onclick={() => decrementTicket(ticket.id)}
                                                disabled={!selectedTickets[ticket.id]}
                                            >
                                                -
                                            </Button>
                                            <span class="w-8 text-center font-medium">
                                                {selectedTickets[ticket.id] || 0}
                                            </span>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                class="size-8 p-0"
                                                onclick={() => incrementTicket(ticket.id, ticket.available_quantity)}
                                                disabled={ticket.available_quantity === 0 || (selectedTickets[ticket.id] || 0) >= 5}
                                            >
                                                +
                                            </Button>
                                        </div>
                                    </div>
                                </div>
                            {/each}
                        </div>

                        {#if totalTickets > 0}
                            <div class="mt-6 space-y-4">
                                <Separator />
                                <div class="flex justify-between text-sm">
                                    <span>Tickets ({totalTickets})</span>
                                    <span class="font-bold">৳{(totalPrice / 100).toFixed(2)}</span>
                                </div>
                                <Button
                                    size="lg"
                                    class="w-full"
                                    onclick={proceedToCheckout}
                                >
                                    <Check class="mr-2 h-4 w-4" />
                                    Checkout ৳{(totalPrice / 100).toFixed(2)}
                                </Button>
                            </div>
                        {/if}

                        <Button
                            variant="outline"
                            class="mt-4 w-full gap-2"
                        >
                            <Share2 size={16} />
                            Share Event
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    {:else}
        <div class="flex h-[50vh] items-center justify-center">
            <p>Event not found</p>
        </div>
    {/if}
</div>
