<script lang="ts">
    const partners = [
        "Shohoz",
        "BusBD",
        "Railway",
        "LaunchBD",
        "Uber",
        "Pathao",
        "Emirates",
        "Biman",
    ];
</script>

<section class="py-20 overflow-hidden bg-gray-50 dark:bg-black/50">
    <div class="container mx-auto px-4 mb-10 text-center">
        <p
            class="text-sm font-bold uppercase tracking-widest text-[#137fec] mb-2"
        >
            Our Network
        </p>
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white opacity-80">
            Trusted by leading operators nationwide
        </h2>
    </div>

    <!-- Infinite Scroll Marquee -->
    <div class="relative flex overflow-x-hidden group">
        <div class="animate-marquee whitespace-nowrap flex gap-16 items-center">
            {#each [...partners, ...partners, ...partners] as p}
                <div
                    class="text-2xl md:text-3xl font-black text-gray-300 hover:text-gray-400 dark:text-gray-700 dark:hover:text-gray-600 transition-colors uppercase select-none"
                >
                    {p}
                </div>
            {/each}
        </div>

        <div
            class="animate-marquee2 absolute top-0 whitespace-nowrap flex gap-16 items-center"
        >
            {#each [...partners, ...partners, ...partners] as p}
                <div
                    class="text-2xl md:text-3xl font-black text-gray-300 hover:text-gray-400 dark:text-gray-700 dark:hover:text-gray-600 transition-colors uppercase select-none"
                >
                    {p}
                </div>
            {/each}
        </div>
    </div>
</section>

<style>
    .animate-marquee {
        animation: marquee 25s linear infinite;
    }
    .animate-marquee2 {
        animation: marquee2 25s linear infinite;
    }

    @keyframes marquee {
        0% {
            transform: translateX(0%);
        }
        100% {
            transform: translateX(-100%);
        }
    }
    @keyframes marquee2 {
        0% {
            transform: translateX(100%);
        }
        100% {
            transform: translateX(0%);
        }
    }
</style>
