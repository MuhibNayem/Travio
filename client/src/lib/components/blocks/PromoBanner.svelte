<script lang="ts">
    import { Copy } from "@lucide/svelte";
    function copyCode() {
        navigator.clipboard?.writeText("SUMMER25");
        // Could add a toast notification here
    }
</script>

<section
    class="border-t border-black/5 bg-white/50 px-4 py-16 backdrop-blur-sm dark:border-white/5 dark:bg-[#0c0c0e]/50 md:px-10"
>
    <div class="mx-auto max-w-[1000px]">
        <div
            class="relative overflow-hidden rounded-3xl bg-gradient-to-r from-blue-700 to-indigo-800 p-8 text-white shadow-2xl md:p-12"
        >
            <!-- Background effects -->
            <div
                class="absolute -right-24 -top-24 h-96 w-96 rounded-full bg-white/10 blur-3xl"
            ></div>
            <div
                class="absolute -bottom-20 left-10 h-60 w-60 rounded-full bg-blue-400/20 blur-2xl"
            ></div>

            <div
                class="relative z-10 flex flex-col items-center justify-between gap-10 md:flex-row"
            >
                <div class="flex-1 text-center md:text-left">
                    <div
                        class="mb-4 inline-block rounded-full border border-white/20 bg-white/20 px-3 py-1 text-xs font-bold uppercase tracking-wider shadow-inner backdrop-blur-md"
                    >
                        Limited Time Offer
                    </div>
                    <h2
                        class="mb-4 text-3xl font-black leading-tight tracking-tight md:text-5xl drop-shadow-lg"
                    >
                        Summer Travel <br /> Mega Sale!
                    </h2>
                    <p class="mb-8 max-w-lg text-lg text-blue-100 font-medium">
                        Plan your nationwide tour now. Get up to <span
                            class="font-bold text-yellow-300">25% OFF</span
                        > on all domestic bus and launch tickets.
                    </p>
                    <button
                        class="rounded-xl bg-white px-10 py-4 text-lg font-bold text-blue-700 shadow-xl shadow-blue-900/40 transition-all hover:-translate-y-1 hover:bg-gray-50 hover:shadow-2xl active:scale-95"
                        type="button"
                    >
                        Grab Offer Now
                    </button>
                </div>

                <!-- Coupon Code Card -->
                <div
                    class="flex-shrink-0 transform rounded-2xl border border-white/30 bg-white/10 p-6 shadow-inner backdrop-blur-xl transition-transform duration-300 hover:rotate-0 hover:scale-105 md:rotate-2"
                >
                    <div class="text-center">
                        <p
                            class="mb-2 text-xs font-bold uppercase tracking-widest text-blue-200"
                        >
                            Coupon Code
                        </p>
                        <div class="flex items-center gap-3">
                            <p
                                class="rounded-lg border-2 border-dashed border-white/40 bg-black/10 px-6 py-3 font-mono text-3xl font-bold tracking-wider shadow-inner"
                            >
                                SUMMER25
                            </p>
                            <button
                                class="rounded-lg bg-white/20 p-3 text-white transition-all hover:bg-white/30 active:scale-90"
                                type="button"
                                aria-label="Copy Code"
                                onclick={copyCode}
                            >
                                <Copy size={24} />
                            </button>
                        </div>
                        <p class="mt-3 text-xs text-blue-200 opacity-80">
                            Valid until 30th November
                        </p>
                    </div>
                </div>
            </div>
        </div>
    </div>
</section>
