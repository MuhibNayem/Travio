<script lang="ts">
    import Navbar from "$lib/components/layouts/Navbar.svelte";
    import Footer from "$lib/components/layouts/Footer.svelte";

    let { children } = $props();
</script>

<div class="flex min-h-screen w-full flex-col overflow-x-hidden antialiased">
    <Navbar />
    <main class="flex flex-col flex-grow">
        {@render children()}
    </main>
    <Footer />
</div>
