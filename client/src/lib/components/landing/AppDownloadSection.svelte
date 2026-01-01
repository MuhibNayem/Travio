<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Ticket, Apple, Play } from "@lucide/svelte";
</script>

<section class="py-24 px-4 bg-[#101922] text-white relative overflow-hidden">
    <!-- Decorative Blurs -->
    <div
        class="absolute top-0 right-0 w-[500px] h-[500px] bg-blue-500/20 blur-[120px] rounded-full"
    ></div>
    <div
        class="absolute bottom-0 left-0 w-[500px] h-[500px] bg-purple-500/20 blur-[120px] rounded-full"
    ></div>

    <div class="container mx-auto relative z-10">
        <div class="flex flex-col lg:flex-row items-center gap-16">
            <!-- Content -->
            <div class="flex-1 text-center lg:text-left">
                <span
                    class="inline-block px-4 py-1.5 rounded-full border border-blue-400/30 bg-blue-500/10 text-blue-400 text-sm font-bold mb-6"
                >
                    Mobile App Coming Soon
                </span>
                <h2 class="text-4xl md:text-6xl font-black mb-6 leading-tight">
                    Travel Smart. <br />
                    <span
                        class="text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400"
                        >Download the App.</span
                    >
                </h2>
                <p class="text-xl text-gray-400 mb-10 max-w-xl mx-auto lg:mx-0">
                    Get real-time updates, exclusive app-only deals, and manage
                    all your tickets in one secure place.
                </p>

                <div
                    class="flex flex-col sm:flex-row gap-4 justify-center lg:justify-start"
                >
                    <button
                        class="flex items-center gap-3 bg-white text-black px-6 py-3 rounded-xl hover:bg-gray-100 transition-colors"
                    >
                        <Apple size={30} fill="currentColor" />
                        <div class="text-left">
                            <div class="text-xs font-semibold">
                                Download on the
                            </div>
                            <div class="text-xl font-bold leading-none">
                                App Store
                            </div>
                        </div>
                    </button>
                    <button
                        class="flex items-center gap-3 bg-transparent border border-white/20 text-white px-6 py-3 rounded-xl hover:bg-white/10 transition-colors"
                    >
                        <Play size={30} fill="currentColor" />
                        <div class="text-left">
                            <div class="text-xs font-semibold">GET IT ON</div>
                            <div class="text-xl font-bold leading-none">
                                Google Play
                            </div>
                        </div>
                    </button>
                </div>
            </div>

            <!-- Visual -->
            <div class="flex-1 relative">
                <!-- Simple CSS Phone Mockup -->
                <div
                    class="relative mx-auto border-gray-800 bg-gray-800 border-[14px] rounded-[2.5rem] h-[600px] w-[300px] shadow-xl"
                >
                    <div
                        class="h-[32px] w-[3px] bg-gray-800 absolute -start-[17px] top-[72px] rounded-s-lg"
                    ></div>
                    <div
                        class="h-[46px] w-[3px] bg-gray-800 absolute -start-[17px] top-[124px] rounded-s-lg"
                    ></div>
                    <div
                        class="h-[46px] w-[3px] bg-gray-800 absolute -start-[17px] top-[178px] rounded-s-lg"
                    ></div>
                    <div
                        class="h-[64px] w-[3px] bg-gray-800 absolute -end-[17px] top-[142px] rounded-e-lg"
                    ></div>
                    <div
                        class="rounded-[2rem] overflow-hidden w-full h-full bg-white dark:bg-[#101922]"
                    >
                        <!-- Screen Content -->
                        <div
                            class="relative w-full h-full bg-[url('https://images.unsplash.com/photo-1544620347-c4fd4a3d5957?q=80&w=987&auto=format&fit=crop')] bg-cover bg-center"
                        >
                            <div
                                class="absolute inset-0 bg-black/40 flex items-center justify-center"
                            >
                                <div class="text-center">
                                    <div
                                        class="size-16 bg-white/20 backdrop-blur-md rounded-2xl mx-auto mb-4 flex items-center justify-center"
                                    >
                                        <Ticket size={32} class="text-white" />
                                    </div>
                                    <h3 class="text-white font-bold text-2xl">
                                        TicketNation
                                    </h3>
                                </div>
                            </div>

                            <!-- Bottom Nav Mockup -->
                            <div
                                class="absolute bottom-0 left-0 w-full h-20 bg-white/10 backdrop-blur-md border-t border-white/10 flex justify-around items-center px-4"
                            >
                                <div
                                    class="size-8 rounded-full bg-white/20"
                                ></div>
                                <div
                                    class="size-8 rounded-full bg-white/20"
                                ></div>
                                <div
                                    class="size-12 rounded-full bg-[#137fec] shadow-lg"
                                ></div>
                                <div
                                    class="size-8 rounded-full bg-white/20"
                                ></div>
                                <div
                                    class="size-8 rounded-full bg-white/20"
                                ></div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</section>
