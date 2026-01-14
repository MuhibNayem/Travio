<script lang="ts">
    import { onMount } from "svelte";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import { toast } from "svelte-sonner";
    import { getOrganization, updateOrganization, type OrganizationProfile } from "$lib/api/auth";
    import { pricingApi, type PricingRule, type CreatePricingRuleRequest } from "$lib/api/pricing";
    import { catalogApi, type Station } from "$lib/api/catalog";
    import * as Dialog from "$lib/components/ui/dialog";

    let organization: OrganizationProfile | null = null;
    let loading = true;
    let saving = false;
    let rulesLoading = true;
    let rulesSaving = false;
    let rules: PricingRule[] = [];
    let showRuleModal = false;
    let editingRule: PricingRule | null = null;
    let stations: Station[] = [];

    let name = "";
    let address = "";
    let phone = "";
    let email = "";
    let website = "";
    let currency = "BDT";

    let ruleName = "";
    let ruleDescription = "";
    let ruleType = "time_window";
    let ruleCondition = "";
    let rulePriority = 10;
    let adjustmentType = "multiplier";
    let multiplierValue = "1.0";
    let adjustmentValue = "0";
    let ruleActive = true;
    let ruleDays = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"];
    const ruleTypeOptions = [
        { value: "time_window", label: "Time Window" },
        { value: "advance_purchase", label: "Advance Purchase" },
        { value: "occupancy", label: "Occupancy" },
        { value: "segment", label: "Segment" },
        { value: "promo", label: "Promo Code" },
        { value: "seat_category", label: "Seat Category" },
        { value: "vehicle_class", label: "Vehicle Class" },
        { value: "custom", label: "Custom" },
    ];
    let selectedDays = new Set<string>();
    let startHour = "8";
    let endHour = "18";
    let minDays = "0";
    let maxDays = "30";
    let occupancyThreshold = "0.8";
    let segmentFrom = "";
    let segmentTo = "";
    let promoCode = "";
    let seatCategory = "";
    let vehicleClass = "";

    onMount(async () => {
        try {
            organization = await getOrganization();
            name = organization.name || "";
            address = organization.address || "";
            phone = organization.phone || "";
            email = organization.email || "";
            website = organization.website || "";
            currency = organization.currency || "BDT";
            rules = await pricingApi.listRules(true);
            stations = await catalogApi.getStations();
        } catch (error) {
            console.error(error);
            toast.error("Failed to load organization settings");
        } finally {
            loading = false;
            rulesLoading = false;
        }
    });

    async function handleSave() {
        saving = true;
        try {
            organization = await updateOrganization({
                name,
                address,
                phone,
                email,
                website,
                currency,
            });
            toast.success("Organization updated");
        } catch (error) {
            console.error(error);
            toast.error("Failed to update organization");
        } finally {
            saving = false;
        }
    }

    function resetRuleForm() {
        ruleName = "";
        ruleDescription = "";
        ruleType = "time_window";
        ruleCondition = "";
        rulePriority = 10;
        adjustmentType = "multiplier";
        multiplierValue = "1.0";
        adjustmentValue = "0";
        ruleActive = true;
        selectedDays = new Set();
        startHour = "8";
        endHour = "18";
        minDays = "0";
        maxDays = "30";
        occupancyThreshold = "0.8";
        segmentFrom = "";
        segmentTo = "";
        promoCode = "";
        seatCategory = "";
        vehicleClass = "";
    }

    function buildCondition(): string {
        switch (ruleType) {
            case "time_window": {
                const days = Array.from(selectedDays);
                const dayCondition = days.length
                    ? days.map((d) => `day_of_week == "${d}"`).join(" || ")
                    : "true";
                return `(${dayCondition}) && hour >= ${Number(startHour)} && hour < ${Number(endHour)}`;
            }
            case "advance_purchase":
                return `days_until_departure >= ${Number(minDays)} && days_until_departure <= ${Number(maxDays)}`;
            case "occupancy":
                return `occupancy_rate >= ${Number(occupancyThreshold)}`;
            case "segment":
                return `from_station_id == "${segmentFrom}" && to_station_id == "${segmentTo}"`;
            case "promo":
                return `promo_code == "${promoCode}"`;
            case "seat_category":
                return `seat_category == "${seatCategory}"`;
            case "vehicle_class":
                return `vehicle_class == "${vehicleClass}"`;
            default:
                return ruleCondition;
        }
    }

    function openCreateRule() {
        editingRule = null;
        resetRuleForm();
        showRuleModal = true;
    }

    function openEditRule(rule: PricingRule) {
        editingRule = rule;
        ruleName = rule.name;
        ruleDescription = rule.description || "";
        ruleType = "custom";
        ruleCondition = rule.condition;
        rulePriority = rule.priority || 0;
        adjustmentType = rule.adjustment_type || "multiplier";
        multiplierValue = String(rule.multiplier || 1);
        adjustmentValue = String(rule.adjustment_value || 0);
        ruleActive = rule.is_active;
        showRuleModal = true;
    }

    async function handleSaveRule() {
        if (ruleType === "segment" && (!segmentFrom || !segmentTo)) {
            toast.error("Select both segment stations");
            return;
        }
        if (ruleType === "promo" && !promoCode) {
            toast.error("Promo code is required");
            return;
        }
        if (ruleType === "seat_category" && !seatCategory) {
            toast.error("Seat category is required");
            return;
        }
        if (ruleType === "vehicle_class" && !vehicleClass) {
            toast.error("Vehicle class is required");
            return;
        }

        rulesSaving = true;
        try {
            const condition = buildCondition();
            const payload: CreatePricingRuleRequest = {
                name: ruleName,
                description: ruleDescription,
                condition,
                multiplier: Number(multiplierValue) || 1,
                adjustment_type: adjustmentType,
                adjustment_value: Number(adjustmentValue) || 0,
                priority: Number(rulePriority) || 0,
            };
            if (editingRule) {
                const updated = await pricingApi.updateRule(editingRule.id, {
                    ...payload,
                    is_active: ruleActive,
                });
                rules = rules.map((r) => (r.id === updated.id ? updated : r));
            } else {
                const created = await pricingApi.createRule(payload);
                rules = [created, ...rules];
            }
            toast.success("Pricing rule saved");
            showRuleModal = false;
        } catch (error) {
            console.error(error);
            toast.error("Failed to save pricing rule");
        } finally {
            rulesSaving = false;
        }
    }

    async function handleDeleteRule(ruleId: string) {
        try {
            await pricingApi.deleteRule(ruleId);
            rules = rules.filter((rule) => rule.id !== ruleId);
            toast.success("Pricing rule deleted");
        } catch (error) {
            console.error(error);
            toast.error("Failed to delete pricing rule");
        }
    }
</script>

<div class="glass-card rounded-xl p-6">
    <h2 class="text-2xl font-bold">Organization Settings</h2>
    <p class="mt-2 text-muted-foreground">
        Manage organization profile, currency, and operational policies here.
    </p>

    {#if loading}
        <div class="mt-6 text-muted-foreground">Loading...</div>
    {:else}
        <div class="mt-6 grid gap-4">
            <div class="grid gap-2">
                <Label for="org_name">Organization Name</Label>
                <Input id="org_name" bind:value={name} placeholder="Organization Name" />
            </div>
            <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                    <Label for="org_phone">Phone</Label>
                    <Input id="org_phone" bind:value={phone} placeholder="+880..." />
                </div>
                <div class="grid gap-2">
                    <Label for="org_email">Email</Label>
                    <Input id="org_email" bind:value={email} placeholder="support@example.com" />
                </div>
            </div>
            <div class="grid gap-2">
                <Label for="org_address">Address</Label>
                <Input id="org_address" bind:value={address} placeholder="Street, city" />
            </div>
            <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                    <Label for="org_website">Website</Label>
                    <Input id="org_website" bind:value={website} placeholder="https://example.com" />
                </div>
                <div class="grid gap-2">
                    <Label for="org_currency">Currency</Label>
                    <Input id="org_currency" bind:value={currency} placeholder="BDT" maxlength={3} />
                </div>
            </div>
            <div class="flex justify-end">
                <Button onclick={handleSave} disabled={saving}>
                    {saving ? "Saving..." : "Save Changes"}
                </Button>
            </div>
        </div>

        <div class="mt-10 border-t border-border pt-8">
            <div class="flex items-center justify-between">
                <div>
                    <h3 class="text-xl font-semibold">Pricing Rules</h3>
                    <p class="text-sm text-muted-foreground">
                        Configure time, segment, promo, and demand-based adjustments.
                    </p>
                </div>
                <Button onclick={openCreateRule}>Create Rule</Button>
            </div>

            {#if rulesLoading}
                <div class="mt-4 text-muted-foreground">Loading pricing rules...</div>
            {:else if rules.length === 0}
                <div class="mt-4 text-muted-foreground">No pricing rules configured.</div>
            {:else}
                <div class="mt-4 grid gap-3">
                    {#each rules as rule}
                        <div class="rounded-lg border bg-card p-4">
                            <div class="flex items-center justify-between">
                                <div>
                                    <p class="font-semibold">{rule.name}</p>
                                    <p class="text-xs text-muted-foreground">
                                        {rule.description || "No description"}
                                    </p>
                                    <p class="text-xs text-muted-foreground mt-1">
                                        {rule.condition}
                                    </p>
                                </div>
                                <div class="flex items-center gap-2">
                                    <Button variant="outline" onclick={() => openEditRule(rule)}>
                                        Edit
                                    </Button>
                                    <Button variant="destructive" onclick={() => handleDeleteRule(rule.id)}>
                                        Delete
                                    </Button>
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    {/if}
</div>

<Dialog.Root bind:open={showRuleModal}>
    <Dialog.Content class="sm:max-w-[620px]">
        <Dialog.Header>
            <Dialog.Title>{editingRule ? "Edit Pricing Rule" : "Create Pricing Rule"}</Dialog.Title>
            <Dialog.Description>
                Define rule conditions and pricing adjustments.
            </Dialog.Description>
        </Dialog.Header>

        <div class="grid gap-4 py-4">
            <div class="grid gap-2">
                <Label for="rule_name">Rule Name</Label>
                <Input id="rule_name" bind:value={ruleName} placeholder="Weekend Surge" />
            </div>
            <div class="grid gap-2">
                <Label for="rule_desc">Description</Label>
                <Input id="rule_desc" bind:value={ruleDescription} placeholder="Optional description" />
            </div>
            <div class="grid gap-2">
                <Label for="rule_type">Rule Type</Label>
                <select id="rule_type" class="rounded-md border px-3 py-2" bind:value={ruleType}>
                    {#each ruleTypeOptions as option}
                        <option value={option.value}>{option.label}</option>
                    {/each}
                </select>
            </div>

            {#if ruleType === "time_window"}
                <div class="grid gap-2">
                    <Label>Days</Label>
                    <div class="flex flex-wrap gap-2">
                        {#each ruleDays as day}
                            <Button
                                type="button"
                                variant={selectedDays.has(day) ? "default" : "outline"}
                                onclick={() => {
                                    if (selectedDays.has(day)) selectedDays.delete(day);
                                    else selectedDays.add(day);
                                    selectedDays = new Set(selectedDays);
                                }}
                            >
                                {day.slice(0, 3)}
                            </Button>
                        {/each}
                    </div>
                </div>
                <div class="grid grid-cols-2 gap-4">
                    <div class="grid gap-2">
                        <Label>Start Hour</Label>
                        <Input type="number" min="0" max="23" bind:value={startHour} />
                    </div>
                    <div class="grid gap-2">
                        <Label>End Hour</Label>
                        <Input type="number" min="0" max="23" bind:value={endHour} />
                    </div>
                </div>
            {:else if ruleType === "advance_purchase"}
                <div class="grid grid-cols-2 gap-4">
                    <div class="grid gap-2">
                        <Label>Minimum Days</Label>
                        <Input type="number" min="0" bind:value={minDays} />
                    </div>
                    <div class="grid gap-2">
                        <Label>Maximum Days</Label>
                        <Input type="number" min="0" bind:value={maxDays} />
                    </div>
                </div>
            {:else if ruleType === "occupancy"}
                <div class="grid gap-2">
                    <Label>Occupancy Threshold (0-1)</Label>
                    <Input type="number" min="0" max="1" step="0.01" bind:value={occupancyThreshold} />
                </div>
            {:else if ruleType === "segment"}
                <div class="grid grid-cols-2 gap-4">
                    <div class="grid gap-2">
                        <Label>From Station</Label>
                        <select class="rounded-md border px-3 py-2" bind:value={segmentFrom}>
                            <option value="">Select</option>
                            {#each stations as station}
                                <option value={station.id}>{station.name}</option>
                            {/each}
                        </select>
                    </div>
                    <div class="grid gap-2">
                        <Label>To Station</Label>
                        <select class="rounded-md border px-3 py-2" bind:value={segmentTo}>
                            <option value="">Select</option>
                            {#each stations as station}
                                <option value={station.id}>{station.name}</option>
                            {/each}
                        </select>
                    </div>
                </div>
            {:else if ruleType === "promo"}
                <div class="grid gap-2">
                    <Label>Promo Code</Label>
                    <Input bind:value={promoCode} placeholder="EID20" />
                </div>
            {:else if ruleType === "seat_category"}
                <div class="grid gap-2">
                    <Label>Seat Category</Label>
                    <Input bind:value={seatCategory} placeholder="VIP" />
                </div>
            {:else if ruleType === "vehicle_class"}
                <div class="grid gap-2">
                    <Label>Vehicle Class</Label>
                    <Input bind:value={vehicleClass} placeholder="business" />
                </div>
            {:else}
                <div class="grid gap-2">
                    <Label>Condition</Label>
                    <Input bind:value={ruleCondition} placeholder='seat_class == "business"' />
                </div>
            {/if}

            <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                    <Label>Adjustment Type</Label>
                    <select class="rounded-md border px-3 py-2" bind:value={adjustmentType}>
                        <option value="multiplier">Multiplier</option>
                        <option value="additive">Additive (paisa)</option>
                        <option value="override">Override (paisa)</option>
                    </select>
                </div>
                <div class="grid gap-2">
                    <Label>Priority</Label>
                    <Input type="number" bind:value={rulePriority} />
                </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                    <Label>Multiplier</Label>
                    <Input type="number" step="0.01" bind:value={multiplierValue} disabled={adjustmentType !== "multiplier"} />
                </div>
                <div class="grid gap-2">
                    <Label>Adjustment Value</Label>
                    <Input type="number" step="0.01" bind:value={adjustmentValue} disabled={adjustmentType === "multiplier"} />
                </div>
            </div>
            <div class="grid gap-2">
                <Label>Active</Label>
                <input type="checkbox" bind:checked={ruleActive} />
            </div>
        </div>

        <Dialog.Footer>
            <Button variant="outline" onclick={() => (showRuleModal = false)}>Cancel</Button>
            <Button onclick={handleSaveRule} disabled={rulesSaving}>
                {rulesSaving ? "Saving..." : "Save Rule"}
            </Button>
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
