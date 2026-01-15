<script lang="ts">
    import * as Command from "$lib/components/ui/command";
    import * as Popover from "$lib/components/ui/popover";
    import * as Button from "$lib/components/ui/button";
    import { Check, ChevronsUpDown, Loader2 } from "@lucide/svelte";
    import { cn } from "$lib/utils";
    import { tick, type Snippet } from "svelte";

    let {
        value = $bindable(""),
        items = [],
        placeholder = "Select item...",
        searchPlaceholder = "Search...",
        emptyText = "No item found.",
        width = "w-[200px]",
        class: className,
        icon,
        loading = false,
        loadingMore = false,
        onSearch,
        onEndReached,
    } = $props<{
        value: string;
        items: { value: string; label: string }[];
        placeholder?: string;
        searchPlaceholder?: string;
        emptyText?: string;
        width?: string;
        class?: string;
        icon?: Snippet;
        loading?: boolean;
        loadingMore?: boolean;
        onSearch?: (term: string) => void;
        onEndReached?: () => void;
    }>();

    let open = $state(false);
    let triggerRef = $state<HTMLButtonElement>(null!);
    let searchQuery = $state("");
    let debounceTimer: any;

    // Find selected label based on value
    let selectedLabel = $derived(
        items.find(
            (item: { value: string; label: string }) => item.value === value,
        )?.label ?? placeholder,
    );
    // Use derived state properly for selectedLabel, but we need to handle when value isn't in items yet.
    // If value exists but items doesn't have it (e.g. initial load), we might want to allow passing a separate display value or just show placeholder.
    // However, sticking to current logic: find in items.

    function closeAndFocusTrigger() {
        open = false;
        tick().then(() => {
            triggerRef?.focus();
        });
    }

    function handleSearch(term: string) {
        searchQuery = term;
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => {
            onSearch?.(term);
        }, 300);
    }

    function handleScroll(e: Event) {
        const target = e.target as HTMLElement;
        const { scrollTop, scrollHeight, clientHeight } = target;

        // Check if we are near the bottom (within 50px)
        if (scrollHeight - scrollTop <= clientHeight + 50) {
            if (!loadingMore && !loading && onEndReached) {
                onEndReached();
            }
        }
    }
</script>

<Popover.Root bind:open>
    <Popover.Trigger bind:ref={triggerRef}>
        {#snippet child({ props })}
            <Button.Root
                variant="glass"
                size="xl"
                role="combobox"
                aria-expanded={open}
                class={cn(
                    "justify-between font-semibold focus:ring-2 focus:ring-primary/20 data-[state=open]:bg-white/90 dark:data-[state=open]:bg-white/20 transition-all font-sans",
                    width,
                    className,
                    !value && "text-muted-foreground",
                )}
                {...props}
            >
                {#if icon}
                    <span class="shrink-0 text-muted-foreground mr-2">
                        {@render icon()}
                    </span>
                {/if}
                <span class="truncate flex-1 text-left">{selectedLabel}</span>
                <ChevronsUpDown class="opacity-50 ml-2 h-4 w-4 shrink-0" />
            </Button.Root>
        {/snippet}
    </Popover.Trigger>
    <Popover.Content class={cn("p-0", width)}>
        <Command.Root shouldFilter={false}>
            <Command.Input
                placeholder={searchPlaceholder}
                value={searchQuery}
                oninput={(e) => handleSearch(e.currentTarget.value)}
            />
            <Command.List>
                <div
                    class="max-h-[300px] overflow-y-auto overflow-x-hidden"
                    onscroll={handleScroll}
                >
                    {#if loading}
                        <div
                            class="py-6 flex items-center justify-center text-sm text-muted-foreground"
                        >
                            <Loader2 class="h-4 w-4 animate-spin mr-2" />
                            Loading...
                        </div>
                    {:else if items.length === 0}
                        <div
                            class="py-6 text-center text-sm text-muted-foreground"
                        >
                            {emptyText}
                        </div>
                    {:else}
                        <Command.Group>
                            {#each items as item (item.value)}
                                <Command.Item
                                    value={item.value}
                                    onSelect={() => {
                                        value = item.value;
                                        closeAndFocusTrigger();
                                    }}
                                >
                                    <Check
                                        class={cn(
                                            "mr-2 h-4 w-4",
                                            value !== item.value &&
                                                "text-transparent",
                                        )}
                                    />
                                    {item.label}
                                </Command.Item>
                            {/each}
                        </Command.Group>

                        {#if loadingMore}
                            <div
                                class="py-2 flex items-center justify-center text-xs text-muted-foreground"
                            >
                                <Loader2 class="h-3 w-3 animate-spin mr-2" />
                                Loading more...
                            </div>
                        {/if}
                    {/if}
                </div>
            </Command.List>
        </Command.Root>
    </Popover.Content>
</Popover.Root>
