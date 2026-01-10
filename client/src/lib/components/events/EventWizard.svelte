<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card";
    import { Textarea } from "$lib/components/ui/textarea";
    import { eventsApi, type Venue } from "$lib/api/events";
    import { toast } from "svelte-sonner";
    import { auth } from "$lib/runes/auth.svelte";
    import {
        Loader2,
        Calendar,
        MapPin,
        Image as ImageIcon,
        CheckCircle,
        ArrowRight,
        ArrowLeft,
        Plus,
    } from "@lucide/svelte";
    import { goto } from "$app/navigation";
    import { onMount } from "svelte";

    let step = $state(1);
    let loading = $state(false);
    let venues = $state<Venue[]>([]);

    // Form State
    let title = $state("");
    let description = $state("");
    let category = $state("Music");
    let venueId = $state("");
    let startDate = $state("");
    let startTime = $state("");
    let endDate = $state("");
    let endTime = $state("");
    let bannerUrl = $state("");

    onMount(async () => {
        try {
            const orgId = auth.user?.organizationId;
            if (orgId) {
                venues = await eventsApi.getVenues(orgId);
            }
        } catch (e) {
            console.error("Failed to load venues", e);
        }
    });

    async function handleSubmit() {
        if (!title || !venueId || !startDate || !startTime) {
            toast.error("Please complete all required fields");
            return;
        }

        loading = true;
        try {
            const orgId = auth.user?.organizationId;
            if (!orgId) throw new Error("User organization not found");

            // Combine Date & Time to ISO
            const startISO = new Date(
                `${startDate}T${startTime}`,
            ).toISOString();
            const endISO =
                endDate && endTime
                    ? new Date(`${endDate}T${endTime}`).toISOString()
                    : new Date(
                          new Date(`${startDate}T${startTime}`).getTime() +
                              3 * 3600000,
                      ).toISOString(); // Default 3h

            await eventsApi.createEvent({
                organization_id: orgId,
                venue_id: venueId,
                title,
                description,
                category,
                start_time: startISO,
                end_time: endISO,
            });

            toast.success("Event created successfully!");
            goto("/organization/events");
        } catch (e) {
            console.error(e);
            toast.error("Failed to create event");
        } finally {
            loading = false;
        }
    }

    const categories = [
        "Music",
        "Sports",
        "Technology",
        "Arts",
        "Conference",
        "Workshop",
    ];
</script>

<div class="max-w-3xl mx-auto py-8">
    <!-- Steps Indicator -->
    <div class="flex justify-between mb-8 relative">
        <div
            class="absolute top-1/2 left-0 w-full h-1 bg-muted -z-10 rounded-full"
        ></div>
        <div
            class="absolute top-1/2 left-0 h-1 bg-primary -z-10 rounded-full transition-all duration-300"
            style="width: {(step - 1) * 33}%"
        ></div>

        {#each [1, 2, 3, 4] as s}
            <div class="flex flex-col items-center gap-2 bg-background px-2">
                <div
                    class="h-8 w-8 rounded-full border-2 flex items-center justify-center font-bold text-sm transition-colors
                    {step >= s
                        ? 'border-primary bg-primary text-primary-foreground'
                        : 'border-muted-foreground text-muted-foreground bg-card'}"
                >
                    {#if step > s}
                        <CheckCircle class="h-5 w-5" />
                    {:else}
                        {s}
                    {/if}
                </div>
                <span
                    class="text-xs font-medium {step >= s
                        ? 'text-primary'
                        : 'text-muted-foreground'}"
                >
                    {s === 1
                        ? "Details"
                        : s === 2
                          ? "Venue"
                          : s === 3
                            ? "Schedule"
                            : "Preview"}
                </span>
            </div>
        {/each}
    </div>

    <!-- Step 1: Basic Details -->
    {#if step === 1}
        <Card.Root>
            <Card.Header>
                <Card.Title>Event Details</Card.Title>
                <Card.Description
                    >Let's start with the basics of your event.</Card.Description
                >
            </Card.Header>
            <Card.Content class="space-y-4">
                <div class="space-y-2">
                    <Label>Event Title</Label>
                    <Input
                        bind:value={title}
                        placeholder="e.g. Winter Rockfest 2026"
                        class="text-lg"
                    />
                </div>

                <div class="space-y-2">
                    <Label>Description</Label>
                    <Textarea
                        bind:value={description}
                        placeholder="Describe what attendees can expect..."
                        rows={4}
                    />
                </div>

                <div class="space-y-2">
                    <Label>Category</Label>
                    <div class="grid grid-cols-3 gap-2">
                        {#each categories as cat}
                            <button
                                class="border rounded-md p-3 text-sm font-medium transition-all hover:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20
                                {category === cat
                                    ? 'border-primary bg-primary/5 text-primary'
                                    : 'bg-card text-muted-foreground'}"
                                onclick={() => (category = cat)}
                            >
                                {cat}
                            </button>
                        {/each}
                    </div>
                </div>
            </Card.Content>
            <Card.Footer class="justify-end">
                <Button onclick={() => (step = 2)} disabled={!title}>
                    Next Step <ArrowRight class="ml-2 h-4 w-4" />
                </Button>
            </Card.Footer>
        </Card.Root>

        <!-- Step 2: Venue Selection -->
    {:else if step === 2}
        <Card.Root>
            <Card.Header>
                <Card.Title>Venue Selection</Card.Title>
                <Card.Description
                    >Where will this event happen?</Card.Description
                >
            </Card.Header>
            <Card.Content class="space-y-4">
                {#if venues.length === 0}
                    <div
                        class="text-center p-8 border rounded-lg border-dashed"
                    >
                        <MapPin
                            class="h-10 w-10 mx-auto text-muted-foreground opacity-50 mb-4"
                        />
                        <h3 class="font-medium text-lg">No Venues Found</h3>
                        <p class="text-muted-foreground mb-4">
                            You need to create a venue first.
                        </p>
                        <Button
                            variant="outline"
                            href="/organization/events/venues"
                        >
                            Create Venue
                        </Button>
                    </div>
                {:else}
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {#each venues as v}
                            <button
                                class="text-left border rounded-xl p-4 transition-all hover:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 relative overflow-hidden
                                {venueId === v.id
                                    ? 'border-primary bg-primary/5 ring-1 ring-primary'
                                    : 'bg-card'}"
                                onclick={() => (venueId = v.id)}
                            >
                                <div class="flex items-start justify-between">
                                    <div>
                                        <h4 class="font-semibold">{v.name}</h4>
                                        <p
                                            class="text-sm text-muted-foreground"
                                        >
                                            {v.city}, {v.country}
                                        </p>
                                    </div>
                                    {#if venueId === v.id}
                                        <CheckCircle
                                            class="h-5 w-5 text-primary"
                                        />
                                    {/if}
                                </div>
                                <div
                                    class="mt-4 flex items-center gap-2 text-xs text-muted-foreground"
                                >
                                    <span class="bg-muted px-2 py-1 rounded-md"
                                        >{v.type
                                            .toString()
                                            .replace("VENUE_TYPE_", "")}</span
                                    >
                                    <span>Capacity: {v.capacity}</span>
                                </div>
                            </button>
                        {/each}
                        <Button
                            variant="outline"
                            class="h-full border-dashed"
                            href="/organization/events/venues"
                        >
                            <Plus class="mr-2 h-4 w-4" /> Create New Venue
                        </Button>
                    </div>
                {/if}
            </Card.Content>
            <Card.Footer class="justify-between">
                <Button variant="ghost" onclick={() => (step = 1)}>
                    <ArrowLeft class="mr-2 h-4 w-4" /> Back
                </Button>
                <Button onclick={() => (step = 3)} disabled={!venueId}>
                    Next Step <ArrowRight class="ml-2 h-4 w-4" />
                </Button>
            </Card.Footer>
        </Card.Root>

        <!-- Step 3: Schedule -->
    {:else if step === 3}
        <Card.Root>
            <Card.Header>
                <Card.Title>Date & Time</Card.Title>
                <Card.Description>When is standard time?</Card.Description>
            </Card.Header>
            <Card.Content class="space-y-6">
                <div class="grid grid-cols-2 gap-6">
                    <div class="space-y-2">
                        <Label>Start Date</Label>
                        <Input type="date" bind:value={startDate} />
                    </div>
                    <div class="space-y-2">
                        <Label>Start Time</Label>
                        <Input type="time" bind:value={startTime} />
                    </div>
                </div>

                <div class="grid grid-cols-2 gap-6 pt-4 border-t">
                    <div class="space-y-2">
                        <Label>End Date (Optional)</Label>
                        <Input type="date" bind:value={endDate} />
                    </div>
                    <div class="space-y-2">
                        <Label>End Time (Optional)</Label>
                        <Input type="time" bind:value={endTime} />
                    </div>
                </div>
            </Card.Content>
            <Card.Footer class="justify-between">
                <Button variant="ghost" onclick={() => (step = 2)}>
                    <ArrowLeft class="mr-2 h-4 w-4" /> Back
                </Button>
                <Button
                    onclick={() => (step = 4)}
                    disabled={!startDate || !startTime}
                >
                    Next Step <ArrowRight class="ml-2 h-4 w-4" />
                </Button>
            </Card.Footer>
        </Card.Root>

        <!-- Step 4: Preview & Image -->
    {:else}
        <Card.Root>
            <Card.Header>
                <Card.Title>Final Touches</Card.Title>
                <Card.Description
                    >Add a cover image and review.</Card.Description
                >
            </Card.Header>
            <Card.Content class="space-y-6">
                <div class="space-y-2">
                    <Label>Cover Image URL (Optional)</Label>
                    <div class="flex gap-2">
                        <Input
                            bind:value={bannerUrl}
                            placeholder="https://..."
                        />
                        {#if bannerUrl}
                            <div
                                class="h-10 w-10 rounded overflow-hidden border"
                            >
                                <img
                                    src={bannerUrl}
                                    alt="Preview"
                                    class="h-full w-full object-cover"
                                />
                            </div>
                        {:else}
                            <div
                                class="h-10 w-10 rounded border bg-muted flex items-center justify-center"
                            >
                                <ImageIcon
                                    class="h-5 w-5 text-muted-foreground"
                                />
                            </div>
                        {/if}
                    </div>
                </div>

                <div class="bg-muted/50 rounded-xl p-6 border">
                    <h3 class="text-xl font-bold">{title}</h3>
                    <div
                        class="flex items-center gap-2 mt-2 text-sm text-muted-foreground"
                    >
                        <Calendar class="h-4 w-4" />
                        <span
                            >{new Date(
                                `${startDate}T${startTime}`,
                            ).toLocaleDateString("en-US", {
                                weekday: "long",
                                month: "long",
                                day: "numeric",
                                year: "numeric",
                            })}</span
                        >
                        <span
                            >at {new Date(
                                `2000-01-01T${startTime}`,
                            ).toLocaleTimeString("en-US", {
                                hour: "numeric",
                                minute: "2-digit",
                            })}</span
                        >
                    </div>
                    <div
                        class="flex items-center gap-2 mt-1 text-sm text-muted-foreground"
                    >
                        <MapPin class="h-4 w-4" />
                        <span
                            >{venues.find((v) => v.id === venueId)?.name}, {venues.find(
                                (v) => v.id === venueId,
                            )?.city}</span
                        >
                    </div>
                    <div class="mt-4 pt-4 border-t border-dashed">
                        <p class="text-sm line-clamp-3">
                            {description || "No description provided."}
                        </p>
                    </div>
                </div>
            </Card.Content>
            <Card.Footer class="justify-between">
                <Button variant="ghost" onclick={() => (step = 3)}>
                    <ArrowLeft class="mr-2 h-4 w-4" /> Back
                </Button>
                <Button onclick={handleSubmit} disabled={loading} size="lg">
                    {#if loading}
                        <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                    {:else}
                        <CheckCircle class="mr-2 h-4 w-4" />
                    {/if}
                    Create Event
                </Button>
            </Card.Footer>
        </Card.Root>
    {/if}
</div>
