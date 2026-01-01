<script lang="ts">
    import { ArrowRight } from "@lucide/svelte";
    const trending = [
        {
            tag: "Music",
            title: "Summer Music Fest 2023",
            month: "Nov",
            day: "20",
            desc: "Experience the biggest open air music festival of the year with top artists.",
            price: "$45.00",
            image: "https://lh3.googleusercontent.com/aida-public/AB6AXuChEQGC60iLbzGy94jeyJy-ORtRVJvramY554bKcS6IAOjSqs1tI-S5eswSMTY4sjt1v2zbJZxWRjkFSlXMNMVSGrcixOBtvtMwSvZnwB4Bsvo6ViGl6gnv-qVqxaWVMRCba-00fS7pEpm5v92Ovc2Tid0WQq4sM5pwI1cy6KFEBW3aDBeP4hKxA3hQ1K-LVtO3KFXlxOPg8XLtAflCLdNFP69gHS4swLXu_pDZTOgfYq7H0o_phvQdLklrLdGRIM13pEt8JUES87sp",
        },
        {
            tag: "Sports",
            title: "National League Final",
            month: "Dec",
            day: "12",
            desc: "The ultimate showdown between the titans. Secure your seat now!",
            price: "$25.00",
            image: "https://lh3.googleusercontent.com/aida-public/AB6AXuBQCkaDSayXnQ078yE3uqa45vepfj5JeRnf4doy4K8C9wt30cmZhPF5AIwbvz1O_OfmRNYn5IGRdZJPd4n0-pLidpiOmYQbVpMiZQgzmvLmAnUuNhugEeWANpugIZgpsKzHHfH65ZCH-ZfFlMS7P89n0tmSuwnDmAdcVkt9ZOtwD7gIcT_lep9tDIl-gPsxGN2Ppsq6UrwQgxXQkKublEg9gktDweD_Qq_uQbdzgkRM9lnAKRzBGqEadiEPDUPbgNrbgBC96i3lOyvT",
        },
        {
            tag: "Arts",
            title: "The Royal Theater Show",
            month: "Jan",
            day: "05",
            desc: "A magical evening of drama and art at the grand theater.",
            price: "$60.00",
            image: "https://lh3.googleusercontent.com/aida-public/AB6AXuBZFw5m3SZz26nFcdLX7vmQEbPjprqOlWmbcv32kGO6Ig9OxExGJ_jcbrndDW4CYWS4cf6GDJNCPxo4WZZjfWTeeC05cQjV361w_4dwUAHmzutlOkLuhyW2L1aKDfFzE_73-1xQFNjW-p9O4EjX0YxGBYFHzepOqbbvP5T1pgMv_U_puF-KhLI1kX9aJyiQH392i1b5Grn2rS3L5sr9G5ZBrs1mj22YOlfe7cKRTqeEnAT1kREYo8-Trrw97eg-FNvalwNWLKMUpYVD",
        },
    ] as const;
</script>

<section class="container mx-auto max-w-[1000px] px-4 py-8 md:px-10">
    <div class="mb-8 flex items-center justify-between">
        <div>
            <h3 class="text-2xl font-bold text-foreground">Trending Events</h3>
            <p class="mt-1 text-sm text-muted-foreground">
                Don't miss out on the hottest tickets
            </p>
        </div>
        <a
            class="flex items-center gap-1 text-sm font-bold text-primary hover:underline hover:text-primary-hover transition-colors"
            href="#"
        >
            View All <ArrowRight size={16} />
        </a>
    </div>

    <div class="grid grid-cols-1 gap-8 sm:grid-cols-2 md:grid-cols-3">
        {#each trending as e (e.title)}
            <div
                class="glass-panel group flex transform flex-col overflow-hidden bg-white/40 dark:bg-white/5 transition-all duration-300 hover:-translate-y-2 hover:shadow-2xl"
            >
                <!-- Image -->
                <div
                    class="relative h-52 overflow-hidden bg-gray-200 dark:bg-gray-800"
                >
                    <div
                        class="absolute inset-0 bg-cover bg-center transition-transform duration-700 group-hover:scale-110"
                        style={`background-image: url(\"${e.image}\")`}
                    ></div>
                    <div
                        class="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-80"
                    ></div>
                    <div
                        class="absolute right-4 top-4 rounded-lg bg-white/90 px-3 py-1.5 text-xs font-bold shadow-sm backdrop-blur dark:bg-black/80 dark:text-white"
                    >
                        {e.tag}
                    </div>
                </div>

                <!-- Content -->
                <div class="flex flex-1 flex-col gap-4 p-5">
                    <div class="flex items-start justify-between gap-2">
                        <h4
                            class="cursor-pointer text-lg font-bold leading-tight transition-colors hover:text-primary"
                        >
                            {e.title}
                        </h4>
                        <div
                            class="min-w-[50px] rounded-lg bg-primary/10 px-2 py-1 text-center backdrop-blur-sm"
                        >
                            <span
                                class="block text-xs font-bold uppercase text-primary"
                                >{e.month}</span
                            >
                            <span
                                class="block text-lg font-black leading-none text-foreground"
                                >{e.day}</span
                            >
                        </div>
                    </div>

                    <p class="line-clamp-2 text-sm text-muted-foreground">
                        {e.desc}
                    </p>

                    <div
                        class="mt-auto flex items-center justify-between border-t border-black/5 pt-4 dark:border-white/5"
                    >
                        <div>
                            <span class="block text-xs text-muted-foreground"
                                >Starting from</span
                            >
                            <span class="text-lg font-bold text-foreground"
                                >{e.price}</span
                            >
                        </div>
                        <button
                            class="rounded-xl bg-primary px-4 py-2 text-sm font-bold text-white shadow-lg shadow-blue-500/30 transition-all hover:bg-primary-hover hover:scale-105 active:scale-95"
                            type="button"
                        >
                            Book
                        </button>
                    </div>
                </div>
            </div>
        {/each}
    </div>
</section>
