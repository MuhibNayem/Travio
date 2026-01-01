<script lang="ts">
    import { Users, Bus, BadgeCheck, Lock } from "@lucide/svelte";

    const stats = [
        { label: "Active Users", value: "1M+", icon: Users },
        { label: "Daily Trips", value: "50k+", icon: Bus },
        { label: "Verified Partners", value: "500+", icon: BadgeCheck },
        { label: "Secure Payments", value: "100%", icon: Lock },
    ];
</script>

<section class="relative z-30 -mt-20 px-4">
    <div class="container mx-auto">
        <div
            class="glass-panel p-6 md:p-8 rounded-3xl bg-white/80 dark:bg-[#1c1c1e]/80 border border-white/20 shadow-2xl backdrop-blur-xl"
        >
            <div class="grid grid-cols-2 md:grid-cols-4 gap-8">
                {#each stats as s}
                    <div
                        class="flex flex-col items-center text-center gap-2 group"
                    >
                        <div
                            class="mb-2 text-[#137fec] group-hover:scale-110 transition-transform duration-300"
                        >
                            <s.icon size={32} />
                        </div>
                        <h3
                            class="text-3xl md:text-4xl font-black text-gray-900 dark:text-white"
                        >
                            {s.value}
                        </h3>
                        <p
                            class="text-sm font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400"
                        >
                            {s.label}
                        </p>
                    </div>
                {/each}
            </div>
        </div>
    </div>
</section>
