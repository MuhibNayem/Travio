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
    let selectedLabel = $derived(
        items.find(
            (item: { value: string; label: string }) => item.value === value,
        )?.label ?? placeholder,
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
            <Button.Root
                variant="glass"
                size="xl"
                role="combobox"
                aria-expanded={open}
                class={cn(
                    "justify-between font-semibold focus:ring-2 focus:ring-primary/20 data-[state=open]:bg-white/90 dark:data-[state=open]:bg-white/20 transition-all font-sans",
                    width,
                    className,
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
