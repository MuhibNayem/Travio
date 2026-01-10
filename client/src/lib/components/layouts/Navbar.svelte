<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { auth } from "$lib/runes/auth.svelte";
    import { goto } from "$app/navigation";
    import { Ticket, Sun, Moon, Menu } from "@lucide/svelte";
    import { toast } from "svelte-sonner";

    let theme = $state<"light" | "dark">("light");

    function toggleTheme() {
        theme = theme === "light" ? "dark" : "light";
        if (typeof document !== "undefined") {
            document.documentElement.classList.toggle("dark", theme === "dark");
        }
    }

    async function handleLogout() {
        await auth.logout();
        toast.success("Signed out", {
            description: "You have been logged out successfully.",
        });
        goto("/login");
    }

    $effect(() => {
        // Sync theme on mount
        if (typeof document !== "undefined") {
            if (document.documentElement.classList.contains("dark")) {
                theme = "dark";
            }
        }
    });
</script>

<header
    class="sticky top-0 z-50 flex items-center justify-between whitespace-nowrap border-b border-white/20 bg-white/80 px-4 py-3 shadow-sm backdrop-blur-xl transition-all dark:border-white/5 dark:bg-[#101922]/80 md:px-10"
>
    <!-- Logo -->
    <a href="/" class="flex cursor-pointer items-center gap-4">
        <div
            class="flex size-9 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-md transition-transform hover:scale-105"
        >
            <Ticket size={20} />
        </div>
        <h2
            class="text-xl font-extrabold leading-tight tracking-tight text-gradient"
        >
            TicketNation
        </h2>
    </a>

    <!-- Desktop Nav -->
    <div class="hidden flex-1 justify-end gap-8 md:flex">
        <nav class="flex items-center gap-8">
            {#each ["Transport", "Events", "Sports", "Support"] as item}
                <a
                    class="text-sm font-semibold text-gray-600 hover:text-primary transition-colors dark:text-gray-300 dark:hover:text-white"
                    href="#">{item}</a
                >
            {/each}
        </nav>

        <div class="flex gap-3">
            {#if auth.isAuthenticated}
                {#if auth.user?.role === "admin" || auth.user?.role === "operator"}
                    <Button
                        variant="ghost"
                        class="h-10 px-4 font-bold text-gray-700 hover:text-primary dark:text-gray-300 dark:hover:text-white"
                        href="/organization"
                    >
                        Organization
                    </Button>
                {:else}
                    <Button
                        variant="ghost"
                        class="h-10 px-4 font-bold text-gray-700 hover:text-primary dark:text-gray-300 dark:hover:text-white"
                        href="/dashboard"
                    >
                        Dashboard
                    </Button>
                {/if}

                <Button
                    variant="outline"
                    class="glass-button h-10 border-transparent bg-transparent hover:bg-black/5 dark:hover:bg-white/10 text-foreground dark:text-white"
                    onclick={handleLogout}
                >
                    Sign Out
                </Button>
                <!-- Avatar placeholder could go here -->
                <div
                    class="flex size-10 items-center justify-center rounded-full bg-gradient-to-br from-indigo-500 to-purple-500 text-white font-bold shadow-md"
                >
                    {auth.user?.name.charAt(0).toUpperCase()}
                </div>
            {:else}
                <Button
                    variant="outline"
                    class="glass-button h-10 border-transparent bg-transparent hover:bg-black/5 dark:hover:bg-white/10 text-foreground dark:text-white"
                    href="/login"
                >
                    Sign In
                </Button>
                <Button
                    class="h-10 rounded-xl bg-primary px-5 font-bold text-white shadow-lg shadow-blue-500/20 hover:bg-primary-hover active:scale-95 transition-all"
                    href="/register"
                >
                    Register
                </Button>
            {/if}

            <Button
                variant="ghost"
                class="h-10 w-10 p-0 rounded-full hover:bg-black/5 dark:hover:bg-white/10"
                title="Toggle theme"
                onclick={toggleTheme}
            >
                {#if theme === "dark"}
                    <Sun
                        size={22}
                        class="transition-transform duration-500 rotate-0"
                    />
                {:else}
                    <Moon
                        size={22}
                        class="transition-transform duration-500 rotate-0"
                    />
                {/if}
            </Button>
        </div>
    </div>

    <!-- Mobile Menu -->
    <div class="md:hidden">
        <Button variant="ghost" size="icon">
            <Menu size={28} />
        </Button>
    </div>
</header>
