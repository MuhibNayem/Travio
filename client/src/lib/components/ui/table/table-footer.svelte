<script lang="ts">
    import { cn } from "$lib/utils.js";
    import type { WithElementRef } from "$lib/utils.js";
    import type { HTMLAttributes } from "svelte/elements";

    let {
        ref = $bindable(null),
        class: className,
        children,
        ...restProps
    }: WithElementRef<HTMLAttributes<HTMLTableSectionElement>> = $props();
</script>

<tfoot
    bind:this={ref}
    class={cn(
        "bg-muted/50 font-medium text-primary-foreground [&>tr]:last:border-b-0",
        className,
    )}
    {...restProps}
>
    {@render children?.()}
</tfoot>
