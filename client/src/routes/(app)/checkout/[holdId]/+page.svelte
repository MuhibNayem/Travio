<script lang="ts">
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import { inventoryApi } from "$lib/api/inventory";
    import { catalogApi } from "$lib/api/catalog";
    import { orderApi } from "$lib/api/order";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import {
        Select,
        SelectContent,
        SelectItem,
        SelectTrigger,
        SelectValue,
    } from "$lib/components/ui/select";
    import { Checkbox } from "$lib/components/ui/checkbox";
    import {
        User,
        Phone,
        Mail,
        Calendar,
        CreditCard,
        Shield,
        Clock,
        AlertTriangle,
        Loader2,
        Ticket,
        Percent,
    } from "@lucide/svelte";
    import { toast } from "svelte-sonner";
    import { auth } from "$lib/runes/auth.svelte";

    let holdId = $derived($page.params.holdId);

    // Hold state
    let holdData = $state<any>(null);
    let isLoading = $state(true);
    let holdExpiry = $state<Date | null>(null);
    let timeRemaining = $state(0);

    // Passenger details
    let passengers = $state<any[]>([]);
    let contactEmail = $state("");
    let contactPhone = $state("");

    // Coupon
    let couponCode = $state("");
    let couponValidated = $state(false);
    let couponDiscount = $state(0);
    let couponMessage = $state("");
    let isCheckingCoupon = $state(false);

    // Payment
    let paymentMethod = $state("sslcommerz");
    let acceptedTerms = $state(false);

    // Booking
    let isBooking = $state(false);

    // Pricing
    let subtotal = $derived(
        passengers.reduce((sum, p) => sum + (p.price || 0), 0),
    );
    let tax = $derived(subtotal * 0.05); // 5% tax
    let bookingFee = $derived(passengers.length * 20); // 20 BDT per passenger
    let discount = $derived(couponDiscount);
    let total = $derived(subtotal + tax + bookingFee - discount);

    // Countdown timer
    let countdownInterval: any = null;

    function startCountdown() {
        if (!holdExpiry) return;

        const update = () => {
            const now = new Date();
            const diff = holdExpiry.getTime() - now.getTime();
            timeRemaining = Math.max(0, Math.floor(diff / 1000));

            if (timeRemaining <= 0) {
                clearInterval(countdownInterval);
                toast.error("Hold expired", {
                    description: "Your seat hold has expired. Please search again.",
                });
                goto("/search");
            }
        };

        update();
        countdownInterval = setInterval(update, 1000);
    }

    function formatTime(seconds: number): string {
        const m = Math.floor(seconds / 60);
        const s = seconds % 60;
        return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
    }

    async function fetchHoldDetails() {
        try {
            // Fetch hold details from inventory service
            const response = await inventoryApi.getHoldStatus(holdId!);
            holdData = response;
            holdExpiry = new Date(response.expires_at);

            // Initialize passenger forms
            const seatIds = response.held_seat_ids || [];
            passengers = seatIds.map((seatId: string, index: number) => ({
                seat_id: seatId,
                seat_number: `Seat ${index + 1}`,
                nid: "",
                name: "",
                date_of_birth: "",
                gender: "male",
                phone: "",
                email: "",
                price: (response.total_paisa || 0) / seatIds.length / 100,
            }));

            // Pre-fill contact from first passenger
            if (passengers.length > 0) {
                contactEmail = passengers[0].email;
                contactPhone = passengers[0].phone;
            }

            startCountdown();
        } catch (error) {
            console.error("Failed to fetch hold details:", error);
            toast.error("Failed to load checkout details");
            goto("/search");
        } finally {
            isLoading = false;
        }
    }

    async function applyCoupon() {
        if (!couponCode.trim()) {
            toast.error("Enter a coupon code");
            return;
        }

        isCheckingCoupon = true;
        try {
            // Call CRM validation endpoint
            const response = await fetch(
                `${import.meta.env.VITE_API_URL}/v1/coupons/validate`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    credentials: "include",
                    body: JSON.stringify({
                        code: couponCode,
                        org_id: auth.user?.organizationId || "",
                        cart_total: Math.round(total * 100),
                    }),
                },
            );

            const data = await response.json();

            if (data.valid) {
                couponValidated = true;
                couponDiscount = (data.discount_paisa || 0) / 100;
                couponMessage = data.message || "Coupon applied!";
                toast.success("Coupon applied", {
                    description: `You saved ৳${couponDiscount.toFixed(2)}`,
                });
            } else {
                couponValidated = false;
                couponDiscount = 0;
                couponMessage = data.message || "Invalid coupon";
                toast.error("Invalid coupon", {
                    description: data.message,
                });
            }
        } catch (error) {
            couponMessage = "Failed to validate coupon";
            toast.error("Coupon validation failed");
        } finally {
            isCheckingCoupon = false;
        }
    }

    async function proceedToPayment() {
        // Validate all passengers
        for (let i = 0; i < passengers.length; i++) {
            const p = passengers[i];
            if (!p.name?.trim()) {
                toast.error("Missing information", {
                    description: `Please enter name for passenger ${i + 1}`,
                });
                return;
            }
            if (!p.nid?.trim()) {
                toast.error("Missing information", {
                    description: `Please enter NID for passenger ${i + 1}`,
                });
                return;
            }
            if (!p.date_of_birth) {
                toast.error("Missing information", {
                    description: `Please enter date of birth for passenger ${i + 1}`,
                });
                return;
            }
        }

        if (!contactEmail.trim() || !contactPhone.trim()) {
            toast.error("Missing contact information", {
                description: "Please provide contact email and phone",
            });
            return;
        }

        if (!acceptedTerms) {
            toast.error("Accept terms & conditions", {
                description: "Please accept the terms to proceed",
            });
            return;
        }

        isBooking = true;

        try {
            // Create order via API
            const orderPayload = {
                trip_id: holdData.trip_id,
                from_station_id: holdData.from_station_id,
                to_station_id: holdData.to_station_id,
                hold_id: holdId,
                passengers: passengers.map((p) => ({
                    nid: p.nid,
                    name: p.name,
                    seat_id: p.seat_id,
                    date_of_birth: p.date_of_birth,
                    gender: p.gender,
                })),
                payment_method: {
                    type: paymentMethod,
                },
                contact_email: contactEmail,
                contact_phone: contactPhone,
                coupon_code: couponValidated ? couponCode : "",
                idempotency_key: crypto.randomUUID(),
            };

            const order = await orderApi.createOrder(orderPayload);

            toast.success("Booking initiated", {
                description: "Redirecting to payment...",
            });

            // Redirect to payment page
            goto(`/payment/${order.id}`);
        } catch (error: any) {
            console.error("Booking failed:", error);
            toast.error("Booking failed", {
                description: error.message || "Please try again",
            });
        } finally {
            isBooking = false;
        }
    }

    $effect(() => {
        if (holdId) fetchHoldDetails();

        return () => {
            if (countdownInterval) clearInterval(countdownInterval);
        };
    });
</script>

<div class="min-h-screen bg-muted/30 pb-32 pt-20">
    <div class="container mx-auto max-w-6xl px-4">
        {#if isLoading}
            <div class="flex h-[50vh] flex-col items-center justify-center gap-4">
                <Loader2 class="animate-spin text-primary" size={48} />
                <p class="text-muted-foreground">Loading checkout details...</p>
            </div>
        {:else if holdData}
            <!-- Header with countdown -->
            <div class="mb-8 flex items-center justify-between">
                <div>
                    <h1 class="text-3xl font-bold">Complete Your Booking</h1>
                    <p class="text-muted-foreground mt-1">
                        Fill in passenger details and proceed to payment
                    </p>
                </div>
                <div
                    class="flex items-center gap-3 rounded-xl bg-white px-6 py-4 shadow-sm"
                >
                    <Clock size={20} class="text-orange-500" />
                    <div class="text-right">
                        <p class="text-xs text-muted-foreground">
                            Hold expires in
                        </p>
                        <p
                            class="text-2xl font-bold tabular-nums {timeRemaining < 300 ? 'text-red-500 animate-pulse' : 'text-foreground'}"
                        >
                            {formatTime(timeRemaining)}
                        </p>
                    </div>
                </div>
            </div>

            <div class="grid gap-8 lg:grid-cols-3">
                <!-- Left: Passenger Details -->
                <div class="space-y-6 lg:col-span-2">
                    {#each passengers as passenger, index}
                        <div class="glass-card rounded-xl p-6">
                            <div class="mb-4 flex items-center gap-3">
                                <div
                                    class="flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary"
                                >
                                    <User size={20} />
                                </div>
                                <div>
                                    <h3 class="text-lg font-bold">
                                        Passenger {index + 1}
                                    </h3>
                                    <p class="text-sm text-muted-foreground">
                                        {passenger.seat_number}
                                    </p>
                                </div>
                            </div>

                            <div class="grid gap-4 md:grid-cols-2">
                                <div class="space-y-2">
                                    <label class="text-sm font-medium"
                                        >Full Name *</label
                                    >
                                    <Input
                                        type="text"
                                        bind:value={passenger.name}
                                        placeholder="As per NID"
                                        class="bg-white/50 backdrop-blur-sm"
                                    />
                                </div>

                                <div class="space-y-2">
                                    <label class="text-sm font-medium"
                                        >NID Number *</label
                                    >
                                    <Input
                                        type="text"
                                        bind:value={passenger.nid}
                                        placeholder="National ID"
                                        class="bg-white/50 backdrop-blur-sm"
                                    />
                                </div>

                                <div class="space-y-2">
                                    <label class="text-sm font-medium"
                                        >Date of Birth *</label
                                    >
                                    <div class="relative">
                                        <Calendar
                                            class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                                            size={16}
                                        />
                                        <Input
                                            type="date"
                                            bind:value={passenger.date_of_birth}
                                            class="bg-white/50 backdrop-blur-sm pl-10"
                                        />
                                    </div>
                                </div>

                                <div class="space-y-2">
                                    <label class="text-sm font-medium"
                                        >Gender</label
                                    >
                                    <Select
                                        bind:value={passenger.gender}
                                    >
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select gender" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="male"
                                                >Male</SelectItem
                                            >
                                            <SelectItem value="female"
                                                >Female</SelectItem
                                            >
                                            <SelectItem value="other"
                                                >Other</SelectItem
                                            >
                                        </SelectContent>
                                    </Select>
                                </div>

                                <div class="space-y-2">
                                    <label class="text-sm font-medium"
                                        >Phone</label
                                    >
                                    <div class="relative">
                                        <Phone
                                            class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                                            size={16}
                                        />
                                        <Input
                                            type="tel"
                                            bind:value={passenger.phone}
                                            placeholder="+880 1XXX-XXXXXX"
                                            class="bg-white/50 backdrop-blur-sm pl-10"
                                        />
                                    </div>
                                </div>

                                <div class="space-y-2">
                                    <label class="text-sm font-medium"
                                        >Email</label
                                    >
                                    <div class="relative">
                                        <Mail
                                            class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                                            size={16}
                                        />
                                        <Input
                                            type="email"
                                            bind:value={passenger.email}
                                            placeholder="email@example.com"
                                            class="bg-white/50 backdrop-blur-sm pl-10"
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>
                    {/each}

                    <!-- Contact Information -->
                    <div class="glass-card rounded-xl p-6">
                        <h3 class="mb-4 text-lg font-bold">
                            Contact Information
                        </h3>
                        <div class="grid gap-4 md:grid-cols-2">
                            <div class="space-y-2">
                                <label class="text-sm font-medium"
                                    >Email *</label
                                >
                                <div class="relative">
                                    <Mail
                                        class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                                        size={16}
                                    />
                                    <Input
                                        type="email"
                                        bind:value={contactEmail}
                                        placeholder="booking@example.com"
                                        class="bg-white/50 backdrop-blur-sm pl-10"
                                    />
                                </div>
                            </div>
                            <div class="space-y-2">
                                <label class="text-sm font-medium"
                                    >Phone *</label
                                >
                                <div class="relative">
                                    <Phone
                                        class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                                        size={16}
                                    />
                                    <Input
                                        type="tel"
                                        bind:value={contactPhone}
                                        placeholder="+880 1XXX-XXXXXX"
                                        class="bg-white/50 backdrop-blur-sm pl-10"
                                    />
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Coupon Code -->
                    <div class="glass-card rounded-xl p-6">
                        <h3 class="mb-4 text-lg font-bold">
                            Have a Coupon?
                        </h3>
                        <div class="flex gap-3">
                            <div class="relative flex-1">
                                <Percent
                                    class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                                    size={16}
                                />
                                <Input
                                    type="text"
                                    bind:value={couponCode}
                                    placeholder="Enter coupon code"
                                    class="bg-white/50 backdrop-blur-sm pl-10 uppercase"
                                    onkeydown={(e) => {
                                        if (e.key === "Enter") applyCoupon();
                                    }}
                                />
                            </div>
                            <Button
                                onclick={applyCoupon}
                                disabled={isCheckingCoupon || !couponCode}
                            >
                                {#if isCheckingCoupon}
                                    <Loader2
                                        class="mr-2 h-4 w-4 animate-spin"
                                    />
                                {:else}
                                    Apply
                                {/if}
                            </Button>
                        </div>
                        {#if couponMessage}
                            <p
                                class="mt-2 text-sm {couponValidated ? 'text-green-600' : 'text-red-600'}"
                            >
                                {couponMessage}
                            </p>
                        {/if}
                    </div>
                </div>

                <!-- Right: Summary & Payment -->
                <div class="space-y-6">
                    <div class="glass-card sticky top-24 rounded-xl p-6">
                        <h3 class="mb-4 text-lg font-bold">
                            Booking Summary
                        </h3>

                        <!-- Seats -->
                        <div class="mb-4 space-y-2">
                            {#each passengers as p}
                                <div
                                    class="flex items-center justify-between text-sm"
                                >
                                    <div class="flex items-center gap-2">
                                        <Ticket size={14} class="text-primary" />
                                        <span>{p.seat_number}</span>
                                    </div>
                                    <span class="font-medium"
                                        >৳{p.price?.toFixed(2)}</span
                                    >
                                </div>
                            {/each}
                        </div>

                        <Separator class="my-4" />

                        <!-- Pricing Breakdown -->
                        <div class="space-y-2 text-sm">
                            <div
                                class="flex justify-between text-muted-foreground"
                            >
                                <span>Subtotal</span>
                                <span>৳{subtotal.toFixed(2)}</span>
                            </div>
                            <div
                                class="flex justify-between text-muted-foreground"
                            >
                                <span>Tax & Service Charge (5%)</span>
                                <span>৳{tax.toFixed(2)}</span>
                            </div>
                            <div
                                class="flex justify-between text-muted-foreground"
                            >
                                <span>Booking Fee</span>
                                <span>৳{bookingFee.toFixed(2)}</span>
                            </div>
                            {#if discount > 0}
                                <div class="flex justify-between text-green-600">
                                    <span>Coupon Discount</span>
                                    <span>-৳{discount.toFixed(2)}</span>
                                </div>
                            {/if}
                        </div>

                        <Separator class="my-4" />

                        <div
                            class="mb-6 flex justify-between text-xl font-bold"
                        >
                            <span>Total</span>
                            <span class="text-primary"
                                >৳{total.toFixed(2)}</span
                            >
                        </div>

                        <!-- Payment Method -->
                        <div class="mb-4">
                            <label class="mb-2 block text-sm font-medium"
                                >Payment Method</label
                            >
                            <Select bind:value={paymentMethod}>
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="sslcommerz">
                                        <div class="flex items-center gap-2">
                                            <CreditCard size={16} />
                                            <span>Credit/Debit Card</span>
                                        </div>
                                    </SelectItem>
                                    <SelectItem value="bkash">
                                        <div class="flex items-center gap-2">
                                            <Phone size={16} />
                                            <span>bKash</span>
                                        </div>
                                    </SelectItem>
                                    <SelectItem value="nagad">
                                        <div class="flex items-center gap-2">
                                            <Phone size={16} />
                                            <span>Nagad</span>
                                        </div>
                                    </SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <!-- Terms & Conditions -->
                        <div class="mb-6 flex items-start gap-2">
                            <Checkbox
                                bind:checked={acceptedTerms}
                                id="terms"
                            />
                            <label
                                for="terms"
                                class="cursor-pointer text-sm leading-tight"
                            >
                                I agree to the
                                <a
                                    href="/terms"
                                    class="text-primary hover:underline"
                                    >Terms & Conditions</a
                                >
                                and
                                <a
                                    href="/privacy"
                                    class="text-primary hover:underline"
                                    >Privacy Policy</a
                                >
                            </label>
                        </div>

                        <!-- Checkout Button -->
                        <Button
                            size="lg"
                            class="w-full h-14 text-lg font-bold shadow-xl shadow-primary/20"
                            onclick={proceedToPayment}
                            disabled={isBooking || !acceptedTerms}
                        >
                            {#if isBooking}
                                <Loader2 class="mr-2 h-5 w-5 animate-spin" />
                                Processing...
                            {:else}
                                <Shield class="mr-2" size={20} />
                                Proceed to Pay ৳{total.toFixed(2)}
                            {/if}
                        </Button>

                        <!-- Security Badge -->
                        <div
                            class="mt-4 flex items-center justify-center gap-2 text-xs text-muted-foreground"
                        >
                            <Shield size={14} class="text-green-600" />
                            <span
                                >256-bit SSL encryption • Secure Payment</span
                            >
                        </div>
                    </div>

                    <!-- Warning -->
                    {#if timeRemaining < 300}
                        <div
                            class="flex items-start gap-3 rounded-xl bg-red-50 p-4 text-red-700 dark:bg-red-900/20 dark:text-red-400"
                        >
                            <AlertTriangle size={20} class="mt-0.5 shrink-0" />
                            <div>
                                <p class="text-sm font-medium">
                                    Hold expiring soon!
                                </p>
                                <p class="text-xs text-red-600 dark:text-red-500">
                                    Complete your booking within
                                    {formatTime(timeRemaining)} or seats will be
                                    released.
                                </p>
                            </div>
                        </div>
                    {/if}
                </div>
            </div>
        {:else}
            <div class="flex h-[50vh] items-center justify-center">
                <p>Hold not found</p>
            </div>
        {/if}
    </div>
</div>
