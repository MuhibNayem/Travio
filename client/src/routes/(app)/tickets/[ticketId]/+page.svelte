<script lang="ts">
    import { page } from "$app/stores";
    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";
    import {
        Ticket,
        Download,
        Calendar,
        MapPin,
        Clock,
        User,
        Armchair,
        Loader2,
        Check,
        X,
        AlertTriangle,
        QrCode,
    } from "@lucide/svelte";
    import QRCode from "qrcode";
    import { toast } from "svelte-sonner";

    let ticketId = $derived($page.params.ticketId);
    let ticket = $state<any>(null);
    let isLoading = $state(true);
    let qrCodeUrl = $state("");
    let isDownloading = $state(false);

    async function fetchTicket() {
        try {
            // Mock ticket data - in production, call API
            ticket = {
                id: ticketId,
                booking_id: "BK-12345",
                order_id: "ORD-67890",
                trip_id: "TRIP-001",
                route_name: "Dhaka → Chittagong",
                from_station: "Dhaka",
                to_station: "Chittagong",
                departure_time: "2026-04-15T08:00:00Z",
                arrival_time: "2026-04-15T14:00:00Z",
                passenger_name: "John Doe",
                passenger_nid: "1234567890",
                seat_number: "A12",
                seat_class: "AC Deluxe",
                price_paisa: 80000,
                currency: "BDT",
                status: "active",
                created_at: "2026-04-07T10:00:00Z",
            };

            // Generate QR code
            const qrData = JSON.stringify({
                ticket_id: ticket.id,
                booking_id: ticket.booking_id,
                passenger: ticket.passenger_name,
                seat: ticket.seat_number,
            });
            qrCodeUrl = await QRCode.toDataURL(qrData, {
                width: 256,
                margin: 2,
            });
        } catch (error) {
            toast.error("Failed to load ticket");
        } finally {
            isLoading = false;
        }
    }

    function getStatusBadge(status: string) {
        switch (status) {
            case "active":
                return {
                    color: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
                    icon: Check,
                    text: "Active",
                };
            case "used":
                return {
                    color: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400",
                    icon: Check,
                    text: "Used",
                };
            case "cancelled":
                return {
                    color: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400",
                    icon: X,
                    text: "Cancelled",
                };
            case "expired":
                return {
                    color: "bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400",
                    icon: AlertTriangle,
                    text: "Expired",
                };
            default:
                return {
                    color: "bg-gray-100 text-gray-800",
                    icon: AlertTriangle,
                    text: status,
                };
        }
    }

    async function downloadPDF() {
        isDownloading = true;
        try {
            toast.success("Ticket download started");
        } catch (error) {
            toast.error("Failed to download ticket");
        } finally {
            isDownloading = false;
        }
    }

    $effect(() => {
        if (ticketId) fetchTicket();
    });
</script>

<div class="min-h-screen bg-muted/30 py-20">
    <div class="container mx-auto max-w-3xl px-4">
        {#if isLoading}
            <div class="flex h-[50vh] flex-col items-center justify-center gap-4">
                <Loader2 class="animate-spin text-primary" size={48} />
                <p class="text-muted-foreground">Loading ticket...</p>
            </div>
        {:else if ticket}
            {@const statusBadge = getStatusBadge(ticket.status)}

            <!-- Ticket Header -->
            <div class="mb-8 text-center">
                <div class="mb-4 flex items-center justify-center gap-3">
                    <Ticket size={32} class="text-primary" />
                    <h1 class="text-3xl font-bold">E-Ticket</h1>
                </div>
                <span
                    class="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-sm font-semibold {statusBadge.color}"
                >
                    <statusBadge.icon size={14} />
                    {statusBadge.text}
                </span>
            </div>

            <!-- Ticket Card -->
            <div class="glass-card mb-8 overflow-hidden rounded-2xl shadow-xl">
                <!-- Header -->
                <div class="bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-6 text-white">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-sm text-white/80">Booking ID</p>
                            <p class="text-xl font-bold">
                                {ticket.booking_id}
                            </p>
                        </div>
                        <div class="text-right">
                            <p class="text-sm text-white/80">Ticket ID</p>
                            <p class="text-sm font-mono">
                                {ticket.id.slice(0, 12)}...
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Body -->
                <div class="px-6 py-8">
                    <!-- Route -->
                    <div class="mb-6 text-center">
                        <div class="flex items-center justify-center gap-4">
                            <div class="text-right">
                                <p class="text-2xl font-black">
                                    {ticket.from_station}
                                </p>
                                <p class="text-sm text-muted-foreground">
                                    {new Date(
                                        ticket.departure_time,
                                    ).toLocaleTimeString([], {
                                        hour: "2-digit",
                                        minute: "2-digit",
                                    })}
                                </p>
                            </div>
                            <div class="flex flex-col items-center">
                                <Clock size={20} class="text-primary" />
                                <div class="mx-2 h-0.5 w-16 bg-border"></div>
                                <p class="text-xs text-muted-foreground">
                                    6 hours
                                </p>
                            </div>
                            <div class="text-left">
                                <p class="text-2xl font-black">
                                    {ticket.to_station}
                                </p>
                                <p class="text-sm text-muted-foreground">
                                    {new Date(
                                        ticket.arrival_time,
                                    ).toLocaleTimeString([], {
                                        hour: "2-digit",
                                        minute: "2-digit",
                                    })}
                                </p>
                            </div>
                        </div>
                    </div>

                    <Separator class="my-6" />

                    <!-- Details Grid -->
                    <div class="grid gap-4 sm:grid-cols-2">
                        <div>
                            <p class="text-xs text-muted-foreground">
                                Passenger
                            </p>
                            <p class="flex items-center gap-2 font-medium">
                                <User size={16} class="text-primary" />
                                {ticket.passenger_name}
                            </p>
                        </div>
                        <div>
                            <p class="text-xs text-muted-foreground">Seat</p>
                            <p class="flex items-center gap-2 font-medium">
                                <Armchair size={16} class="text-primary" />
                                {ticket.seat_number} ({ticket.seat_class})
                            </p>
                        </div>
                        <div>
                            <p class="text-xs text-muted-foreground">Date</p>
                            <p class="flex items-center gap-2 font-medium">
                                <Calendar size={16} class="text-primary" />
                                {new Date(
                                    ticket.departure_time,
                                ).toLocaleDateString()}
                            </p>
                        </div>
                        <div>
                            <p class="text-xs text-muted-foreground">
                                Amount Paid
                            </p>
                            <p class="text-xl font-bold text-primary">
                                ৳{(ticket.price_paisa / 100).toFixed(2)}
                            </p>
                        </div>
                    </div>
                </div>

                <!-- QR Code Section -->
                <div class="border-t border-dashed border-border px-6 py-8">
                    <div class="flex flex-col items-center">
                        <div class="mb-4 rounded-xl bg-white p-4 shadow-lg">
                            {#if qrCodeUrl}
                                <img
                                    src={qrCodeUrl}
                                    alt="Ticket QR Code"
                                    class="size-64"
                                />
                            {:else}
                                <div class="flex size-64 items-center justify-center">
                                    <Loader2 class="animate-spin text-primary" />
                                </div>
                            {/if}
                        </div>
                        <p class="text-sm text-muted-foreground">
                            Show this QR code to the conductor
                        </p>
                    </div>
                </div>
            </div>

            <!-- Actions -->
            <div class="flex flex-col gap-3 sm:flex-row">
                <Button
                    class="flex-1 gap-2"
                    size="lg"
                    onclick={downloadPDF}
                    disabled={isDownloading}
                >
                    {#if isDownloading}
                        <Loader2 class="h-4 w-4 animate-spin" />
                    {:else}
                        <Download size={18} />
                    {/if}
                    Download PDF
                </Button>
                <Button variant="outline" class="flex-1 gap-2" href="/orders">
                    Back to Bookings
                </Button>
            </div>

            <!-- Important Info -->
            <div class="mt-6 rounded-xl bg-muted/50 p-4">
                <h4 class="mb-2 font-semibold">Important Notes</h4>
                <ul class="space-y-1 text-sm text-muted-foreground">
                    <li>• Arrive at the station 30 minutes before departure</li>
                    <li>• Carry a valid ID matching the passenger name</li>
                    <li>• This ticket is non-transferable</li>
                    <li>• Cancellation must be done 24 hours before departure</li>
                </ul>
            </div>
        {:else}
            <div class="flex h-[50vh] items-center justify-center">
                <p>Ticket not found</p>
            </div>
        {/if}
    </div>
</div>
