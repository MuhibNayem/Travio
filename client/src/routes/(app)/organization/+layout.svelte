<script lang="ts">
    import Sidebar from "$lib/components/layouts/Sidebar.svelte";
</script>

<div class="flex min-h-screen pt-16">
    <Sidebar />

    <main class="ml-0 flex-1 p-4 transition-all duration-300 md:ml-64 md:p-8">
        <div class="mx-auto max-w-7xl animate-in fade-in zoom-in duration-300">
            <slot />
        </div>
    </main>
</div>
