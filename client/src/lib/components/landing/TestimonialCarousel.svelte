<script lang="ts">
    const testimonials = [
        {
            name: "Sarah Ahmed",
            role: "Frequent Traveler",
            text: "TicketNation changed how I travel. The booking process is buttery smooth and the app is gorgeous.",
            avatar: "S",
        },
        {
            name: "Rahim Chowdhury",
            role: "Business Commuter",
            text: "I use this every week for my Dhaka-Ctg trips. Reliable, fast, and secure. Highly recommended!",
            avatar: "R",
        },
        {
            name: "Emily Blunt",
            role: "Tourist",
            text: "Found the best launch cabin prices here. The virtual tour feature for cabins is amazing.",
            avatar: "E",
        },
    ];
</script>

<section class="py-24 px-4 relative overflow-hidden">
    <div
        class="absolute inset-0 bg-gradient-to-b from-transparent to-blue-50/50 dark:to-blue-900/10 pointer-events-none"
    ></div>

    <div class="container mx-auto relative z-10">
        <div class="text-center mb-16">
            <h2
                class="text-4xl md:text-5xl font-black text-gray-900 dark:text-white mb-6"
            >
                Loved by Travelers
            </h2>
            <p
                class="max-w-xl mx-auto text-lg text-gray-500 dark:text-gray-400"
            >
                Don't just take our word for it. Here's what our community has
                to say.
            </p>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
            {#each testimonials as t}
                <div
                    class="glass-panel p-8 rounded-3xl relative hover:-translate-y-2 transition-transform duration-300"
                >
                    <span
                        class="absolute top-6 right-8 text-6xl text-blue-100 dark:text-blue-900 font-serif opacity-50"
                        >"</span
                    >

                    <p
                        class="text-lg text-gray-600 dark:text-gray-300 mb-8 leading-relaxed relative z-10"
                    >
                        {t.text}
                    </p>

                    <div class="flex items-center gap-4">
                        <div
                            class="size-12 rounded-full bg-gradient-to-tr from-blue-500 to-purple-500 flex items-center justify-center text-white font-bold text-xl shadow-lg"
                        >
                            {t.avatar}
                        </div>
                        <div>
                            <h4 class="font-bold text-gray-900 dark:text-white">
                                {t.name}
                            </h4>
                            <p class="text-sm text-gray-500 dark:text-gray-400">
                                {t.role}
                            </p>
                        </div>
                    </div>
                </div>
            {/each}
        </div>
    </div>
</section>
