<script lang="ts">
    import { Input } from "$lib/components/ui/input";
    import {
        Tabs,
        TabsList,
        TabsTrigger,
        TabsContent,
    } from "$lib/components/ui/tabs";
    import {
        Train,
        Bus,
        Ship,
        Ticket,
        Calendar,
        Search,
        ChevronsUpDown,
        Check,
        MapPin,
    } from "@lucide/svelte";
    import { goto } from "$app/navigation";
    import { STATIONS } from "$lib/mocks/data";
    import { tick } from "svelte";
    import { Combobox } from "$lib/components/ui/combobox";

    let tab = $state<"train" | "bus" | "launch" | "events">("train");
    let from = $state("");
    let to = $state("");
    let date = $state("");

    let openFrom = $state(false);
    let openTo = $state(false);

    // Trigger refs to refocus after selection
    let fromTriggerRef = $state<HTMLButtonElement>(null!);
    let toTriggerRef = $state<HTMLButtonElement>(null!);

    const heroBg =
        "https://lh3.googleusercontent.com/aida-public/AB6AXuDx0xRWj6lZ2pXRY_kVSNDWjvTO05IfP0WOg4aQulSWF6Q9jZn0OJ4JbT3ukEiAvY04yGZVb3oSQ-LK6U1zzZjjhifzxKZwVV2dsXoaXGAeNEqHIO4Y-PtQnsgAqnsKwWfdLr0c9RE0J4rOHxfMJG706YLzHOHYd6SbUnTsOg1X5xBNbsuRrtRCgGWWxp5i59FsYE7pegcXR_bdNU7X9UR1q01EvZ3p941bUdcAbEmIpjMRovkRmZceZk9zH-ULCv3dWJoa5TKcLbCO";

    function onSearch() {
        if (!from || !to) return;
        goto(`/search?from=${from}&to=${to}&type=${tab}&date=${date}`);
    }

    function closeAndFocusTrigger(trigger: "from" | "to") {
        if (trigger === "from") {
            openFrom = false;
            tick().then(() => fromTriggerRef?.focus());
        } else {
            openTo = false;
            tick().then(() => toTriggerRef?.focus());
        }
    }
</script>

<section
    class="relative flex min-h-[650px] w-full items-center justify-center px-4 pb-24 pt-32 dark:text-white"
>
    <!-- Background Image with Overlay -->
    <div class="absolute inset-0 z-0 overflow-hidden">
        <div
            class="h-full w-full bg-cover bg-center transition-transform duration-[20s] hover:scale-105"
            style={`background-image: url(\"${heroBg}\")`}
        ></div>
        <div
            class="absolute inset-0 bg-gradient-to-b from-white/80 via-white/40 to-white/90 dark:from-black/80 dark:via-black/40 dark:to-black/90"
        ></div>
    </div>

    <div
        class="relative z-10 flex w-full max-w-[1000px] flex-col items-center gap-10"
    >
        <div class="mx-auto flex max-w-4xl flex-col gap-6 px-4 text-center">
            <div
                class="mx-auto inline-flex items-center justify-center gap-2 rounded-full border border-black/5 bg-white/30 px-4 py-1.5 text-xs font-bold uppercase tracking-wider text-black backdrop-blur-md dark:border-white/10 dark:bg-black/30 dark:text-white"
            >
                <span
                    class="h-2 w-2 animate-pulse rounded-full bg-green-500 shadow-[0_0_10px_theme(colors.green.500)]"
                ></span>
                Nationwide Coverage
            </div>

            <h1
                class="text-5xl font-black leading-[1.1] tracking-tight text-foreground drop-shadow-sm md:text-7xl"
            >
                Tickets for Every
                <br class="hidden md:block" />
                <span
                    class="bg-gradient-to-r from-blue-600 to-indigo-500 bg-clip-text text-transparent dark:from-blue-400 dark:to-indigo-300"
                >
                    Journey &amp; Experience
                </span>
            </h1>

            <p
                class="mx-auto max-w-2xl text-lg font-medium leading-relaxed text-muted-foreground drop-shadow-sm md:text-xl"
            >
                Seamlessly book intercity trains, buses, launches, concerts, and
                sports matches across the country.
            </p>
        </div>

        <!-- Search Card -->
        <div
            class="glass-panel w-full overflow-visible p-1 shadow-2xl ring-1 ring-black/5 dark:ring-white/10"
        >
            <Tabs
                value={tab}
                onValueChange={(v) => {
                    tab = v as any;
                }}
            >
                <div
                    class="flex overflow-x-auto border-b border-black/5 p-2 dark:border-white/5"
                >
                    <TabsList
                        class="h-auto w-full justify-between gap-2 bg-transparent p-0"
                    >
                        <TabsTrigger
                            value="train"
                            class="flex-1 gap-2 rounded-xl py-3 data-[state=active]:bg-white/80 data-[state=active]:text-primary data-[state=active]:shadow-sm data-[state=active]:backdrop-blur-md dark:data-[state=active]:bg-white/10 dark:data-[state=active]:text-white"
                        >
                            <Train size={24} />
                            <span class="text-sm font-bold">Train</span>
                        </TabsTrigger>
                        <TabsTrigger
                            value="bus"
                            class="flex-1 gap-2 rounded-xl py-3 text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 data-[state=active]:bg-white/80 data-[state=active]:text-primary data-[state=active]:shadow-sm data-[state=active]:backdrop-blur-md dark:data-[state=active]:bg-white/10 dark:data-[state=active]:text-white"
                        >
                            <Bus size={24} />
                            <span class="text-sm font-bold">Bus</span>
                        </TabsTrigger>
                        <TabsTrigger
                            value="launch"
                            class="flex-1 gap-2 rounded-xl py-3 text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 data-[state=active]:bg-white/80 data-[state=active]:text-primary data-[state=active]:shadow-sm data-[state=active]:backdrop-blur-md dark:data-[state=active]:bg-white/10 dark:data-[state=active]:text-white"
                        >
                            <Ship size={24} />
                            <span class="text-sm font-bold">Launch</span>
                        </TabsTrigger>
                        <TabsTrigger
                            value="events"
                            class="flex-1 gap-2 rounded-xl py-3 text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 data-[state=active]:bg-white/80 data-[state=active]:text-primary data-[state=active]:shadow-sm data-[state=active]:backdrop-blur-md dark:data-[state=active]:bg-white/10 dark:data-[state=active]:text-white"
                        >
                            <Ticket size={24} />
                            <span class="text-sm font-bold">Events</span>
                        </TabsTrigger>
                    </TabsList>
                </div>

                {#each ["train", "bus", "launch", "events"] as t (t)}
                    <TabsContent
                        value={t}
                        class="m-0 animate-in fade-in slide-in-from-bottom-4 duration-500"
                    >
                        <div
                            class="grid grid-cols-1 gap-5 p-6 md:grid-cols-2 md:p-8 lg:grid-cols-4"
                        >
                            <!-- From -->
                            <div class="group flex flex-col gap-2">
                                <label
                                    class="text-xs font-bold uppercase tracking-wide text-muted-foreground"
                                    >From</label
                                >
                                <Combobox
                                    items={STATIONS.map((s) => ({
                                        value: s.id,
                                        label: s.name,
                                    }))}
                                    bind:value={from}
                                    placeholder="Select origin..."
                                    searchPlaceholder="Search station..."
                                    emptyText="No station found."
                                    width="w-full"
                                    class="text-base font-semibold"
                                >
                                    {#snippet icon()}
                                        <MapPin size={24} />
                                    {/snippet}
                                </Combobox>
                            </div>

                            <!-- To -->
                            <div class="group flex flex-col gap-2">
                                <label
                                    class="text-xs font-bold uppercase tracking-wide text-muted-foreground"
                                    >To</label
                                >
                                <Combobox
                                    items={STATIONS.map((s) => ({
                                        value: s.id,
                                        label: s.name,
                                    }))}
                                    bind:value={to}
                                    placeholder="Select destination..."
                                    searchPlaceholder="Search station..."
                                    emptyText="No station found."
                                    width="w-full"
                                    class="text-base font-semibold"
                                >
                                    {#snippet icon()}
                                        <MapPin size={24} />
                                    {/snippet}
                                </Combobox>
                            </div>

                            <!-- Date -->
                            <div class="group relative flex flex-col gap-2">
                                <label
                                    class="text-xs font-bold uppercase tracking-wide text-muted-foreground"
                                    >Date</label
                                >
                                <div class="relative flex items-center">
                                    <div
                                        class="absolute left-4 z-10 text-muted-foreground pointer-events-none"
                                    >
                                        <Calendar size={24} />
                                    </div>
                                    <Input
                                        class="!h-14 rounded-xl border border-white/20 bg-white/60 backdrop-blur-md pl-12 text-base font-semibold placeholder:text-muted-foreground hover:bg-white/80 focus:bg-white/90 focus:ring-2 focus:ring-primary/20 dark:border-white/10 dark:bg-white/10 dark:hover:bg-white/15 dark:focus:bg-white/20 transition-all shadow-sm"
                                        placeholder="Select Date"
                                        type="date"
                                        bind:value={date}
                                    />
                                </div>
                            </div>

                            <!-- Search Button -->
                            <div class="flex flex-col gap-2">
                                <label
                                    class="text-xs font-bold uppercase tracking-wide text-transparent select-none"
                                    >Search</label
                                >
                                <button
                                    class="flex h-14 w-full items-center justify-center gap-2 rounded-xl bg-primary text-lg font-bold text-white shadow-lg shadow-blue-500/30 transition-all hover:scale-[1.02] hover:bg-primary-hover hover:shadow-xl active:scale-95"
                                    type="button"
                                    onclick={onSearch}
                                >
                                    <Search size={24} />
                                    Search
                                </button>
                            </div>
                        </div>
                    </TabsContent>
                {/each}
            </Tabs>
        </div>
    </div>
</section>
