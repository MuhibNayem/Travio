<script lang="ts">
    import { Ticket } from "@lucide/svelte";
</script>

<footer
    class="border-t border-black/5 bg-white/50 px-4 pb-8 pt-16 backdrop-blur-lg dark:border-white/5 dark:bg-[#0c0c0e]/80 md:px-10"
>
    <div class="mx-auto flex max-w-[1000px] flex-col gap-10">
        <div class="flex flex-col justify-between gap-10 md:flex-row">
            <div class="flex max-w-xs flex-col gap-4">
                <div class="flex items-center gap-2">
                    <div
                        class="flex size-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 text-white"
                    >
                        <Ticket size={20} />
                    </div>
                    <h2 class="text-lg font-bold leading-tight">Travio</h2>
                </div>
                <p
                    class="text-sm leading-relaxed text-gray-500 dark:text-gray-400"
                >
                    The leading platform for all your ticketing needs. Travel,
                    events, and more, all in one place.
                </p>
            </div>

            <div class="flex flex-wrap gap-10 md:gap-20">
                <div class="flex flex-col gap-4">
                    <h5
                        class="text-xs font-bold uppercase tracking-wider text-gray-400"
                    >
                        About
                    </h5>
                    <a
                        class="text-sm text-gray-600 hover:text-primary transition-colors dark:text-gray-400 dark:hover:text-white"
                        href="#">Our Story</a
                    >
                    <a
                        class="text-sm text-gray-600 hover:text-primary transition-colors dark:text-gray-400 dark:hover:text-white"
                        href="#">Careers</a
                    >
                    <a
                        class="text-sm text-gray-600 hover:text-primary transition-colors dark:text-gray-400 dark:hover:text-white"
                        href="#">Blog</a
                    >
                </div>

                <div class="flex flex-col gap-4">
                    <h5
                        class="text-xs font-bold uppercase tracking-wider text-gray-400"
                    >
                        Support
                    </h5>
                    <a
                        class="text-sm text-gray-600 hover:text-primary transition-colors dark:text-gray-400 dark:hover:text-white"
                        href="#">Help Center</a
                    >
                    <a
                        class="text-sm text-gray-600 hover:text-primary transition-colors dark:text-gray-400 dark:hover:text-white"
                        href="#">Privacy Policy</a
                    >
                    <a
                        class="text-sm text-gray-600 hover:text-primary transition-colors dark:text-gray-400 dark:hover:text-white"
                        href="#">Terms of Service</a
                    >
                </div>

                <div class="flex flex-col gap-4">
                    <h5
                        class="text-xs font-bold uppercase tracking-wider text-gray-400"
                    >
                        Social
                    </h5>
                    <div class="flex gap-4">
                        {#each ["facebook", "twitter", "instagram", "linkedin"] as social}
                            <a
                                href="#"
                                class="text-gray-400 transition-colors hover:text-primary hover:scale-110"
                            >
                                <img
                                    src="https://cdn.simpleicons.org/{social}"
                                    alt={social}
                                    class="h-5 w-5 opacity-60 hover:opacity-100 dark:invert transition-opacity duration-200"
                                />
                            </a>
                        {/each}
                    </div>
                </div>
            </div>
        </div>

        <div
            class="flex flex-col items-center border-t border-black/5 pt-8 text-center dark:border-white/5 md:flex-row md:justify-between"
        >
            <p class="text-sm text-gray-400">
                &copy; {new Date().getFullYear()} Travio. All rights reserved.
            </p>
            <p class="text-sm text-gray-400">
                Designed with <span class="text-red-500">♥</span> by Travio Team
            </p>
        </div>
    </div>
</footer>
