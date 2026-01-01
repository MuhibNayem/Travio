<script lang="ts">
    import * as Command from "$lib/components/ui/command";
    import * as Popover from "$lib/components/ui/popover";
    import { Button } from "$lib/components/ui/button";
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
        onInput,
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
        onInput?: (value: string) => void;
    }>();

    let open = $state(false);
    let triggerRef = $state<HTMLButtonElement>(null!);

    // Find selected label based on value
    // We use a derived value to ensure it updates when value or items change
    let selectedLabel = $derived(
        items.find((item) => item.value === value)?.label ?? placeholder,
    );

    function closeAndFocusTrigger() {
        open = false;
        tick().then(() => {
            triggerRef?.focus();
        });
    }
</script>

<Popover.Root bind:open>
    <Popover.Trigger bind:ref={triggerRef}>
        {#snippet child({ props })}
            <div class="relative w-full">
                {#if icon}
                    <div
                        class="absolute left-4 top-1/2 -translate-y-1/2 z-10 text-muted-foreground pointer-events-none"
                    >
                        {@render icon()}
                    </div>
                {/if}
                <Button
                    variant="ghost"
                    role="combobox"
                    aria-expanded={open}
                    class={cn(
                        "justify-between h-14 rounded-xl border-transparent bg-black/5 text-base font-semibold placeholder:text-muted-foreground hover:bg-black/10 focus:bg-white focus:ring-2 focus:ring-primary/20 data-[state=open]:bg-white dark:bg-white/5 dark:hover:bg-white/10 dark:focus:bg-white/10 dark:data-[state=open]:bg-white/10 transition-all font-sans",
                        width,
                        icon ? "pl-12" : "px-4",
                        className,
                    )}
                    {...props}
                >
                    {selectedLabel}
                    <ChevronsUpDown class="opacity-50 ml-2 h-4 w-4 shrink-0" />
                </Button>
            </div>
        {/snippet}
    </Popover.Trigger>
    <Popover.Content class={cn("p-0", width)}>
        <Command.Root>
            <Command.Input
                placeholder={searchPlaceholder}
                oninput={(e) => onInput?.(e.currentTarget.value)}
            />
            <Command.List>
                {#if loading}
                    <div
                        class="py-6 flex items-center justify-center text-sm text-muted-foreground"
                    >
                        <Loader2 class="h-4 w-4 animate-spin mr-2" />
                        Loading...
                    </div>
                {:else}
                    <Command.Empty>{emptyText}</Command.Empty>
                    <Command.Group>
                        {#each items as item (item.value)}
                            <Command.Item
                                value={item.label}
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
                {/if}
            </Command.List>
        </Command.Root>
    </Popover.Content>
</Popover.Root>
