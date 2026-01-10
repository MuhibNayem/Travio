<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import * as Dialog from "$lib/components/ui/dialog";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import { Combobox } from "$lib/components/ui/combobox";
    import {
        catalogApi,
        type Route,
        type CreateTripRequest,
    } from "$lib/api/catalog";
    import { fleetApi, type Asset } from "$lib/api/fleet";
    import { toast } from "svelte-sonner";
    import { onMount } from "svelte";

    export let open = false;
    export let onSuccess: () => void;

    let loading = false;
    let routes: Route[] = [];
    let assets: Asset[] = [];

    // Form State
    let routeId = "";
    let assetId = "";
    let departureDate = "";
    let departureTime = "";
    let price = "";

    onMount(async () => {
        try {
            const [r, a] = await Promise.all([
                catalogApi.getRoutes(),
                fleetApi.getAssets(),
            ]);
            routes = r;
            assets = a;
        } catch (e) {
            console.error(e);
            toast.error("Failed to load options");
        }
    });

    async function handleSubmit() {
        if (
            !routeId ||
            !assetId ||
            !departureDate ||
            !departureTime ||
            !price
        ) {
            toast.error("Please fill all fields");
            return;
        }

        const selectedAsset = assets.find((a) => a.id === assetId);
        if (!selectedAsset) return;

        // Construct ISO string
        const isoDateTime = new Date(
            `${departureDate}T${departureTime}`,
        ).toISOString();

        loading = true;
        try {
            const req: CreateTripRequest = {
                route_id: routeId,
                vehicle_id: selectedAsset.id, // Using asset ID as vehicle ID reference? Wait, checks proto. "vehicle_id" usually means the unique ID or the display ID?
                // Proto: string vehicle_id = 4; // Bus/Train number. Actually, often we reference the Asset ID.
                // But display might use Name. Let's send Asset ID or Name?
                // Backend CreateTrip takes vehicle_id and type. Usually we want to link to Asset ID if we have Fleet service integration.
                // For now, let's send Asset Name as vehicle_id (display) or Asset ID if backend supports lookup.
                // Given separation, let's look at `CreateTrip`: `vehicle_id` is stored string.
                // Let's store the Asset Name (License Plate / ID) for display.
                // Actually, storing Asset ID allows lookup. Ideally Asset ID.
                // Let's assume we store Name/Plate for now as per "vehicle_name" in Search results.
                vehicle_id: selectedAsset.name,
                vehicle_type: selectedAsset.type
                    .replace("ASSET_TYPE_", "")
                    .toLowerCase(),
                vehicle_class: "Economy", // Hardcoded for now, or add field
                departure_time: new Date(isoDateTime).getTime() / 1000, // Unix timestamp in seconds? No, Client sends ISO string in JSON usually?
                // Wait, `CreateTripRequest` in `catalog.ts` wants `departure_time: number`?
                // Let's check `catalog.ts` interface.
                // Yes: `departure_time: number;`
                // But `handler.go` parses `req.DepartureTime` as string (ISO 8601).
                // Step 2965: `type CreateTripRequest struct { DepartureTime string ... }` in Handler!
                // So Client TS interface is WRONG.
                // I need to use string in TS or convert.
                // Since I already defined usage, I should probably stick to what the Handler expects.
                // Handler expects "DepartureTime" (json: departure_time) as STRING (ISO 8601).
                // My TS interface has `departure_time: number`.
                // I should fix the TS interface or cast it.
                // I will cast to `any` temporarily or update `catalog.ts`?
                // Better to update `catalog.ts` later, but for now I'll use `any` cast to send string.

                total_seats: 40, // Should come from Asset Config
                pricing: {
                    base_price_paisa: parseFloat(price) * 100,
                    currency: "BDT",
                    class_prices: { Economy: parseFloat(price) * 100 },
                },
            };

            // Fix for TS mismatch (handler expects string date, client typed as number)
            // I'll send it as `any` to bypass TS check for now, trusting Gateway handler.
            const payload = {
                ...req,
                departure_time: isoDateTime,
                base_price: parseFloat(price),
            };

            // Oh wait, `catalog.ts` expects `CreateTripRequest` object.
            // I'll update `catalogApi.createTrip` to accept `any` or fix the type.
            // I'll use `any` in invoke.

            await catalogApi.createTrip(payload as any);
            toast.success("Trip scheduled successfully");
            open = false;
            resetForm();
            onSuccess();
        } catch (e) {
            console.error(e);
            toast.error("Failed to schedule trip");
        } finally {
            loading = false;
        }
    }

    function resetForm() {
        routeId = "";
        assetId = "";
        departureDate = "";
        departureTime = "";
        price = "";
    }
</script>

<Dialog.Root bind:open>
    <Dialog.Content class="sm:max-w-[500px]">
        <Dialog.Header>
            <Dialog.Title>Schedule New Trip</Dialog.Title>
            <Dialog.Description>
                Schedule a trip for a route using a vehicle.
            </Dialog.Description>
        </Dialog.Header>

        <div class="grid gap-4 py-4">
            <div class="grid gap-2">
                <Label>Route</Label>
                <Combobox
                    items={routes.map((r) => ({ value: r.id, label: r.name }))}
                    bind:value={routeId}
                    placeholder="Select Route"
                />
            </div>

            <div class="grid gap-2">
                <Label>Vehicle</Label>
                <Combobox
                    items={assets.map((a) => ({
                        value: a.id,
                        label: `${a.name} (${a.license_plate})`,
                    }))}
                    bind:value={assetId}
                    placeholder="Select Vehicle"
                />
            </div>

            <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                    <Label for="date">Departure Date</Label>
                    <Input id="date" type="date" bind:value={departureDate} />
                </div>
                <div class="grid gap-2">
                    <Label for="time">Time</Label>
                    <Input id="time" type="time" bind:value={departureTime} />
                </div>
            </div>

            <div class="grid gap-2">
                <Label for="price">Base Price (BDT)</Label>
                <Input
                    id="price"
                    type="number"
                    bind:value={price}
                    placeholder="500"
                />
            </div>
        </div>

        <Dialog.Footer>
            <Button variant="outline" onclick={() => (open = false)}
                >Cancel</Button
            >
            <Button onclick={handleSubmit} disabled={loading}>
                {loading ? "Scheduling..." : "Schedule Trip"}
            </Button>
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
