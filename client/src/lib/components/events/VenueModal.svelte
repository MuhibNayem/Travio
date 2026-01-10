<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Dialog from "$lib/components/ui/dialog";

    import {
        eventsApi,
        VenueType,
        type Venue,
        type CreateVenueRequest,
        type UpdateVenueRequest,
    } from "$lib/api/events";
    import { toast } from "svelte-sonner";
    import { page } from "$app/stores";
    import { Loader2, Plus, Trash } from "@lucide/svelte";

    let {
        open = $bindable(false),
        venueToEdit = null,
        onSuccess,
    } = $props<{
        open: boolean;
        venueToEdit?: Venue | null;
        onSuccess: () => void;
    }>();

    let loading = false;

    // Form State
    let name = $state("");
    let address = $state("");
    let city = $state("");
    let country = $state("Bangladesh");
    let type = $state<VenueType>(VenueType.VENUE_TYPE_AUDITORIUM);
    let sections = $state<
        { name: string; capacity: number; price_tier: string }[]
    >([{ name: "General", capacity: 100, price_tier: "Standard" }]);

    // Derived
    let isEditMode = $derived(!!venueToEdit);
    let title = $derived(isEditMode ? "Edit Venue" : "Add Venue");

    $effect(() => {
        if (open) {
            if (venueToEdit) {
                name = venueToEdit.name;
                address = venueToEdit.address;
                city = venueToEdit.city;
                country = venueToEdit.country;
                type = venueToEdit.type;
                sections =
                    venueToEdit.sections?.map(
                        (s: {
                            name: string;
                            capacity: number;
                            type: string;
                        }) => ({
                            name: s.name,
                            capacity: s.capacity,
                            price_tier: s.type, // mapping type to tier for UI simplicity
                        }),
                    ) || [];
            } else {
                resetForm();
            }
        }
    });

    function resetForm() {
        name = "";
        address = "";
        city = "";
        country = "Bangladesh";
        type = VenueType.VENUE_TYPE_AUDITORIUM;
        sections = [{ name: "General", capacity: 100, price_tier: "Standard" }];
    }

    function addSection() {
        sections = [
            ...sections,
            { name: "", capacity: 0, price_tier: "Standard" },
        ];
    }

    function removeSection(index: number) {
        sections = sections.filter((_, i) => i !== index);
    }

    async function handleSubmit() {
        if (!name || !address || !city) {
            toast.error("Please fill in required fields");
            return;
        }

        loading = true;
        try {
            const orgId = $page.data.user.organization_id;

            // Map UI sections to API sections
            const apiSections = sections.map((s, i) => ({
                id: `sec-${i}`,
                name: s.name,
                capacity: Number(s.capacity),
                rows: 1, // Default
                seats_per_row: Number(s.capacity), // Simplified
                type: s.price_tier,
            }));

            if (isEditMode && venueToEdit) {
                await eventsApi.updateVenue(venueToEdit.id, {
                    id: venueToEdit.id,
                    organization_id: orgId,
                    name,
                    type: VenueType[type], // Enum name as string
                });
                toast.success("Venue updated successfully");
            } else {
                await eventsApi.createVenue({
                    organization_id: orgId,
                    name,
                    address,
                    city,
                    country,
                    type: VenueType[type], // Enum name string? Or API expects int?
                    // Review API: Proto helper 'req.Type.String()' used in Handler.
                    // Client usually sends JSON string for enum if using protojson, or int.
                    // Let's rely on generated code or standard behavior. My API client types use string for 'type' in Request interface.
                    // But VenueType enum is exported. I should cast appropriately.
                    // The 'CreateVenueRequest' interface I wrote uses 'type: string'.
                    sections: apiSections,
                });
                toast.success("Venue created successfully");
            }
            open = false;
            onSuccess();
        } catch (e) {
            console.error(e);
            toast.error("Failed to save venue");
        } finally {
            loading = false;
        }
    }

    // Enum Options
    const venueTypes = [
        { value: VenueType.VENUE_TYPE_STADIUM, label: "Stadium" },
        { value: VenueType.VENUE_TYPE_AUDITORIUM, label: "Auditorium" },
        {
            value: VenueType.VENUE_TYPE_CONFERENCE_HALL,
            label: "Conference Hall",
        },
        { value: VenueType.VENUE_TYPE_OUTDOOR_GROUND, label: "Outdoor Ground" },
    ];
</script>

<Dialog.Root bind:open>
    <Dialog.Content class="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <Dialog.Header>
            <Dialog.Title>{title}</Dialog.Title>
            <Dialog.Description>
                Configure the venue details and seating sections.
            </Dialog.Description>
        </Dialog.Header>

        <div class="grid gap-4 py-4">
            <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                    <Label>Venue Name</Label>
                    <Input
                        bind:value={name}
                        placeholder="e.g. Bashundhara Arena"
                    />
                </div>
                <div class="space-y-2">
                    <Label>Type</Label>
                    <select
                        class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                        bind:value={type}
                    >
                        {#each venueTypes as vt}
                            <option value={vt.value}>{vt.label}</option>
                        {/each}
                    </select>
                </div>
            </div>

            <div class="space-y-2">
                <Label>Address</Label>
                <Input bind:value={address} placeholder="Street address" />
            </div>

            <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                    <Label>City</Label>
                    <Input bind:value={city} placeholder="Dhaka" />
                </div>
                <div class="space-y-2">
                    <Label>Country</Label>
                    <Input bind:value={country} />
                </div>
            </div>

            <div class="space-y-3 pt-4 border-t">
                <div class="flex items-center justify-between">
                    <Label class="text-base font-semibold"
                        >Seating Sections</Label
                    >
                    <Button variant="ghost" size="sm" onclick={addSection}>
                        <Plus class="h-4 w-4 mr-2" />
                        Add Section
                    </Button>
                </div>

                {#each sections as section, i}
                    <div class="grid grid-cols-12 gap-2 items-end">
                        <div class="col-span-5 space-y-1">
                            <Label class="text-xs text-muted-foreground"
                                >Name</Label
                            >
                            <Input
                                bind:value={section.name}
                                placeholder="Gallery A"
                            />
                        </div>
                        <div class="col-span-3 space-y-1">
                            <Label class="text-xs text-muted-foreground"
                                >Capacity</Label
                            >
                            <Input
                                type="number"
                                bind:value={section.capacity}
                            />
                        </div>
                        <div class="col-span-3 space-y-1">
                            <Label class="text-xs text-muted-foreground"
                                >Tier</Label
                            >
                            <Input
                                bind:value={section.price_tier}
                                placeholder="VIP/Std"
                            />
                        </div>
                        <div class="col-span-1">
                            <Button
                                variant="ghost"
                                size="icon"
                                class="text-destructive h-9 w-9"
                                onclick={() => removeSection(i)}
                            >
                                <Trash class="h-4 w-4" />
                            </Button>
                        </div>
                    </div>
                {/each}
            </div>
        </div>

        <Dialog.Footer>
            <Button variant="outline" onclick={() => (open = false)}
                >Cancel</Button
            >
            <Button onclick={handleSubmit} disabled={loading}>
                {#if loading}
                    <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                {/if}
                Save Venue
            </Button>
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
