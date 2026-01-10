<script lang="ts">
    import { cn } from "$lib/utils";
    import { createEventDispatcher } from "svelte";
    import * as Tooltip from "$lib/components/ui/tooltip";

    export let id: string;
    export let label: string;
    export let status:
        | "available"
        | "selected"
        | "sold"
        | "blocked"
        | "female_only" = "available";
    export let category: string = "Economy";
    export let price: number = 0;
    export let className: string = "";

    const dispatch = createEventDispatcher();

    function handleClick() {
        if (status === "available" || status === "selected") {
            dispatch("click", { id, status });
        }
    }

    // Status colors
    $: statusColor = {
        available:
            "bg-white border-gray-300 hover:border-primary hover:text-primary cursor-pointer",
        selected:
            "bg-primary text-primary-foreground border-primary cursor-pointer",
        sold: "bg-gray-200 text-gray-400 border-gray-200 cursor-not-allowed",
        blocked: "bg-red-100 text-red-300 border-red-200 cursor-not-allowed",
        female_only:
            "bg-pink-50 border-pink-200 text-pink-500 hover:border-pink-400 cursor-pointer",
    }[status];
</script>

<Tooltip.Root>
    <Tooltip.Trigger>
        <button
            type="button"
            class={cn(
                "relative flex h-10 w-10 items-center justify-center rounded-md border transition-all duration-200",
                "text-xs font-medium shadow-sm",
                statusColor,
                className,
            )}
            on:click={handleClick}
            disabled={status === "sold" || status === "blocked"}
        >
            {label}

            <!-- Top curve for seat visual -->
            <div
                class="absolute -top-1 left-1.5 right-1.5 h-1 rounded-t-sm border-t border-x"
                class:border-gray-300={status === "available"}
                class:border-primary={status === "selected"}
                class:bg-primary={status === "selected"}
                class:bg-white={status === "available"}
                class:border-gray-200={status === "sold"}
                class:bg-gray-200={status === "sold"}
            ></div>
        </button>
    </Tooltip.Trigger>
    <Tooltip.Content>
        <div class="text-xs">
            <p class="font-semibold">{category}</p>
            <p>Seat {label}</p>
            <p>৳{price}</p>
        </div>
    </Tooltip.Content>
</Tooltip.Root>
