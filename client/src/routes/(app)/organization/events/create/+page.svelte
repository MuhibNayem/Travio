<script lang="ts">
    import EventWizard from "$lib/components/events/EventWizard.svelte";
</script>

<div class="space-y-6">
    <div class="max-w-3xl mx-auto">
        <h1 class="text-3xl font-bold tracking-tight">Create New Event</h1>
        <p class="text-muted-foreground mt-2">
            Follow the steps to configure your event details, venue, and
            schedule.
        </p>
    </div>

    <EventWizard />
</div>
