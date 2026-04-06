<script lang="ts">
    import { page } from "$app/stores";
    import { orderApi } from "$lib/api/order";
    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";
    import {
        Check,
        Ticket,
        Download,
        Share2,
        Calendar,
        Mail,
        MessageCircle,
        MapPin,
        Clock,
        User,
        Armchair,
        Loader2,
    } from "@lucide/svelte";
    import { toast } from "svelte-sonner";
    import confetti from "canvas-confetti";

    let orderId = $derived($page.params.orderId);
    let order = $state<any>(null);
    let isLoading = $state(true);
    let isDownloading = $state(false);

    async function fetchOrder() {
        try {
            order = await orderApi.getOrder(orderId!);
            
            // Trigger confetti animation
            triggerConfetti();
        } catch (error) {
            toast.error("Failed to load booking details");
        } finally {
            isLoading = false;
        }
    }

    function triggerConfetti() {
        const duration = 3000;
        const end = Date.now() + duration;

        const colors = ["#3B82F6", "#10B981", "#F59E0B", "#EF4444"];

        (function frame() {
            confetti({
                particleCount: 3,
                angle: 60,
                spread: 55,
                origin: { x: 0 },
                colors: colors,
            });
            confetti({
                particleCount: 3,
                angle: 120,
                spread: 55,
                origin: { x: 1 },
                colors: colors,
            });

            if (Date.now() < end) {
                requestAnimationFrame(frame);
            }
        })();
    }

    async function downloadTicket() {
        isDownloading = true;
        try {
            // In production, call the fulfillment API to get PDF
            toast.info("Ticket download started", {
                description: "You'll receive it via email shortly",
            });
        } catch (error) {
            toast.error("Failed to download ticket");
        } finally {
            isDownloading = false;
        }
    }

    function shareViaWhatsApp() {
        const text = `I just booked a trip on TicketNation! Order #${orderId?.slice(0, 8)}`;
        window.open(`https://wa.me/?text=${encodeURIComponent(text)}`, "_blank");
    }

    function shareViaEmail() {
        const subject = `Booking Confirmation - Order #${orderId?.slice(0, 8)}`;
        const body = `My booking has been confirmed! Order ID: ${orderId}`;
        window.location.href = `mailto:?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`;
    }

    $effect(() => {
        if (orderId) fetchOrder();
    });
</script>

<div class="min-h-screen bg-gradient-to-b from-green-50 to-muted/30 pb-32 pt-20 dark:from-green-950/20">
    <div class="container mx-auto max-w-3xl px-4">
        {#if isLoading}
            <div class="flex h-[50vh] flex-col items-center justify-center gap-4">
                <Loader2 class="animate-spin text-primary" size={48} />
                <p class="text-muted-foreground">Loading booking details...</p>
            </div>
        {:else if order}
            <!-- Success Header -->
            <div class="mb-8 text-center">
                <div
                    class="mx-auto mb-6 flex size-24 items-center justify-center rounded-full bg-gradient-to-br from-green-400 to-green-600 shadow-xl shadow-green-500/30"
                >
                    <Check size={48} class="text-white" />
                </div>
                <h1 class="text-4xl font-black text-foreground">
                    Booking Confirmed!
                </h1>
                <p class="mt-2 text-lg text-muted-foreground">
                    Your tickets have been sent to {order.contact_email}
                </p>
            </div>

            <!-- Ticket Card -->
            <div class="glass-card mb-8 overflow-hidden rounded-2xl shadow-2xl">
                <!-- Ticket Header -->
                <div class="bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-8 text-white">
                    <div class="flex items-center justify-between">
                        <div class="flex items-center gap-3">
                            <Ticket size={32} />
                            <div>
                                <h2 class="text-2xl font-bold">E-Ticket</h2>
                                <p class="text-sm text-white/80">
                                    Order #{order.id.slice(0, 8)}
                                </p>
                            </div>
                        </div>
                        <div class="text-right">
                            <p class="text-sm text-white/80">Booking ID</p>
                            <p class="text-xl font-bold">
                                {order.booking_id?.slice(0, 8) || "N/A"}
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Ticket Body -->
                <div class="px-6 py-8">
                    <!-- Passengers -->
                    <div class="mb-6">
                        <h3 class="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                            <User size={16} />
                            Passengers
                        </h3>
                        <div class="space-y-2">
                            {#each order.passengers || [] as passenger}
                                <div class="flex items-center justify-between rounded-lg bg-muted/50 px-4 py-3">
                                    <div class="flex items-center gap-3">
                                        <Armchair
                                            size={18}
                                            class="text-primary"
                                        />
                                        <div>
                                            <p class="font-medium">
                                                {passenger.name}
                                            </p>
                                            <p class="text-xs text-muted-foreground">
                                                Seat: {passenger.seat_id}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    </div>

                    <Separator class="my-6" />

                    <!-- Trip Details -->
                    <div class="grid gap-4 sm:grid-cols-2">
                        <div>
                            <p class="text-xs text-muted-foreground">
                                From
                            </p>
                            <p class="flex items-center gap-2 text-lg font-bold">
                                <MapPin size={18} class="text-primary" />
                                {order.from_station_id}
                            </p>
                        </div>
                        <div>
                            <p class="text-xs text-muted-foreground">To</p>
                            <p class="flex items-center gap-2 text-lg font-bold">
                                <MapPin size={18} class="text-green-600" />
                                {order.to_station_id}
                            </p>
                        </div>
                        <div>
                            <p class="text-xs text-muted-foreground">
                                Date & Time
                            </p>
                            <p class="flex items-center gap-2 font-medium">
                                <Clock size={16} class="text-primary" />
                                {new Date(
                                    order.created_at,
                                ).toLocaleDateString()}
                            </p>
                        </div>
                        <div>
                            <p class="text-xs text-muted-foreground">
                                Total Paid
                            </p>
                            <p class="text-lg font-bold text-primary">
                                ৳{(order.total_paisa / 100).toFixed(2)}
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Dashed line -->
                <div class="mx-6 border-t-2 border-dashed border-border"></div>

                <!-- Ticket Footer -->
                <div class="px-6 py-6">
                    <div class="flex flex-col gap-3 sm:flex-row sm:gap-4">
                        <Button
                            class="flex-1 gap-2"
                            size="lg"
                            onclick={downloadTicket}
                            disabled={isDownloading}
                        >
                            {#if isDownloading}
                                <Loader2
                                    class="h-4 w-4 animate-spin"
                                />
                            {:else}
                                <Download size={18} />
                            {/if}
                            Download PDF
                        </Button>
                        <Button
                            variant="outline"
                            class="flex-1 gap-2"
                            size="lg"
                            onclick={shareViaEmail}
                        >
                            <Mail size={18} />
                            Email
                        </Button>
                        <Button
                            variant="outline"
                            class="flex-1 gap-2"
                            size="lg"
                            onclick={shareViaWhatsApp}
                        >
                            <MessageCircle size={18} />
                            Share
                        </Button>
                    </div>
                </div>
            </div>

            <!-- Next Steps -->
            <div class="glass-card rounded-xl p-6">
                <h3 class="mb-4 text-lg font-bold">What's Next?</h3>
                <div class="space-y-3">
                    <div class="flex items-start gap-3">
                        <div
                            class="mt-1 flex size-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400"
                        >
                            1
                        </div>
                        <div>
                            <p class="font-medium">Check your email</p>
                            <p class="text-sm text-muted-foreground">
                                We've sent a confirmation email with your
                                e-ticket attached
                            </p>
                        </div>
                    </div>
                    <div class="flex items-start gap-3">
                        <div
                            class="mt-1 flex size-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400"
                        >
                            2
                        </div>
                        <div>
                            <p class="font-medium">Download your ticket</p>
                            <p class="text-sm text-muted-foreground">
                                Save the PDF to your phone or print it out
                            </p>
                        </div>
                    </div>
                    <div class="flex items-start gap-3">
                        <div
                            class="mt-1 flex size-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400"
                        >
                            3
                        </div>
                        <div>
                            <p class="font-medium">Show ticket at boarding</p>
                            <p class="text-sm text-muted-foreground">
                                Present the QR code to the conductor
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Actions -->
            <div class="mt-8 flex justify-center gap-4">
                <Button variant="outline" href="/orders">
                    View All Bookings
                </Button>
                <Button href="/search">Book Another Trip</Button>
            </div>
        {:else}
            <div class="flex h-[50vh] items-center justify-center">
                <p>Booking not found</p>
            </div>
        {/if}
    </div>
</div>

<style>
    @keyframes confetti-fall {
        0% {
            transform: translateY(0) rotate(0deg);
            opacity: 1;
        }
        100% {
            transform: translateY(100vh) rotate(720deg);
            opacity: 0;
        }
    }
</style>
