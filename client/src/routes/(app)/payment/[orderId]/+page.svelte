<script lang="ts">
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import { orderApi } from "$lib/api/order";
    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";
    import {
        CreditCard,
        Phone,
        Wallet,
        Shield,
        Loader2,
        AlertCircle,
        Check,
        X,
        ExternalLink,
    } from "@lucide/svelte";
    import { toast } from "svelte-sonner";

    let orderId = $derived($page.params.orderId);
    let order = $state<any>(null);
    let isLoading = $state(true);
    let isProcessing = $state(false);
    let paymentStatus = $state<"pending" | "processing" | "success" | "failed">("pending");
    let errorMessage = $state("");

    const paymentMethods = [
        {
            id: "sslcommerz",
            name: "Credit/Debit Card",
            description: "Visa, Mastercard, Amex",
            icon: CreditCard,
            color: "blue",
        },
        {
            id: "bkash",
            name: "bKash",
            description: "Mobile wallet",
            icon: Phone,
            color: "pink",
        },
        {
            id: "nagad",
            name: "Nagad",
            description: "Digital financial service",
            icon: Wallet,
            color: "orange",
        },
    ];

    async function fetchOrder() {
        try {
            order = await orderApi.getOrder(orderId!);
            if (order.payment_status === "captured") {
                goto(`/confirmation/${orderId}`);
            }
        } catch (error) {
            toast.error("Failed to load order");
            goto("/orders");
        } finally {
            isLoading = false;
        }
    }

    async function initiatePayment(method: string) {
        isProcessing = true;
        paymentStatus = "processing";

        try {
            // In production, this would call the payment API to get a redirect URL
            // For now, simulate the payment flow
            
            // Simulate payment gateway redirect
            const paymentUrl = getPaymentGatewayUrl(order, method);
            
            if (paymentUrl) {
                // Redirect to payment gateway
                window.location.href = paymentUrl;
            } else {
                // Mock payment for development
                await simulatePayment();
            }
        } catch (error: any) {
            paymentStatus = "failed";
            errorMessage = error.message || "Payment failed. Please try again.";
            toast.error("Payment failed", {
                description: errorMessage,
            });
        } finally {
            isProcessing = false;
        }
    }

    function getPaymentGatewayUrl(order: any, method: string): string | null {
        // In production, call the payment API to get the redirect URL
        const baseUrl = import.meta.env.VITE_API_URL || "http://localhost:8888";
        
        // For now, return null to use mock payment flow
        // In production, this would be:
        // return `${baseUrl}/v1/payments?order_id=${order.id}&method=${method}`;
        return null;
    }

    async function simulatePayment() {
        // Simulate payment processing delay
        await new Promise((resolve) => setTimeout(resolve, 3000));

        // Simulate 90% success rate for testing
        const success = Math.random() > 0.1;

        if (success) {
            paymentStatus = "success";
            toast.success("Payment successful!", {
                description: "Your booking is confirmed",
            });
            
            // Redirect to confirmation after short delay
            setTimeout(() => {
                goto(`/confirmation/${orderId}`);
            }, 2000);
        } else {
            paymentStatus = "failed";
            errorMessage = "Payment was declined. Please try a different method.";
            toast.error("Payment declined", {
                description: errorMessage,
            });
        }
    }

    $effect(() => {
        if (orderId) fetchOrder();
    });
</script>

<div class="min-h-screen bg-muted/30 pb-32 pt-20">
    <div class="container mx-auto max-w-4xl px-4">
        {#if isLoading}
            <div class="flex h-[50vh] flex-col items-center justify-center gap-4">
                <Loader2 class="animate-spin text-primary" size={48} />
                <p class="text-muted-foreground">Loading order details...</p>
            </div>
        {:else if order}
            <!-- Header -->
            <div class="mb-8 text-center">
                <h1 class="text-3xl font-bold">Complete Payment</h1>
                <p class="mt-2 text-muted-foreground">
                    {order.route_name || `${order.from_station_name || order.from_station_id} → ${order.to_station_name || order.to_station_id}`} • {(order.passengers || []).length} passenger(s)
                </p>
            </div>

            <div class="grid gap-8 md:grid-cols-3">
                <!-- Left: Payment Methods -->
                <div class="md:col-span-2">
                    <div class="glass-card rounded-xl p-6">
                        <h3 class="mb-6 text-lg font-bold">
                            Select Payment Method
                        </h3>

                        <div class="space-y-3">
                            {#each paymentMethods as method}
                                <button
                                    class="flex w-full items-center justify-between rounded-xl border border-border bg-white/50 p-4 text-left transition-all hover:bg-white/80 hover:shadow-md disabled:cursor-not-allowed disabled:opacity-50 backdrop-blur-sm"
                                    onclick={() => initiatePayment(method.id)}
                                    disabled={isProcessing}
                                >
                                    <div class="flex items-center gap-4">
                                        <div
                                            class="flex size-12 items-center justify-center rounded-xl {method.color === 'blue' ? 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400' : method.color === 'pink' ? 'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400' : 'bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400'}"
                                        >
                                            <method.icon size={24} />
                                        </div>
                                        <div>
                                            <p class="font-semibold">
                                                {method.name}
                                            </p>
                                            <p class="text-sm text-muted-foreground">
                                                {method.description}
                                            </p>
                                        </div>
                                    </div>
                                    <ExternalLink
                                        size={18}
                                        class="text-muted-foreground"
                                    />
                                </button>
                            {/each}
                        </div>

                        <!-- Processing State -->
                        {#if paymentStatus === "processing"}
                            <div
                                class="mt-6 flex flex-col items-center justify-center rounded-xl bg-blue-50 p-8 text-center dark:bg-blue-900/20"
                            >
                                <Loader2
                                    class="mb-4 animate-spin text-blue-600"
                                    size={48}
                                />
                                <h4 class="text-lg font-semibold text-blue-700 dark:text-blue-400">
                                    Processing Payment
                                </h4>
                                <p class="text-sm text-blue-600 dark:text-blue-500">
                                    Please wait while we process your payment...
                                </p>
                            </div>
                        {/if}

                        <!-- Success State -->
                        {#if paymentStatus === "success"}
                            <div
                                class="mt-6 flex flex-col items-center justify-center rounded-xl bg-green-50 p-8 text-center dark:bg-green-900/20"
                            >
                                <div
                                    class="mb-4 flex size-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30"
                                >
                                    <Check
                                        size={32}
                                        class="text-green-600 dark:text-green-400"
                                    />
                                </div>
                                <h4 class="text-lg font-semibold text-green-700 dark:text-green-400">
                                    Payment Successful!
                                </h4>
                                <p class="text-sm text-green-600 dark:text-green-500">
                                    Redirecting to confirmation...
                                </p>
                            </div>
                        {/if}

                        <!-- Failed State -->
                        {#if paymentStatus === "failed"}
                            <div
                                class="mt-6 flex flex-col items-center justify-center rounded-xl bg-red-50 p-8 text-center dark:bg-red-900/20"
                            >
                                <div
                                    class="mb-4 flex size-16 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30"
                                >
                                    <X
                                        size={32}
                                        class="text-red-600 dark:text-red-400"
                                    />
                                </div>
                                <h4 class="text-lg font-semibold text-red-700 dark:text-red-400">
                                    Payment Failed
                                </h4>
                                <p class="mb-4 text-sm text-red-600 dark:text-red-500">
                                    {errorMessage}
                                </p>
                                <Button
                                    variant="outline"
                                    onclick={() => {
                                        paymentStatus = "pending";
                                        errorMessage = "";
                                    }}
                                >
                                    Try Again
                                </Button>
                            </div>
                        {/if}

                        <!-- Security Note -->
                        <div
                            class="mt-6 flex items-start gap-3 rounded-xl bg-muted/50 p-4"
                        >
                            <Shield
                                size={20}
                                class="mt-0.5 shrink-0 text-green-600"
                            />
                            <div class="text-sm text-muted-foreground">
                                <p class="font-medium text-foreground">
                                    Secure Payment
                                </p>
                                <p>
                                    Your payment information is encrypted with
                                    256-bit SSL. We never store your card
                                    details.
                                </p>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Right: Order Summary -->
                <div>
                    <div class="glass-card sticky top-24 rounded-xl p-6">
                        <h3 class="mb-4 text-lg font-bold">Order Summary</h3>

                        <div class="space-y-3 text-sm">
                            <div class="flex justify-between">
                                <span class="text-muted-foreground"
                                    >Subtotal</span
                                >
                                <span
                                    >৳{(order.subtotal_paisa / 100).toFixed(2)}</span
                                >
                            </div>
                            <div class="flex justify-between">
                                <span class="text-muted-foreground"
                                    >Tax & Service Charge</span
                                >
                                <span
                                    >৳{(order.tax_paisa / 100).toFixed(2)}</span
                                >
                            </div>
                            <div class="flex justify-between">
                                <span class="text-muted-foreground"
                                    >Booking Fee</span
                                >
                                <span
                                    >৳{(order.booking_fee_paisa / 100).toFixed(2)}</span
                                >
                            </div>
                            {#if order.discount_paisa > 0}
                                <div class="flex justify-between text-green-600">
                                    <span>Discount</span>
                                    <span
                                        >-৳{(order.discount_paisa / 100).toFixed(2)}</span
                                    >
                                </div>
                            {/if}
                        </div>

                        <Separator class="my-4" />

                        <div class="mb-4 flex justify-between text-xl font-bold">
                            <span>Total</span>
                            <span class="text-primary"
                                >৳{(order.total_paisa / 100).toFixed(2)}</span
                            >
                        </div>

                        <div class="space-y-2 text-xs text-muted-foreground">
                            <div class="flex items-center gap-2">
                                <AlertCircle size={14} />
                                <span>
                                    Payment must be completed within 15 minutes
                                </span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        {:else}
            <div class="flex h-[50vh] items-center justify-center">
                <p>Order not found</p>
            </div>
        {/if}
    </div>
</div>
