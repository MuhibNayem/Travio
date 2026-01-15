<script lang="ts">
    import {
        Bus,
        Train,
        Ship,
        Music,
        Trophy,
        ArrowRight,
    } from "@lucide/svelte";

    import busIcon from "$lib/assets/icons/bus.png";
    import trainIcon from "$lib/assets/icons/train.png";
    import launchIcon from "$lib/assets/icons/launch.png";
    import eventIcon from "$lib/assets/icons/event.png";

    const categories = [
        {
            key: "bus",
            title: "Bus",
            cta: "Book Now",
            icon: Bus,
            image: busIcon,
            gradient: "from-blue-500 to-blue-600",
            chipBg: "bg-white/90",
            arrowBg: "bg-white text-blue-600",
        },
        {
            key: "train",
            title: "Train",
            cta: "Book Now",
            icon: Train,
            image: trainIcon,
            gradient: "from-emerald-500 to-emerald-600",
            chipBg: "bg-white/90",
            arrowBg: "bg-white text-emerald-600",
        },
        {
            key: "launch",
            title: "Launch",
            cta: "Book Now",
            icon: Ship,
            image: launchIcon,
            gradient: "from-cyan-500 to-cyan-600",
            chipBg: "bg-white/90",
            arrowBg: "bg-white text-cyan-600",
        },
        {
            key: "concerts",
            title: "Concerts",
            cta: "Get Tickets",
            icon: Music,
            image: eventIcon,
            gradient: "from-purple-500 to-purple-600",
            chipBg: "bg-white/90",
            arrowBg: "bg-white text-purple-600",
        },
        {
            key: "sports",
            title: "Sports",
            cta: "Get Tickets",
            icon: Trophy,
            // Fallback for now or reuse event? Let's use event for now until dedicated sports icon
            image: eventIcon,
            gradient: "from-orange-500 to-red-600",
            chipBg: "bg-white/90",
            arrowBg: "bg-white text-red-600",
        },
    ] as const;
</script>

<section class="container mx-auto max-w-[1000px] px-4 py-16 md:px-10">
    <div
        class="mb-10 flex flex-col gap-4 md:flex-row md:items-end md:justify-between"
    >
        <div>
            <h3 class="text-3xl font-bold tracking-tight text-foreground">
                Browse Categories
            </h3>
            <p class="mt-2 text-lg text-muted-foreground">
                Select a category to find your next ticket
            </p>
        </div>
    </div>

    <div
        class="grid grid-cols-1 gap-5 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5"
    >
        {#each categories as c (c.key)}
            <a
                href="#"
                class={`group relative flex min-h-[200px] flex-col items-start justify-between overflow-hidden rounded-2xl bg-gradient-to-br ${c.gradient} p-5 text-white shadow-lg transition-all hover:-translate-y-1 hover:shadow-2xl hover:brightness-110 active:scale-95`}
            >
                <!-- Decorative Icon Background -->
                <div
                    class="absolute -right-8 -top-8 size-40 opacity-20 transition-transform duration-500 group-hover:rotate-12 group-hover:scale-110 group-hover:opacity-30 mix-blend-overlay"
                >
                    {#if c.image}
                        <img
                            src={c.image}
                            alt=""
                            class="w-full h-full object-contain"
                        />
                    {:else}
                        <c.icon size={140} strokeWidth={1} />
                    {/if}
                </div>

                <!-- Icon Chip -->
                <div
                    class={`relative z-10 flex size-14 items-center justify-center rounded-xl ${c.chipBg} backdrop-blur-md border border-white/20 shadow-xl transition-transform duration-300 group-hover:scale-110 p-2`}
                >
                    {#if c.image}
                        <img
                            src={c.image}
                            alt=""
                            class="w-full h-full object-contain drop-shadow-sm"
                        />
                    {:else}
                        <c.icon size={32} />
                    {/if}
                </div>

                <div class="relative z-10 mt-6 w-full">
                    <h3 class="mb-1 text-xl font-bold drop-shadow-md">
                        {c.title}
                    </h3>
                    <div
                        class="mt-2 flex items-center justify-between border-t border-white/20 pt-3"
                    >
                        <span
                            class="text-xs font-bold uppercase tracking-wider opacity-90 shadow-black/10 drop-shadow-sm"
                            >{c.cta}</span
                        >
                        <div
                            class={`flex items-center justify-center rounded-full p-1.5 shadow-sm ${c.arrowBg} transition-transform group-hover:translate-x-1`}
                        >
                            <ArrowRight size={16} />
                        </div>
                    </div>
                </div>
            </a>
        {/each}
    </div>
</section>
