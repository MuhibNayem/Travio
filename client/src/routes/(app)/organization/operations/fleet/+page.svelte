<script lang="ts">
    import { onMount } from "svelte";
    import {
        fleetApi,
        type Asset,
        AssetType,
        AssetStatus,
    } from "$lib/api/fleet";
    import { Button } from "$lib/components/ui/button";
    import { Plus, Bus, Train, Ship, Car, AlertCircle } from "@lucide/svelte";
    import AssetModal from "$lib/components/operations/AssetModal.svelte";
    import * as Table from "$lib/components/ui/table";
    import { toast } from "svelte-sonner";

    let assets: Asset[] = [];
    let loading = true;
    let showCreateModal = false;

    async function loadAssets() {
        loading = true;
        try {
            assets = await fleetApi.getAssets();
        } catch (e) {
            console.error(e);
            toast.error("Failed to load assets");
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        loadAssets();
    });

    const icons: Record<string, typeof Bus> = {
        [AssetType.BUS]: Bus,
        [AssetType.TRAIN]: Train,
        [AssetType.LAUNCH]: Ship,
    };
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Fleet Assets</h1>
            <p class="text-muted-foreground mt-2">
                Manage your buses, trains, and other vehicles.
            </p>
        </div>
        <Button onclick={() => (showCreateModal = true)}>
            <Plus class="mr-2 h-4 w-4" />
            Register Asset
        </Button>
    </div>

    <!-- Stats or Summary could go here -->

    <div class="rounded-xl border bg-card shadow-sm">
        <Table.Root>
            <Table.Header>
                <Table.Row>
                    <Table.Head class="w-[50px]"></Table.Head>
                    <Table.Head>NAME</Table.Head>
                    <Table.Head>LICENSE PLATE</Table.Head>
                    <Table.Head>MAKE / MODEL</Table.Head>
                    <Table.Head>YEAR</Table.Head>
                    <Table.Head class="text-right">STATUS</Table.Head>
                </Table.Row>
            </Table.Header>
            <Table.Body>
                {#if loading}
                    <Table.Row>
                        <Table.Cell colspan={6} class="h-24 text-center"
                            >Loading assets...</Table.Cell
                        >
                    </Table.Row>
                {:else if assets.length === 0}
                    <Table.Row>
                        <Table.Cell colspan={6} class="h-64 text-center">
                            <div
                                class="flex flex-col items-center justify-center text-muted-foreground"
                            >
                                <div class="p-4 rounded-full bg-muted mb-4">
                                    <Bus class="h-8 w-8" />
                                </div>
                                <p class="text-lg font-medium text-foreground">
                                    No assets found
                                </p>
                                <p class="text-sm">
                                    Register your first vehicle to get started.
                                </p>
                            </div>
                        </Table.Cell>
                    </Table.Row>
                {:else}
                    {#each assets as asset}
                        <Table.Row>
                            <Table.Cell>
                                <div
                                    class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary"
                                >
                                    <svelte:component
                                        this={icons[asset.type] || Bus}
                                        size={20}
                                    />
                                </div>
                            </Table.Cell>
                            <Table.Cell class="font-medium"
                                >{asset.name}</Table.Cell
                            >
                            <Table.Cell
                                class="text-muted-foreground font-mono bg-muted/50 px-2 py-1 rounded w-fit text-xs"
                                >{asset.license_plate}</Table.Cell
                            >
                            <Table.Cell>{asset.make} {asset.model}</Table.Cell>
                            <Table.Cell>{asset.year}</Table.Cell>
                            <Table.Cell class="text-right">
                                <span
                                    class={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold 
                                    ${
                                        asset.status === AssetStatus.ACTIVE
                                            ? "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
                                            : asset.status ===
                                                AssetStatus.MAINTENANCE
                                              ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400"
                                              : "bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400"
                                    }`}
                                >
                                    {asset.status.replace("ASSET_STATUS_", "")}
                                </span>
                            </Table.Cell>
                        </Table.Row>
                    {/each}
                {/if}
            </Table.Body>
        </Table.Root>
    </div>
</div>

<AssetModal bind:open={showCreateModal} onSuccess={loadAssets} />
