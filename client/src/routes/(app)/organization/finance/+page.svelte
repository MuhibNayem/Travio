<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";
    import {
        Wallet,
        TrendingUp,
        CreditCard,
        ArrowUpRight,
        ArrowDownRight,
        Download,
        BarChart3,
        Loader2,
    } from "@lucide/svelte";
    import { toast } from "svelte-sonner";

    let financeStats = $state({
        totalRevenue: 1250000,
        totalPayouts: 980000,
        pendingPayouts: 125000,
        platformFees: 62500,
        revenueGrowth: "+15.3%",
        payoutGrowth: "+12.1%",
    });

    let paymentBreakdown = $state([
        { method: "SSLCommerz", amount: 625000, percentage: 50, color: "bg-blue-500" },
        { method: "bKash", amount: 375000, percentage: 30, color: "bg-pink-500" },
        { method: "Nagad", amount: 125000, percentage: 10, color: "bg-orange-500" },
        { method: "Cash", amount: 125000, percentage: 10, color: "bg-green-500" },
    ]);

    let recentTransactions = $state([
        { id: "TXN-001", type: "payment", amount: 1600, method: "bKash", status: "completed", date: "2026-04-07 10:30" },
        { id: "TXN-002", type: "payout", amount: -12500, method: "Bank Transfer", status: "pending", date: "2026-04-07 09:15" },
        { id: "TXN-003", type: "payment", amount: 800, method: "SSLCommerz", status: "completed", date: "2026-04-06 16:45" },
        { id: "TXN-004", type: "refund", amount: -800, method: "SSLCommerz", status: "completed", date: "2026-04-06 14:20" },
    ]);

    let isLoading = $state(false);

    function downloadReport() {
        toast.info("Report download started", {
            description: "Your report will be ready shortly",
        });
    }

    function getTransactionIcon(type: string) {
        switch (type) {
            case "payment":
                return ArrowUpRight;
            case "payout":
                return ArrowDownRight;
            case "refund":
                return ArrowDownRight;
            default:
                return ArrowUpRight;
        }
    }

    function getTransactionColor(type: string) {
        switch (type) {
            case "payment":
                return "text-green-600";
            case "payout":
                return "text-blue-600";
            case "refund":
                return "text-red-600";
            default:
                return "text-muted-foreground";
        }
    }
</script>

<div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight">Finance</h1>
            <p class="mt-2 text-muted-foreground">
                Revenue, payouts, and financial reports
            </p>
        </div>
        <Button variant="outline" onclick={downloadReport}>
            <Download class="mr-2 h-4 w-4" />
            Export Report
        </Button>
    </div>

    <!-- Stats Grid -->
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div class="glass-card rounded-xl p-6">
            <div class="flex items-center justify-between">
                <div>
                    <p class="text-sm font-medium text-muted-foreground">
                        Total Revenue
                    </p>
                    <h3 class="mt-2 text-2xl font-bold">
                        ৳{financeStats.totalRevenue.toLocaleString()}
                    </h3>
                </div>
                <div class="rounded-full bg-green-100 p-3 text-green-600 dark:bg-green-900/30 dark:text-green-400">
                    <Wallet size={20} />
                </div>
            </div>
            <div class="mt-4 flex items-center text-xs text-green-500">
                <TrendingUp size={14} class="mr-1" />
                <span class="font-medium">{financeStats.revenueGrowth}</span>
                <span class="ml-1 text-muted-foreground">vs last month</span>
            </div>
        </div>

        <div class="glass-card rounded-xl p-6">
            <div class="flex items-center justify-between">
                <div>
                    <p class="text-sm font-medium text-muted-foreground">
                        Total Payouts
                    </p>
                    <h3 class="mt-2 text-2xl font-bold">
                        ৳{financeStats.totalPayouts.toLocaleString()}
                    </h3>
                </div>
                <div class="rounded-full bg-blue-100 p-3 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400">
                    <ArrowDownRight size={20} />
                </div>
            </div>
            <div class="mt-4 flex items-center text-xs text-blue-500">
                <TrendingUp size={14} class="mr-1" />
                <span class="font-medium">{financeStats.payoutGrowth}</span>
                <span class="ml-1 text-muted-foreground">vs last month</span>
            </div>
        </div>

        <div class="glass-card rounded-xl p-6">
            <div class="flex items-center justify-between">
                <div>
                    <p class="text-sm font-medium text-muted-foreground">
                        Pending Payouts
                    </p>
                    <h3 class="mt-2 text-2xl font-bold">
                        ৳{financeStats.pendingPayouts.toLocaleString()}
                    </h3>
                </div>
                <div class="rounded-full bg-yellow-100 p-3 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-400">
                    <CreditCard size={20} />
                </div>
            </div>
            <div class="mt-4 text-xs text-muted-foreground">
                Next payout: April 15, 2026
            </div>
        </div>

        <div class="glass-card rounded-xl p-6">
            <div class="flex items-center justify-between">
                <div>
                    <p class="text-sm font-medium text-muted-foreground">
                        Platform Fees
                    </p>
                    <h3 class="mt-2 text-2xl font-bold">
                        ৳{financeStats.platformFees.toLocaleString()}
                    </h3>
                </div>
                <div class="rounded-full bg-purple-100 p-3 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400">
                    <BarChart3 size={20} />
                </div>
            </div>
            <div class="mt-4 text-xs text-muted-foreground">
                5% commission rate
            </div>
        </div>
    </div>

    <!-- Main Content -->
    <div class="grid gap-6 lg:grid-cols-3">
        <!-- Payment Method Breakdown -->
        <div class="glass-card rounded-xl p-6">
            <h3 class="mb-4 text-lg font-bold">Payment Methods</h3>
            <div class="space-y-4">
                {#each paymentBreakdown as item}
                    <div>
                        <div class="mb-1 flex justify-between text-sm">
                            <span class="font-medium">{item.method}</span>
                            <span class="text-muted-foreground">
                                ৳{item.amount.toLocaleString()} ({item.percentage}%)
                            </span>
                        </div>
                        <div class="h-2 w-full overflow-hidden rounded-full bg-muted">
                            <div
                                class="h-full {item.color}"
                                style="width: {item.percentage}%"
                            ></div>
                        </div>
                    </div>
                {/each}
            </div>
        </div>

        <!-- Recent Transactions -->
        <div class="glass-card col-span-2 rounded-xl p-6 lg:col-span-2">
            <div class="mb-4 flex items-center justify-between">
                <h3 class="text-lg font-bold">Recent Transactions</h3>
                <Button variant="ghost" size="sm">View All</Button>
            </div>
            <div class="overflow-x-auto">
                <table class="w-full text-sm">
                    <thead>
                        <tr class="border-b border-border">
                            <th class="pb-3 text-left font-medium text-muted-foreground">ID</th>
                            <th class="pb-3 text-left font-medium text-muted-foreground">Type</th>
                            <th class="pb-3 text-left font-medium text-muted-foreground">Method</th>
                            <th class="pb-3 text-left font-medium text-muted-foreground">Amount</th>
                            <th class="pb-3 text-left font-medium text-muted-foreground">Status</th>
                            <th class="pb-3 text-left font-medium text-muted-foreground">Date</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each recentTransactions as txn}
                            {@const Icon = getTransactionIcon(txn.type)}
                            <tr class="border-b border-border/50 last:border-0">
                                <td class="py-3">
                                    <span class="flex items-center gap-2 capitalize {getTransactionColor(txn.type)}">
                                        <Icon size={14} />
                                        {txn.type}
                                    </span>
                                </td>
                                <td class="py-3">{txn.method}</td>
                                <td class="py-3 font-bold {txn.amount > 0 ? 'text-green-600' : 'text-red-600'}">
                                    {txn.amount > 0 ? '+' : ''}৳{Math.abs(txn.amount)}
                                </td>
                                <td class="py-3">
                                    <span class="rounded-full px-2 py-0.5 text-xs font-semibold {txn.status === 'completed' ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' : txn.status === 'pending' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400' : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'}">
                                        {txn.status}
                                    </span>
                                </td>
                                <td class="py-3 text-muted-foreground">{txn.date}</td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</div>
