<script lang="ts" module>
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type {
		HTMLAnchorAttributes,
		HTMLButtonAttributes,
	} from "svelte/elements";
	import { type VariantProps, tv } from "tailwind-variants";

	export const buttonVariants = tv({
		base: "focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive inline-flex shrink-0 items-center justify-center gap-2 rounded-xl text-sm font-bold whitespace-nowrap transition-all duration-200 outline-none focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 aria-disabled:pointer-events-none aria-disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 active:scale-95",
		variants: {
			variant: {
				default:
					"bg-primary text-primary-foreground hover:bg-primary-hover shadow-lg shadow-blue-500/20 backdrop-blur-sm",
				destructive:
					"bg-destructive hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60 text-white shadow-xs",
				outline:
					"bg-white/50 hover:bg-white/80 text-foreground border border-white/20 dark:bg-white/5 dark:hover:bg-white/10 shadow-sm backdrop-blur-md",
				secondary:
					"bg-secondary text-secondary-foreground hover:bg-secondary/80 shadow-xs",
				ghost: "hover:bg-black/5 hover:text-accent-foreground dark:hover:bg-white/10",
				link: "text-primary underline-offset-4 hover:underline",
			},
			size: {
				default: "h-11 px-6 py-2 has-[>svg]:px-4",
				sm: "h-9 gap-1.5 rounded-lg px-3 has-[>svg]:px-2.5",
				lg: "h-12 rounded-xl px-8 has-[>svg]:px-6 text-base",
				icon: "size-10 rounded-xl",
				"icon-sm": "size-8 rounded-lg",
				"icon-lg": "size-12 rounded-2xl",
			},
		},
		defaultVariants: {
			variant: "default",
			size: "default",
		},
	});

	export type ButtonVariant = VariantProps<typeof buttonVariants>["variant"];
	export type ButtonSize = VariantProps<typeof buttonVariants>["size"];

	export type ButtonProps = WithElementRef<HTMLButtonAttributes> &
		WithElementRef<HTMLAnchorAttributes> & {
			variant?: ButtonVariant;
			size?: ButtonSize;
		};
</script>

<script lang="ts">
	let {
		class: className,
		variant = "default",
		size = "default",
		ref = $bindable(null),
		href = undefined,
		type = "button",
		disabled,
		children,
		...restProps
	}: ButtonProps = $props();
</script>

{#if href}
	<a
		bind:this={ref}
		data-slot="button"
		class={cn(buttonVariants({ variant, size }), className)}
		href={disabled ? undefined : href}
		aria-disabled={disabled}
		role={disabled ? "link" : undefined}
		tabindex={disabled ? -1 : undefined}
		{...restProps}
	>
		{@render children?.()}
	</a>
{:else}
	<button
		bind:this={ref}
		data-slot="button"
		class={cn(buttonVariants({ variant, size }), className)}
		{type}
		{disabled}
		{...restProps}
	>
		{@render children?.()}
	</button>
{/if}
