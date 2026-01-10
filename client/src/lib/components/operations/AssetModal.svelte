<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import * as Dialog from "$lib/components/ui/dialog";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import {
        Select,
        SelectContent,
        SelectItem,
        SelectTrigger,
        SelectValue,
    } from "$lib/components/ui/select";
    import {
        fleetApi,
        AssetType,
        AssetStatus,
        type RegisterAssetRequest,
    } from "$lib/api/fleet";
    import { toast } from "svelte-sonner";

    export let open = false;
    export let onSuccess: () => void;

    let loading = false;

    // Form State
    let name = "";
    let type = AssetType.BUS; // Default
    let license_plate = "";
    let vin = "";
    let make = "";
    let model = "";
    let year = new Date().getFullYear().toString();

    // Config State (Mocking for now, can expand later)
    let layout_type = "";

    async function handleSubmit() {
        if (!name || !license_plate) {
            toast.error("Please fill required fields (Name, License Plate)");
            return;
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
                config: {
                    layout_type: layout_type || "standard",
                    features: "AC, WiFi",
                },
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
        layout_type = "";
    }
</script>

<Dialog.Root bind:open>
    <Dialog.Content class="sm:max-w-[500px]">
        <Dialog.Header>
            <Dialog.Title>Register New Asset</Dialog.Title>
            <Dialog.Description>
                Add a new vehicle to your fleet.
            </Dialog.Description>
        </Dialog.Header>

        <div class="grid gap-4 py-4">
            <div class="grid gap-2">
                <Label for="name">Asset Name (e.g. Bus #101)</Label>
                <Input
                    id="name"
                    bind:value={name}
                    placeholder="Dhaka Express 01"
                />
            </div>

            <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                    <Label>Type</Label>
                    <Select type="single" bind:value={type}>
                        <SelectTrigger>
                            <SelectValue placeholder="Select type" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value={AssetType.BUS}>Bus</SelectItem>
                            <SelectItem value={AssetType.TRAIN}
                                >Train</SelectItem
                            >
                            <SelectItem value={AssetType.LAUNCH}
                                >Launch/Ship</SelectItem
                            >
                        </SelectContent>
                    </Select>
                </div>
                <div class="grid gap-2">
                    <Label for="license">License Plate</Label>
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
                    <Input id="make" bind:value={make} placeholder="Scania" />
                </div>
                <div class="grid gap-2">
                    <Label for="model">Model</Label>
                    <Input id="model" bind:value={model} placeholder="K410" />
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
                    <Label for="vin">VIN (Optional)</Label>
                    <Input id="vin" bind:value={vin} placeholder="XYZ..." />
                </div>
            </div>
        </div>

        <Dialog.Footer>
            <Button variant="outline" onclick={() => (open = false)}
                >Cancel</Button
            >
            <Button onclick={handleSubmit} disabled={loading}>
                {loading ? "Registering..." : "Register Asset"}
            </Button>
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
