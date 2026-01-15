<script lang="ts">
    import { page } from "$app/stores";
    import {
        LayoutDashboard,
        Bus,
        Ticket,
        Receipt,
        Settings,
        CalendarDays,
        Users,
        Route,
        Building,
        ChevronDown,
        Music,
        MapPin,
        Plus,
    } from "@lucide/svelte";
    import { slide } from "svelte/transition";
    import { cn } from "$lib/utils";

    const navItems = [
        {
            title: "Overview",
            href: "/organization",
            icon: LayoutDashboard,
        },
        {
            title: "Operations",
            icon: Bus,
            submenu: [
                {
                    title: "Routes",
                    href: "/organization/operations/routes",
                    icon: Route,
                },
                {
                    title: "Trips",
                    href: "/organization/operations/trips",
                    icon: CalendarDays,
                },
                {
                    title: "Fleet",
                    href: "/organization/operations/fleet",
                    icon: Bus,
                },
            ],
        },
        {
            title: "Events",
            icon: Music,
            submenu: [
                {
                    title: "Dashboard",
                    href: "/organization/events",
                    icon: LayoutDashboard,
                },
                {
                    title: "Create Event",
                    href: "/organization/events/create",
                    icon: Plus,
                },
                {
                    title: "Venues",
                    href: "/organization/events/venues",
                    icon: MapPin,
                },
            ],
        },

        {
            title: "Sales (Counter)",
            href: "/organization/sales/counter",
            icon: Ticket,
        },
        {
            title: "Finance",
            href: "/organization/finance",
            icon: Receipt,
        },
        {
            title: "Settings",
            href: "/organization/settings",
            icon: Settings,
            submenu: [
                {
                    title: "Organization",
                    href: "/organization/settings",
                    icon: Building,
                },
                {
                    title: "Members",
                    href: "/organization/members",
                    icon: Users,
                },
            ],
        },
    ];

    let expanded = $state<Record<string, boolean>>({
        Operations: true, // Default open
        Events: false,
        Settings: false,
    });

    function toggle(title: string) {
        expanded[title] = !expanded[title];
    }
</script>

<aside
    class="fixed left-0 top-16 z-30 h-[calc(100vh-4rem)] w-64 border-r border-white/20 bg-white/50 backdrop-blur-xl transition-transform dark:border-white/5 dark:bg-[#101922]/50 md:translate-x-0"
>
    <div class="h-full overflow-y-auto px-3 py-4">
        <ul class="space-y-2 font-medium">
            {#each navItems as item}
                {#if item.submenu}
                    <!-- Submenu Header -->
                    <li>
                        <button
                            type="button"
                            class="flex w-full items-center rounded-lg px-3 py-2 text-left text-gray-900 transition-colors hover:bg-gray-100 dark:text-white dark:hover:bg-gray-700"
                            onclick={() => toggle(item.title)}
                        >
                            <item.icon
                                class="h-5 w-5 flex-shrink-0 text-gray-500 transition duration-75 dark:text-gray-400"
                            />
                            <span class="ml-3 flex-1 whitespace-nowrap"
                                >{item.title}</span
                            >
                            <ChevronDown
                                class={cn(
                                    "h-4 w-4 transition-transform duration-200",
                                    expanded[item.title] ? "rotate-180" : "",
                                )}
                            />
                        </button>
                        {#if expanded[item.title]}
                            <ul
                                transition:slide={{ duration: 200 }}
                                class="ml-6 mt-1 space-y-1 border-l border-gray-200 dark:border-gray-700"
                            >
                                {#each item.submenu as subItem}
                                    <li>
                                        <a
                                            href={subItem.href}
                                            class={cn(
                                                "group flex w-full items-center rounded-r-lg border-l-2 border-transparent px-3 py-2 pl-4 text-sm transition-all hover:bg-gray-100 dark:hover:bg-gray-700",
                                                $page.url.pathname ===
                                                    subItem.href
                                                    ? "border-primary bg-primary/10 text-primary dark:text-primary-400"
                                                    : "text-gray-500 hover:border-gray-300 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white",
                                            )}
                                        >
                                            <subItem.icon
                                                class="mr-2 h-4 w-4"
                                            />
                                            {subItem.title}
                                        </a>
                                    </li>
                                {/each}
                            </ul>
                        {/if}
                    </li>
                {:else}
                    <!-- Single Link -->
                    <li>
                        <a
                            href={item.href}
                            class={cn(
                                "group flex items-center rounded-lg border-l-4 border-transparent px-3 py-2 text-gray-900 hover:bg-gray-100 dark:text-white dark:hover:bg-gray-700",
                                $page.url.pathname === item.href ||
                                    ($page.url.pathname.startsWith(item.href) &&
                                        item.href !== "/organization")
                                    ? "border-primary bg-primary/10 text-primary"
                                    : "",
                            )}
                        >
                            <item.icon
                                class={cn(
                                    "h-5 w-5 flex-shrink-0 transition duration-75 group-hover:text-gray-900 dark:group-hover:text-white",
                                    $page.url.pathname === item.href
                                        ? "text-primary"
                                        : "text-gray-500 dark:text-gray-400",
                                )}
                            />
                            <span class="ml-3 flex-1 whitespace-nowrap"
                                >{item.title}</span
                            >
                        </a>
                    </li>
                {/if}
            {/each}
        </ul>
    </div>
</aside>
