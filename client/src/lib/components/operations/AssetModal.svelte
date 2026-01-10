<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import * as Dialog from "$lib/components/ui/dialog";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Tabs from "$lib/components/ui/tabs";
    import { Loader2, Bus, Train, Ship } from "@lucide/svelte";

    import {
        fleetApi,
        AssetType,
        AssetStatus,
        type RegisterAssetRequest,
        type AssetConfig,
        type BusConfig,
        type TrainConfig,
        type LaunchConfig,
    } from "$lib/api/fleet";
    import BusConfigEditor from "$lib/components/config/BusConfigEditor.svelte";
    import TrainConfigEditor from "$lib/components/config/TrainConfigEditor.svelte";
    import LaunchConfigEditor from "$lib/components/config/LaunchConfigEditor.svelte";
    import { toast } from "svelte-sonner";

    export let open = false;
    export let onSuccess: () => void;

    let loading = false;
    let activeTab = "details";

    // Form State
    let name = "";
    let type: AssetType = AssetType.BUS;
    let license_plate = "";
    let vin = "";
    let make = "";
    let model = "";
    let year = new Date().getFullYear().toString();

    // Config State - type-specific
    let busConfig: BusConfig = {
        rows: 10,
        seats_per_row: 4,
        aisle_after_seat: 2,
        has_toilet: false,
        has_sleeper: false,
        categories: [],
    };

    let trainConfig: TrainConfig = {
        coaches: [],
    };

    let launchConfig: LaunchConfig = {
        decks: [],
    };

    // Features common to all
    let features: string[] = ["AC"];

    async function handleSubmit() {
        if (!name || !license_plate) {
            toast.error("Please fill required fields (Name, License Plate)");
            return;
        }

        // Build config based on type
        const config: AssetConfig = {
            features,
        };

        if (type === AssetType.BUS) {
            config.bus = busConfig;
        } else if (type === AssetType.TRAIN) {
            if (!trainConfig.coaches || trainConfig.coaches.length === 0) {
                toast.error("Please add at least one coach for the train");
                return;
            }
            config.train = trainConfig;
        } else if (type === AssetType.LAUNCH) {
            if (!launchConfig.decks || launchConfig.decks.length === 0) {
                toast.error("Please add at least one deck for the launch");
                return;
            }
            config.launch = launchConfig;
        }

        loading = true;
        try {
            const req: RegisterAssetRequest = {
                name,
                type,
                license_plate,
                vin,
                make,
                model,
                year: parseInt(year) || 2024,
                status: AssetStatus.ACTIVE,
                config,
            };

            await fleetApi.registerAsset(req);
            toast.success("Asset registered successfully");
            open = false;
            resetForm();
            onSuccess();
        } catch (e: any) {
            console.error(e);
            toast.error(
                "Failed to register asset: " + (e.message || "Unknown error"),
            );
        } finally {
            loading = false;
        }
    }

    function resetForm() {
        name = "";
        type = AssetType.BUS;
        license_plate = "";
        vin = "";
        make = "";
        model = "";
        year = new Date().getFullYear().toString();
        activeTab = "details";
        busConfig = {
            rows: 10,
            seats_per_row: 4,
            aisle_after_seat: 2,
            has_toilet: false,
            has_sleeper: false,
            categories: [],
        };
        trainConfig = { coaches: [] };
        launchConfig = { decks: [] };
        features = ["AC"];
    }

    const typeIcons = {
        [AssetType.BUS]: Bus,
        [AssetType.TRAIN]: Train,
        [AssetType.LAUNCH]: Ship,
    };
</script>

<Dialog.Root bind:open>
    <Dialog.Content class="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
        <Dialog.Header>
            <Dialog.Title class="flex items-center gap-2">
                <svelte:component this={typeIcons[type]} class="h-5 w-5" />
                Register New {type === AssetType.BUS
                    ? "Bus"
                    : type === AssetType.TRAIN
                      ? "Train"
                      : "Launch"}
            </Dialog.Title>
            <Dialog.Description>
                Configure the vehicle details and seating layout.
            </Dialog.Description>
        </Dialog.Header>

        <Tabs.Root bind:value={activeTab} class="mt-4">
            <Tabs.List class="grid w-full grid-cols-2">
                <Tabs.Trigger value="details">Basic Details</Tabs.Trigger>
                <Tabs.Trigger value="layout">Seat Layout</Tabs.Trigger>
            </Tabs.List>

            <Tabs.Content value="details" class="space-y-4 pt-4">
                <div class="grid gap-4">
                    <div class="grid gap-2">
                        <Label for="name">Asset Name</Label>
                        <Input
                            id="name"
                            bind:value={name}
                            placeholder="e.g., Dhaka Express 01, Subarna Express"
                        />
                    </div>

                    <div class="grid grid-cols-2 gap-4">
                        <div class="grid gap-2">
                            <Label>Type</Label>
                            <select
                                class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                                bind:value={type}
                            >
                                <option value={AssetType.BUS}>🚌 Bus</option>
                                <option value={AssetType.TRAIN}>🚂 Train</option
                                >
                                <option value={AssetType.LAUNCH}
                                    >🚢 Launch/Ship</option
                                >
                            </select>
                        </div>
                        <div class="grid gap-2">
                            <Label for="license">License/Registration</Label>
                            <Input
                                id="license"
                                bind:value={license_plate}
                                placeholder="DHAKA-D-11-0000"
                            />
                        </div>
                    </div>

                    <div class="grid grid-cols-2 gap-4">
                        <div class="grid gap-2">
                            <Label for="make">Make</Label>
                            <Input
                                id="make"
                                bind:value={make}
                                placeholder="Scania / BD Railway"
                            />
                        </div>
                        <div class="grid gap-2">
                            <Label for="model">Model</Label>
                            <Input
                                id="model"
                                bind:value={model}
                                placeholder="K410 / Ballam"
                            />
                        </div>
                    </div>

                    <div class="grid grid-cols-2 gap-4">
                        <div class="grid gap-2">
                            <Label for="year">Year</Label>
                            <Input
                                id="year"
                                type="number"
                                bind:value={year}
                                placeholder="2024"
                            />
                        </div>
                        <div class="grid gap-2">
                            <Label for="vin">VIN / ID (Optional)</Label>
                            <Input
                                id="vin"
                                bind:value={vin}
                                placeholder="XYZ..."
                            />
                        </div>
                    </div>
                </div>
            </Tabs.Content>

            <Tabs.Content value="layout" class="pt-4">
                {#if type === AssetType.BUS}
                    <BusConfigEditor bind:config={busConfig} />
                {:else if type === AssetType.TRAIN}
                    <TrainConfigEditor bind:config={trainConfig} />
                {:else if type === AssetType.LAUNCH}
                    <LaunchConfigEditor bind:config={launchConfig} />
                {/if}
            </Tabs.Content>
        </Tabs.Root>

        <Dialog.Footer class="mt-6">
            <Button variant="outline" onclick={() => (open = false)}>
                Cancel
            </Button>
            <Button onclick={handleSubmit} disabled={loading}>
                {#if loading}
                    <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                {/if}
                Register Asset
            </Button>
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
