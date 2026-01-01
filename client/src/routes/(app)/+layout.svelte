<script lang="ts">
    import { goto } from "$app/navigation";
    import { auth } from "$lib/runes/auth.svelte";
    import Navbar from "$lib/components/layouts/Navbar.svelte";
    import Footer from "$lib/components/layouts/Footer.svelte";

    let { children } = $props();

    $effect(() => {
        if (!auth.isAuthenticated) {
            goto("/login");
        }
    });
</script>

<div class="flex min-h-screen w-full flex-col overflow-x-hidden antialiased">
    <Navbar />
    <main class="flex flex-col flex-grow">
        {@render children()}
    </main>
    <Footer />
</div>
