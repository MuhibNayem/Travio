<script lang="ts">
	import "../app.css";
	import "./layout.css";
	import favicon from "$lib/assets/favicon.svg";
	import { Toaster } from "svelte-sonner";

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
<Toaster
	position="top-right"
	richColors
	closeButton
	expand={true}
	duration={4000}
/>
{@render children()}
