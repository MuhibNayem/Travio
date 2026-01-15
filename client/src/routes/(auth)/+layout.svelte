<script lang="ts">
    let { children } = $props();
</script>

<div
    class="flex min-h-screen w-full flex-col items-center justify-center overflow-x-hidden antialiased bg-gray-50 dark:bg-black p-4"
>
    <!-- Simple Centered Layout for Auth -->
    <div class="w-full max-w-md">
        {@render children()}
    </div>
</div>
