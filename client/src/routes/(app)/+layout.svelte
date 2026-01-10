<script lang="ts">
    import { goto } from "$app/navigation";
    import { auth } from "$lib/runes/auth.svelte";
    import Navbar from "$lib/components/layouts/Navbar.svelte";
    import Footer from "$lib/components/layouts/Footer.svelte";

    let { children } = $props();

    $effect(() => {
        if (!auth.isLoading && !auth.isAuthenticated) {
            goto("/login");
        }
    });
</script>

{#if auth.isLoading}
    <div class="flex h-screen w-full items-center justify-center bg-background">
        <div
            class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary"
        ></div>
    </div>
{:else}
    <div
        class="flex min-h-screen w-full flex-col overflow-x-hidden antialiased"
    >
        <Navbar />
        <main class="flex flex-col flex-grow">
            {@render children()}
        </main>
        <Footer />
    </div>
{/if}
