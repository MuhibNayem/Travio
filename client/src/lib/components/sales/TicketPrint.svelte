<script lang="ts">
    import type { Order } from "$lib/api/orders";
    import { toDataURL } from "qrcode";
    import { onMount } from "svelte";

    export let order: Order;

    let qrDataUrl = "";

    onMount(async () => {
        try {
            qrDataUrl = await toDataURL(order.id);
        } catch (err) {
            console.error(err);
        }
    });
</script>

<div
    class="ticket-container p-4 bg-white text-black font-mono w-[300px] border mx-auto"
>
    <!-- Header -->
    <div class="text-center border-b pb-4 mb-4 border-dashed border-gray-400">
        <h2 class="font-bold text-xl uppercase tracking-widest">Travio</h2>
        <p class="text-xs">Travel & Event Solutions</p>
        <p class="text-xs mt-1">{new Date().toLocaleString()}</p>
    </div>

    <!-- Trip Info -->
    <div class="mb-4">
        <div class="flex justify-between text-xs">
            <span>PNR:</span>
            <span class="font-bold"
                >{order.booking_id || order.id.substring(0, 8)}</span
            >
        </div>
        <div class="flex justify-between text-xs mt-1">
            <span>Trip ID:</span>
            <span>{order.trip_id.substring(0, 8)}</span>
        </div>
        <!-- Add Route info if we had it populated in order or fetched separately -->
    </div>

    <!-- Seats -->
    <div class="border-y py-2 my-2 border-dashed border-gray-400">
        <div class="grid grid-cols-4 gap-2 text-sm font-bold text-center">
            {#each order.passengers as p}
                <div class="border p-1">
                    {p.seat_id}
                </div>
            {/each}
        </div>
    </div>

    <!-- Passengers -->
    <div class="mb-4 space-y-1">
        {#each order.passengers as p}
            <div class="text-xs flex justify-between">
                <span>{p.name}</span>
                <span>{p.seat_id}</span>
            </div>
        {/each}
        <div
            class="text-xs flex justify-between mt-2 pt-2 border-t border-gray-200"
        >
            <span>Phone:</span>
            <span>{order.contact_phone}</span>
        </div>
    </div>

    <!-- Payment -->
    <div class="border-t pt-2 border-dashed border-gray-400 mb-4">
        <div class="flex justify-between font-bold">
            <span>Total:</span>
            <span>৳{(order.total_paisa / 100).toFixed(2)}</span>
        </div>
        <div class="flex justify-between text-xs">
            <span>Status:</span>
            <span class="uppercase">{order.payment_status}</span>
        </div>
    </div>

    <!-- QR Code -->
    <div class="flex flex-col items-center justify-center pt-2">
        {#if qrDataUrl}
            <img src={qrDataUrl} alt="QR Code" class="w-32 h-32" />
        {/if}
        <p class="text-[10px] mt-2 text-center text-gray-500">
            Scan to verify ticket
        </p>
    </div>

    <!-- Footer -->
    <div class="mt-8 text-center text-[10px] border-t pt-2">
        <p>Thank you for traveling with us!</p>
        <p>www.travio.com</p>
    </div>
</div>

<style>
    @media print {
        .ticket-container {
            border: none;
            width: 100%;
            max-width: 100%;
        }
    }
</style>
