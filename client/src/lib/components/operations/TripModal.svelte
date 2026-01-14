<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import * as Dialog from "$lib/components/ui/dialog";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import { Combobox } from "$lib/components/ui/combobox";
    import * as Checkbox from "$lib/components/ui/checkbox";
    import {
        catalogApi,
        type Route,
        type CreateScheduleRequest,
        type Station,
        type SegmentPricing,
    } from "$lib/api/catalog";
    import { getOrganization } from "$lib/api/auth";
    import { fleetApi, type Asset } from "$lib/api/fleet";
    import { toast } from "svelte-sonner";
    import { onMount } from "svelte";

    export let open = false;
    export let onSuccess: () => void;

    let loading = false;
    let routes: Route[] = [];
    let assets: Asset[] = [];
    let stations: Station[] = [];
    let stationNameMap: Record<string, string> = {};

    // Form State
    let routeId = "";
    let assetId = "";
    let departureTime = "";
    let startDate = "";
    let endDate = "";
    let vehicleClass = "economy";
    let totalSeats = 40;
    let basePrice = "";
    let currency = "BDT";
    let classPriceOverrides: Array<{ key: string; value: string }> = [];
    let seatCategoryOverrides: Array<{ key: string; value: string }> = [];
    let segmentPriceOverrides: Array<{ fromId: string; toId: string; basePrice: string }> = [];
    let additionalDepartures = "";
    let timezone = "Asia/Dhaka";
    let selectedDays = new Set<number>();

    const weekdayOptions = [
        { label: "Mon", bit: 1 },
        { label: "Tue", bit: 2 },
        { label: "Wed", bit: 4 },
        { label: "Thu", bit: 8 },
        { label: "Fri", bit: 16 },
        { label: "Sat", bit: 32 },
        { label: "Sun", bit: 64 },
    ];

    onMount(async () => {
        try {
            const [r, a] = await Promise.all([
                catalogApi.getRoutes(),
                fleetApi.getAssets(),
            ]);
            routes = r;
            assets = a;
            try {
                const org = await getOrganization();
                if (org?.currency) {
                    currency = org.currency;
                }
            } catch (error) {
                console.error(error);
            }
            stations = await catalogApi.getStations();
            stationNameMap = stations.reduce(
                (acc, station) => {
                    acc[station.id] = station.name;
                    return acc;
                },
                {} as Record<string, string>,
            );
        } catch (e) {
            console.error(e);
            toast.error("Failed to load options");
        }
    });

    $: if (routeId) {
        refreshSegmentOverrides(routeId);
    }

    $: if (assetId) {
        const selectedAsset = assets.find((a) => a.id === assetId);
        refreshSeatCategoryOverrides(selectedAsset);
    }

    function parseTimeToMinutes(value: string): number | null {
        if (!value) return null;
        const [h, m] = value.split(":").map((v) => Number(v));
        if (Number.isNaN(h) || Number.isNaN(m)) return null;
        return h * 60 + m;
    }

    function parseAdditionalTimes(value: string): number[] {
        if (!value) return [];
        return value
            .split(",")
            .map((v) => v.trim())
            .filter(Boolean)
            .map(parseTimeToMinutes)
            .filter((v): v is number => v !== null);
    }

    function getDaysMask(): number {
        let mask = 0;
        selectedDays.forEach((bit) => {
            mask |= bit;
        });
        return mask;
    }

    function buildClassPricesMap(): Record<string, number> {
        return classPriceOverrides.reduce((acc, entry) => {
            const key = entry.key.trim();
            const value = Number(entry.value);
            if (key && Number.isFinite(value) && value > 0) {
                acc[key] = Math.round(value * 100);
            }
            return acc;
        }, {} as Record<string, number>);
    }

    function buildSeatCategoryPricesMap(): Record<string, number> {
        return seatCategoryOverrides.reduce((acc, entry) => {
            const key = entry.key.trim();
            const value = Number(entry.value);
            if (key && Number.isFinite(value) && value > 0) {
                acc[key] = Math.round(value * 100);
            }
            return acc;
        }, {} as Record<string, number>);
    }

    function buildSegmentPrices(): SegmentPricing[] {
        return segmentPriceOverrides
            .map((segment) => {
                const value = Number(segment.basePrice);
                if (!Number.isFinite(value) || value <= 0) return null;
                return {
                    from_station_id: segment.fromId,
                    to_station_id: segment.toId,
                    base_price_paisa: Math.round(value * 100),
                    class_prices: {},
                    seat_category_prices: {},
                };
            })
            .filter((segment): segment is SegmentPricing => segment !== null);
    }

    function refreshSeatCategoryOverrides(selectedAsset: Asset | undefined) {
        if (!selectedAsset) return;
        const next: Array<{ key: string; value: string }> = [];
        if (selectedAsset.config?.bus?.categories?.length) {
            selectedAsset.config.bus.categories.forEach((category) => {
                next.push({
                    key: category.name,
                    value: category.price_paisa
                        ? String(category.price_paisa / 100)
                        : "",
                });
            });
        }
        if (selectedAsset.config?.train?.coaches?.length) {
            selectedAsset.config.train.coaches.forEach((coach) => {
                next.push({
                    key: coach.name || coach.class,
                    value: coach.price_paisa
                        ? String(coach.price_paisa / 100)
                        : "",
                });
            });
        }
        if (selectedAsset.config?.launch?.decks?.length) {
            selectedAsset.config.launch.decks.forEach((deck) => {
                const baseValue = deck.seat_price_paisa || 0;
                next.push({
                    key: deck.name || deck.id,
                    value: baseValue ? String(baseValue / 100) : "",
                });
            });
        }
        if (next.length > 0) {
            seatCategoryOverrides = next;
        }
    }

    function refreshSegmentOverrides(routeIdValue: string) {
        const route = routes.find((r) => r.id === routeIdValue);
        if (!route) return;
        const stops = [
            route.origin_station_id,
            ...(route.intermediate_stops || [])
                .sort((a, b) => a.sequence - b.sequence)
                .map((stop) => stop.station_id),
            route.destination_station_id,
        ];
        const segments: Array<{ fromId: string; toId: string; basePrice: string }> = [];
        for (let i = 0; i < stops.length - 1; i += 1) {
            segments.push({
                fromId: stops[i],
                toId: stops[i + 1],
                basePrice: "",
            });
        }
        segmentPriceOverrides = segments;
    }

    async function handleSubmit() {
        if (
            !routeId ||
            !assetId ||
            !departureTime ||
            !startDate ||
            !endDate ||
            selectedDays.size === 0
        ) {
            toast.error("Please fill all fields");
            return;
        }

        const selectedAsset = assets.find((a) => a.id === assetId);
        if (!selectedAsset) return;

        const departureMinutes = parseTimeToMinutes(departureTime);
        if (departureMinutes === null) {
            toast.error("Invalid departure time");
            return;
        }

        const additionalTimes = parseAdditionalTimes(additionalDepartures);
        const daysMask = getDaysMask();
        const seats = Number(totalSeats);
        if (!Number.isFinite(seats) || seats <= 0) {
            toast.error("Total seats must be greater than zero");
            return;
        }

        const basePriceNumber = Number(basePrice);
        if (!Number.isFinite(basePriceNumber) || basePriceNumber <= 0) {
            toast.error("Base price must be greater than zero");
            return;
        }

        const basePricePaisa = Math.round(basePriceNumber * 100);
        const currencyCode = currency.trim() || "BDT";
        const classPrices = buildClassPricesMap();
        const seatCategoryPrices = buildSeatCategoryPricesMap();
        const segmentPrices = buildSegmentPrices();

        loading = true;
        try {
            const baseSchedule: CreateScheduleRequest = {
                route_id: routeId,
                vehicle_id: selectedAsset.id,
                vehicle_type: selectedAsset.type
                    .replace("ASSET_TYPE_", "")
                    .toLowerCase(),
                vehicle_class: vehicleClass,
                total_seats: seats,
                pricing: {
                    base_price_paisa: basePricePaisa,
                    tax_paisa: 0,
                    booking_fee_paisa: 0,
                    currency: currencyCode,
                    class_prices: classPrices,
                    seat_category_prices: seatCategoryPrices,
                    segment_prices: segmentPrices,
                },
                departure_minutes: departureMinutes,
                arrival_offset_minutes: 0,
                timezone,
                start_date: startDate,
                end_date: endDate,
                days_of_week: daysMask,
            };

            let schedules: CreateScheduleRequest[] = [baseSchedule];
            if (additionalTimes.length > 0) {
                schedules = [
                    baseSchedule,
                    ...additionalTimes.map((mins) => ({
                        ...baseSchedule,
                        departure_minutes: mins,
                    })),
                ];
            }

            if (schedules.length > 1) {
                const created = await catalogApi.bulkCreateSchedules(schedules);
                for (const schedule of created) {
                    await catalogApi.generateTripInstances(
                        schedule.id,
                        schedule.start_date,
                        schedule.end_date,
                    );
                }
                toast.success("Schedules created and trips generated");
            } else {
                const schedule = await catalogApi.createSchedule(baseSchedule);
                await catalogApi.generateTripInstances(
                    schedule.id,
                    schedule.start_date,
                    schedule.end_date,
                );
                toast.success("Schedule created and trips generated");
            }

            open = false;
            resetForm();
            onSuccess();
        } catch (e) {
            console.error(e);
            toast.error("Failed to schedule trips");
        } finally {
            loading = false;
        }
    }

    function resetForm() {
        routeId = "";
        assetId = "";
        departureTime = "";
        startDate = "";
        endDate = "";
        additionalDepartures = "";
        totalSeats = 40;
        basePrice = "";
        currency = "BDT";
        classPriceOverrides = [];
        seatCategoryOverrides = [];
        segmentPriceOverrides = [];
        selectedDays = new Set();
    }
</script>

<Dialog.Root bind:open>
    <Dialog.Content class="sm:max-w-[560px]">
        <Dialog.Header>
            <Dialog.Title>Create Schedule</Dialog.Title>
            <Dialog.Description>
                Set a recurring schedule and generate trip instances.
            </Dialog.Description>
        </Dialog.Header>

        <div class="grid gap-4 py-4 max-h-[70vh] overflow-y-auto px-2">
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
                    <Label for="start_date">Start Date</Label>
                    <Input id="start_date" type="date" bind:value={startDate} />
                </div>
                <div class="grid gap-2">
                    <Label for="end_date">End Date</Label>
                    <Input id="end_date" type="date" bind:value={endDate} />
                </div>
            </div>

            <div class="grid gap-2">
                <Label for="time">Departure Time</Label>
                <Input id="time" type="time" bind:value={departureTime} />
            </div>

            <div class="grid gap-2">
                <Label>Days of Week</Label>
                <div class="flex flex-wrap gap-3">
                    {#each weekdayOptions as day}
                        <div class="flex items-center gap-2">
                            <Checkbox.Root
                                id={`day-${day.bit}`}
                                checked={selectedDays.has(day.bit)}
                                onCheckedChange={(checked) => {
                                    if (checked) selectedDays.add(day.bit);
                                    else selectedDays.delete(day.bit);
                                    selectedDays = new Set(selectedDays);
                                }}
                            />
                            <Label for={`day-${day.bit}`}>{day.label}</Label>
                        </div>
                    {/each}
                </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                    <Label for="class">Vehicle Class</Label>
                    <Input id="class" bind:value={vehicleClass} />
                </div>
                <div class="grid gap-2">
                    <Label for="seats">Total Seats</Label>
                    <Input
                        id="seats"
                        type="number"
                        bind:value={totalSeats}
                        min={1}
                    />
                </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                    <Label for="base_price">Base Price</Label>
                    <Input
                        id="base_price"
                        type="number"
                        min="0"
                        step="0.01"
                        placeholder="1000"
                        bind:value={basePrice}
                    />
                </div>
                <div class="grid gap-2">
                    <Label for="currency">Currency</Label>
                    <Input
                        id="currency"
                        placeholder="BDT"
                        bind:value={currency}
                        disabled
                    />
                </div>
            </div>

            <div class="grid gap-2">
                <Label>Class Price Overrides</Label>
                <div class="grid gap-3">
                    {#each classPriceOverrides as entry, index}
                        <div class="grid grid-cols-2 gap-3">
                            <Input
                                placeholder="Class (e.g. economy)"
                                bind:value={entry.key}
                            />
                            <Input
                                type="number"
                                min="0"
                                step="0.01"
                                placeholder="Price"
                                bind:value={entry.value}
                            />
                        </div>
                    {/each}
                    <Button
                        type="button"
                        variant="outline"
                        onclick={() => {
                            classPriceOverrides = [
                                ...classPriceOverrides,
                                { key: "", value: "" },
                            ];
                        }}
                    >
                        Add Class Override
                    </Button>
                </div>
            </div>

            <div class="grid gap-2">
                <Label>Seat Category Overrides</Label>
                <div class="grid gap-3">
                    {#each seatCategoryOverrides as entry}
                        <div class="grid grid-cols-2 gap-3">
                            <Input
                                placeholder="Category (e.g. VIP)"
                                bind:value={entry.key}
                            />
                            <Input
                                type="number"
                                min="0"
                                step="0.01"
                                placeholder="Price"
                                bind:value={entry.value}
                            />
                        </div>
                    {/each}
                    <Button
                        type="button"
                        variant="outline"
                        onclick={() => {
                            seatCategoryOverrides = [
                                ...seatCategoryOverrides,
                                { key: "", value: "" },
                            ];
                        }}
                    >
                        Add Seat Category
                    </Button>
                </div>
            </div>

            <div class="grid gap-2">
                <Label>Segment Price Overrides</Label>
                <div class="grid gap-3">
                    {#if segmentPriceOverrides.length === 0}
                        <p class="text-sm text-muted-foreground">
                            Select a route to configure per-segment pricing.
                        </p>
                    {:else}
                        {#each segmentPriceOverrides as segment}
                            <div class="grid grid-cols-2 gap-3">
                                <div class="flex items-center text-sm text-muted-foreground">
                                    {stationNameMap[segment.fromId] || segment.fromId} →
                                    {stationNameMap[segment.toId] || segment.toId}
                                </div>
                                <Input
                                    type="number"
                                    min="0"
                                    step="0.01"
                                    placeholder="Base price"
                                    bind:value={segment.basePrice}
                                />
                            </div>
                        {/each}
                    {/if}
                </div>
            </div>

            <div class="grid gap-2">
                <Label for="bulk_times"
                    >Additional Departures (HH:MM, comma separated)</Label
                >
                <Input
                    id="bulk_times"
                    placeholder="09:30, 13:00, 18:45"
                    bind:value={additionalDepartures}
                />
            </div>
        </div>

        <Dialog.Footer>
            <Button variant="outline" onclick={() => (open = false)}
                >Cancel</Button
            >
            <Button onclick={handleSubmit} disabled={loading}>
                {loading ? "Scheduling..." : "Create Schedule"}
            </Button>
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
