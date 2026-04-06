<script lang="ts">
    import { auth } from "$lib/runes/auth.svelte";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import { Tabs, TabsList, TabsTrigger, TabsContent } from "$lib/components/ui/tabs";
    import {
        Building,
        CreditCard,
        Bell,
        Shield,
        Loader2,
        Check,
        Save,
    } from "@lucide/svelte";
    import { toast } from "svelte-sonner";

    let activeTab = $state("profile");

    // Profile
    let orgName = $state("Green Line Paribahan");
    let orgAddress = $state("123 Kamalapur, Dhaka-1000");
    let orgPhone = $state("+880 1712-345678");
    let orgEmail = $state("info@greenline.com");
    let orgWebsite = $state("https://greenline.com");

    // Payment Config
    let sslcommerzEnabled = $state(true);
    let bkashEnabled = $state(true);
    let nagadEnabled = $state(false);
    let sslcommerzStoreId = $state("green6789");
    let sslcommerzStorePass = $state("********");

    // Notifications
    let emailNotifications = $state(true);
    let smsNotifications = $state(true);
    let bookingAlerts = $state(true);
    let marketingEmails = $state(false);

    let isSaving = $state(false);

    async function saveProfile() {
        isSaving = true;
        try {
            await new Promise((resolve) => setTimeout(resolve, 1000));
            toast.success("Profile updated", {
                description: "Your organization details have been saved",
            });
        } catch (error) {
            toast.error("Failed to update profile");
        } finally {
            isSaving = false;
        }
    }

    async function savePaymentConfig() {
        isSaving = true;
        try {
            await new Promise((resolve) => setTimeout(resolve, 1000));
            toast.success("Payment config updated", {
                description: "Your gateway settings have been saved",
            });
        } catch (error) {
            toast.error("Failed to update payment config");
        } finally {
            isSaving = false;
        }
    }
</script>

<div class="space-y-6">
    <!-- Header -->
    <div>
        <h1 class="text-3xl font-bold tracking-tight">Settings</h1>
        <p class="mt-2 text-muted-foreground">
            Manage your organization settings and preferences
        </p>
    </div>

    <Tabs value={activeTab} onValueChange={(v) => (activeTab = v)}>
        <TabsList class="grid w-full grid-cols-4 lg:w-auto lg:grid-cols-4">
            <TabsTrigger value="profile" class="gap-2">
                <Building size={16} />
                <span class="hidden sm:inline">Profile</span>
            </TabsTrigger>
            <TabsTrigger value="payments" class="gap-2">
                <CreditCard size={16} />
                <span class="hidden sm:inline">Payments</span>
            </TabsTrigger>
            <TabsTrigger value="notifications" class="gap-2">
                <Bell size={16} />
                <span class="hidden sm:inline">Notifications</span>
            </TabsTrigger>
            <TabsTrigger value="security" class="gap-2">
                <Shield size={16} />
                <span class="hidden sm:inline">Security</span>
            </TabsTrigger>
        </TabsList>

        <!-- Profile Tab -->
        <TabsContent value="profile" class="mt-6">
            <div class="glass-card rounded-xl p-6">
                <h3 class="mb-4 text-lg font-bold">Organization Profile</h3>
                <div class="grid gap-4 sm:grid-cols-2">
                    <div class="space-y-2">
                        <label class="text-sm font-medium">Company Name</label>
                        <Input bind:value={orgName} class="bg-white/50" />
                    </div>
                    <div class="space-y-2">
                        <label class="text-sm font-medium">Phone</label>
                        <Input bind:value={orgPhone} class="bg-white/50" />
                    </div>
                    <div class="space-y-2 sm:col-span-2">
                        <label class="text-sm font-medium">Address</label>
                        <Input bind:value={orgAddress} class="bg-white/50" />
                    </div>
                    <div class="space-y-2">
                        <label class="text-sm font-medium">Email</label>
                        <Input bind:value={orgEmail} type="email" class="bg-white/50" />
                    </div>
                    <div class="space-y-2">
                        <label class="text-sm font-medium">Website</label>
                        <Input bind:value={orgWebsite} type="url" class="bg-white/50" />
                    </div>
                </div>
                <Separator class="my-6" />
                <Button onclick={saveProfile} disabled={isSaving}>
                    {#if isSaving}
                        <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                    {:else}
                        <Save class="mr-2 h-4 w-4" />
                    {/if}
                    Save Changes
                </Button>
            </div>
        </TabsContent>

        <!-- Payments Tab -->
        <TabsContent value="payments" class="mt-6">
            <div class="glass-card rounded-xl p-6">
                <h3 class="mb-4 text-lg font-bold">Payment Gateways</h3>
                <div class="space-y-4">
                    {#each [
                        { key: "sslcommerz", name: "SSLCommerz", desc: "Cards, Net Banking, Wallets", enabled: sslcommerzEnabled },
                        { key: "bkash", name: "bKash", desc: "Mobile wallet", enabled: bkashEnabled },
                        { key: "nagad", name: "Nagad", desc: "Digital financial service", enabled: nagadEnabled },
                    ] as gateway}
                        <div class="rounded-lg border border-border p-4">
                            <div class="flex items-center justify-between">
                                <div>
                                    <p class="font-semibold">{gateway.name}</p>
                                    <p class="text-sm text-muted-foreground">{gateway.desc}</p>
                                </div>
                                <label class="relative inline-flex cursor-pointer items-center">
                                    <input
                                        type="checkbox"
                                        class="peer sr-only"
                                        bind:checked={gateway.key === 'sslcommerz' ? sslcommerzEnabled : gateway.key === 'bkash' ? bkashEnabled : nagadEnabled}
                                    />
                                    <div class="peer h-6 w-11 rounded-full bg-muted after:absolute after:start-[2px] after:top-[2px] after:size-5 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all peer-checked:bg-primary peer-checked:after:translate-x-full peer-checked:after:border-white"></div>
                                </label>
                            </div>
                            {#if gateway.key === "sslcommerz" && sslcommerzEnabled}
                                <div class="mt-4 grid gap-3 sm:grid-cols-2">
                                    <div class="space-y-2">
                                        <label class="text-sm font-medium">Store ID</label>
                                        <Input bind:value={sslcommerzStoreId} class="bg-white/50" />
                                    </div>
                                    <div class="space-y-2">
                                        <label class="text-sm font-medium">Store Password</label>
                                        <Input bind:value={sslcommerzStorePass} type="password" class="bg-white/50" />
                                    </div>
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
                <Separator class="my-6" />
                <Button onclick={savePaymentConfig} disabled={isSaving}>
                    {#if isSaving}
                        <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                    {:else}
                        <Check class="mr-2 h-4 w-4" />
                    {/if}
                    Save Gateway Config
                </Button>
            </div>
        </TabsContent>

        <!-- Notifications Tab -->
        <TabsContent value="notifications" class="mt-6">
            <div class="glass-card rounded-xl p-6">
                <h3 class="mb-4 text-lg font-bold">Notification Preferences</h3>
                <div class="space-y-4">
                    {#each [
                        { label: "Email Notifications", desc: "Receive booking confirmations via email", bound: emailNotifications },
                        { label: "SMS Notifications", desc: "Get SMS alerts for important events", bound: smsNotifications },
                        { label: "Booking Alerts", desc: "Notify customers about booking status", bound: bookingAlerts },
                        { label: "Marketing Emails", desc: "Receive updates about new features", bound: marketingEmails },
                    ] as pref}
                        <div class="flex items-center justify-between rounded-lg border border-border p-4">
                            <div>
                                <p class="font-medium">{pref.label}</p>
                                <p class="text-sm text-muted-foreground">{pref.desc}</p>
                            </div>
                            <label class="relative inline-flex cursor-pointer items-center">
                                <input type="checkbox" class="peer sr-only" bind:checked={pref.bound} />
                                <div class="peer h-6 w-11 rounded-full bg-muted after:absolute after:start-[2px] after:top-[2px] after:size-5 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all peer-checked:bg-primary peer-checked:after:translate-x-full peer-checked:after:border-white"></div>
                            </label>
                        </div>
                    {/each}
                </div>
            </div>
        </TabsContent>

        <!-- Security Tab -->
        <TabsContent value="security" class="mt-6">
            <div class="glass-card rounded-xl p-6">
                <h3 class="mb-4 text-lg font-bold">Security Settings</h3>
                <div class="space-y-4">
                    <div class="rounded-lg border border-border p-4">
                        <h4 class="font-medium">Two-Factor Authentication</h4>
                        <p class="text-sm text-muted-foreground">Add an extra layer of security to your account</p>
                        <Button variant="outline" class="mt-3" disabled>Coming Soon</Button>
                    </div>
                    <div class="rounded-lg border border-border p-4">
                        <h4 class="font-medium">API Keys</h4>
                        <p class="text-sm text-muted-foreground">Manage API keys for third-party integrations</p>
                        <Button variant="outline" class="mt-3" disabled>Coming Soon</Button>
                    </div>
                    <div class="rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-900/50 dark:bg-red-900/20">
                        <h4 class="font-medium text-red-700 dark:text-red-400">Delete Organization</h4>
                        <p class="text-sm text-red-600 dark:text-red-500">This action cannot be undone</p>
                        <Button variant="destructive" class="mt-3" disabled>Contact Support</Button>
                    </div>
                </div>
            </div>
        </TabsContent>
    </Tabs>
</div>
