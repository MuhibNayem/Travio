<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import * as Dialog from "$lib/components/ui/dialog";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import { Combobox } from "$lib/components/ui/combobox";
    import {
        catalogApi,
        type Station,
        type CreateRouteRequest,
    } from "$lib/api/catalog";
    import { toast } from "svelte-sonner";
    import { onMount } from "svelte";
    import { stationsStore } from "$lib/stores/stations.svelte";

    export let open = false;
    export let onSuccess: () => void;

    let loading = false;

    // Form State
    let name = "";
    let code = "";
    let originId = "";
    let destinationId = "";
    let distance = "";
    let duration = "";

    onMount(async () => {
        try {
            await stationsStore.load();
        } catch (e) {
            console.error(e);
            toast.error("Failed to load stations");
        }
    });

    async function handleSubmit() {
        if (
            !name ||
            !code ||
            !originId ||
            !destinationId ||
            !distance ||
            !duration
        ) {
            toast.error("Please fill all fields");
            return;
        }

        loading = true;
        try {
            const req: CreateRouteRequest = {
                name,
                code,
                origin_station_id: originId,
                destination_station_id: destinationId,
                distance_km: parseInt(distance),
                estimated_duration_minutes: parseInt(duration),
            };

            await catalogApi.createRoute(req);
            toast.success("Route created successfully");
            open = false;
            resetForm();
            onSuccess();
        } catch (e) {
            console.error(e);
            toast.error("Failed to create route");
        } finally {
            loading = false;
        }
    }

    function resetForm() {
        name = "";
        code = "";
        originId = "";
        destinationId = "";
        distance = "";
        duration = "";
    }
</script>

<Dialog.Root bind:open>
    <Dialog.Content class="sm:max-w-[500px]">
        <Dialog.Header>
            <Dialog.Title>Create New Route</Dialog.Title>
            <Dialog.Description>
                Define a new route between two stations.
            </Dialog.Description>
        </Dialog.Header>

        <div class="grid gap-4 py-4">
            <div class="grid gap-2">
                <Label for="name">Route Name</Label>
                <Input
                    id="name"
                    bind:value={name}
                    placeholder="e.g. Dhaka - Chittagong Express"
                />
            </div>

            <div class="grid gap-2">
                <Label for="code">Route Code</Label>
                <Input
                    id="code"
                    bind:value={code}
                    placeholder="e.g. DHA-CTG-001"
                />
            </div>

            <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                    <Label>Origin Station</Label>
                    <Combobox
                        items={stationsStore.visibleStations.map((s) => ({
                            value: s.id,
                            label: s.name,
                        }))}
                        bind:value={originId}
                        placeholder="Select Origin"
                        loading={stationsStore.loading}
                        loadingMore={stationsStore.loadingMore}
                        onSearch={(q) => stationsStore.handleSearch(q)}
                        onEndReached={() => stationsStore.loadMore()}
                        onClose={() => stationsStore.resetToDefault()}
                    />
                </div>
                <div class="grid gap-2">
                    <Label>Destination Station</Label>
                    <Combobox
                        items={stationsStore.visibleStations.map((s) => ({
                            value: s.id,
                            label: s.name,
                        }))}
                        bind:value={destinationId}
                        placeholder="Select Destination"
                        loading={stationsStore.loading}
                        loadingMore={stationsStore.loadingMore}
                        onSearch={(q) => stationsStore.handleSearch(q)}
                        onEndReached={() => stationsStore.loadMore()}
                        onClose={() => stationsStore.resetToDefault()}
                    />
                </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                    <Label for="distance">Distance (km)</Label>
                    <Input
                        id="distance"
                        type="number"
                        bind:value={distance}
                        placeholder="250"
                    />
                </div>
                <div class="grid gap-2">
                    <Label for="duration">Duration (mins)</Label>
                    <Input
                        id="duration"
                        type="number"
                        bind:value={duration}
                        placeholder="300"
                    />
                </div>
            </div>
        </div>

        <Dialog.Footer>
            <Button variant="outline" onclick={() => (open = false)}
                >Cancel</Button
            >
            <Button onclick={handleSubmit} disabled={loading}>
                {loading ? "Creating..." : "Create Route"}
            </Button>
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
