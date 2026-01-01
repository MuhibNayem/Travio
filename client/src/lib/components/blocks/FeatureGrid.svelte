<script lang="ts">
    import { ShieldCheck, CircleDollarSign, Headset } from "@lucide/svelte";
</script>

<section
    class="container mx-auto max-w-[1000px] px-4 py-20 text-center md:px-10"
>
    <h3 class="mb-4 text-3xl font-bold text-foreground">
        Why Book With TicketNation?
    </h3>
    <p class="mx-auto mb-16 max-w-2xl text-lg text-muted-foreground">
        We provide the most reliable ticketing service with nationwide coverage.
    </p>

    <div class="grid grid-cols-1 gap-10 md:grid-cols-3">
        <div
            class="glass-panel hover:bg-white dark:hover:bg-white/5 flex flex-col items-center gap-6 p-8 transition-colors duration-300"
        >
            <div
                class="flex size-20 items-center justify-center rounded-2xl bg-green-100 text-green-600 shadow-inner ring-1 ring-green-500/20 transition-transform hover:rotate-0 hover:scale-110 dark:bg-green-900/20 dark:text-green-400 rotate-3"
            >
                <ShieldCheck size={40} />
            </div>
            <div>
                <h4 class="mb-3 text-xl font-bold text-foreground">
                    Secure Payments
                </h4>
                <p class="text-sm leading-relaxed text-muted-foreground">
                    All transactions are secured with military-grade encryption
                    for your peace of mind.
                </p>
            </div>
        </div>

        <div
            class="glass-panel hover:bg-white dark:hover:bg-white/5 flex flex-col items-center gap-6 p-8 transition-colors duration-300"
        >
            <div
                class="flex size-20 items-center justify-center rounded-2xl bg-blue-100 text-blue-600 shadow-inner ring-1 ring-blue-500/20 transition-transform hover:rotate-0 hover:scale-110 dark:bg-blue-900/20 dark:text-blue-400 -rotate-3"
            >
                <CircleDollarSign size={40} />
            </div>
            <div>
                <h4 class="mb-3 text-xl font-bold text-foreground">
                    Instant Refund
                </h4>
                <p class="text-sm leading-relaxed text-muted-foreground">
                    Cancel your booking easily and get instant refunds to your
                    original payment method.
                </p>
            </div>
        </div>

        <div
            class="glass-panel hover:bg-white dark:hover:bg-white/5 flex flex-col items-center gap-6 p-8 transition-colors duration-300"
        >
            <div
                class="flex size-20 items-center justify-center rounded-2xl bg-orange-100 text-orange-600 shadow-inner ring-1 ring-orange-500/20 transition-transform hover:rotate-0 hover:scale-110 dark:bg-orange-900/20 dark:text-orange-400 rotate-3"
            >
                <Headset size={40} />
            </div>
            <div>
                <h4 class="mb-3 text-xl font-bold text-foreground">
                    24/7 Support
                </h4>
                <p class="text-sm leading-relaxed text-muted-foreground">
                    Our dedicated support team is available round the clock to
                    assist you with any issues.
                </p>
            </div>
        </div>
    </div>
</section>
