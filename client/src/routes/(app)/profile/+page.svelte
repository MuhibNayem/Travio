<script lang="ts">
    import { auth } from "$lib/runes/auth.svelte";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import {
        User,
        Mail,
        Phone,
        Shield,
        Monitor,
        Clock,
        Loader2,
        Check,
        AlertCircle,
        LogOut,
        Trash2,
    } from "@lucide/svelte";
    import { toast } from "svelte-sonner";
    import { goto } from "$app/navigation";

    let name = $state("");
    let email = $state("");
    let phone = $state("");
    let currentPassword = $state("");
    let newPassword = $state("");
    let confirmPassword = $state("");
    let isUpdating = $state(false);
    let isChangingPassword = $state(false);
    let sessions = $state<any[]>([]);
    let isLoadingSessions = $state(false);

    async function fetchProfile() {
        try {
            // Get user details from auth store
            name = auth.user?.name || "";
            email = auth.user?.email || "";
            phone = auth.user?.phone || "";
        } catch (error) {
            toast.error("Failed to load profile");
        }
    }

    async function fetchSessions() {
        isLoadingSessions = true;
        try {
            // Mock sessions - in production, call API
            sessions = [
                {
                    id: "1",
                    device: "Chrome on macOS",
                    ip: "192.168.1.1",
                    location: "Dhaka, Bangladesh",
                    last_active: "2 minutes ago",
                    current: true,
                },
                {
                    id: "2",
                    device: "Safari on iPhone",
                    ip: "192.168.1.100",
                    location: "Chittagong, Bangladesh",
                    last_active: "1 day ago",
                    current: false,
                },
            ];
        } catch (error) {
            toast.error("Failed to load sessions");
        } finally {
            isLoadingSessions = false;
        }
    }

    async function updateProfile() {
        isUpdating = true;
        try {
            // Mock update - in production, call API
            await new Promise((resolve) => setTimeout(resolve, 1000));
            toast.success("Profile updated", {
                description: "Your changes have been saved",
            });
        } catch (error) {
            toast.error("Failed to update profile");
        } finally {
            isUpdating = false;
        }
    }

    async function changePassword() {
        if (newPassword !== confirmPassword) {
            toast.error("Passwords don't match");
            return;
        }
        if (newPassword.length < 8) {
            toast.error("Password must be at least 8 characters");
            return;
        }

        isChangingPassword = true;
        try {
            // Mock update - in production, call API
            await new Promise((resolve) => setTimeout(resolve, 1000));
            toast.success("Password changed", {
                description: "Please login with your new password",
            });
            currentPassword = "";
            newPassword = "";
            confirmPassword = "";
        } catch (error) {
            toast.error("Failed to change password");
        } finally {
            isChangingPassword = false;
        }
    }

    async function revokeSession(sessionId: string) {
        try {
            toast.success("Session revoked", {
                description: "That device has been logged out",
            });
            fetchSessions();
        } catch (error) {
            toast.error("Failed to revoke session");
        }
    }

    async function logoutAllDevices() {
        try {
            await auth.logout();
            toast.success("Logged out from all devices");
            goto("/login");
        } catch (error) {
            toast.error("Failed to logout");
        }
    }

    $effect(() => {
        fetchProfile();
        fetchSessions();
    });
</script>

<div class="min-h-screen bg-muted/30 py-20">
    <div class="container mx-auto max-w-3xl px-4">
        <div class="mb-8">
            <h1 class="text-3xl font-bold">Profile Settings</h1>
            <p class="mt-1 text-muted-foreground">
                Manage your account details and preferences
            </p>
        </div>

        <div class="space-y-6">
            <!-- Personal Information -->
            <div class="glass-card rounded-xl p-6">
                <div class="mb-6 flex items-center gap-3">
                    <div
                        class="flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary"
                    >
                        <User size={20} />
                    </div>
                    <h3 class="text-lg font-bold">Personal Information</h3>
                </div>

                <div class="space-y-4">
                    <div class="space-y-2">
                        <label class="text-sm font-medium">Full Name</label>
                        <div class="relative">
                            <User
                                class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                                size={16}
                            />
                            <Input
                                type="text"
                                bind:value={name}
                                class="pl-10 bg-white/50 backdrop-blur-sm"
                                placeholder="John Doe"
                            />
                        </div>
                    </div>

                    <div class="space-y-2">
                        <label class="text-sm font-medium">Email</label>
                        <div class="relative">
                            <Mail
                                class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                                size={16}
                            />
                            <Input
                                type="email"
                                bind:value={email}
                                class="pl-10 bg-white/50 backdrop-blur-sm"
                                placeholder="email@example.com"
                            />
                        </div>
                    </div>

                    <div class="space-y-2">
                        <label class="text-sm font-medium">Phone</label>
                        <div class="relative">
                            <Phone
                                class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                                size={16}
                            />
                            <Input
                                type="tel"
                                bind:value={phone}
                                class="pl-10 bg-white/50 backdrop-blur-sm"
                                placeholder="+880 1XXX-XXXXXX"
                            />
                        </div>
                    </div>

                    <Button
                        class="mt-2"
                        onclick={updateProfile}
                        disabled={isUpdating}
                    >
                        {#if isUpdating}
                            <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                        {:else}
                            <Check class="mr-2 h-4 w-4" />
                        {/if}
                        Save Changes
                    </Button>
                </div>
            </div>

            <!-- Change Password -->
            <div class="glass-card rounded-xl p-6">
                <div class="mb-6 flex items-center gap-3">
                    <div
                        class="flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary"
                    >
                        <Shield size={20} />
                    </div>
                    <h3 class="text-lg font-bold">Change Password</h3>
                </div>

                <div class="space-y-4">
                    <div class="space-y-2">
                        <label class="text-sm font-medium"
                            >Current Password</label
                        >
                        <Input
                            type="password"
                            bind:value={currentPassword}
                            placeholder="••••••••"
                            class="bg-white/50 backdrop-blur-sm"
                        />
                    </div>

                    <div class="space-y-2">
                        <label class="text-sm font-medium">New Password</label>
                        <Input
                            type="password"
                            bind:value={newPassword}
                            placeholder="••••••••"
                            class="bg-white/50 backdrop-blur-sm"
                        />
                    </div>

                    <div class="space-y-2">
                        <label class="text-sm font-medium"
                            >Confirm New Password</label
                        >
                        <Input
                            type="password"
                            bind:value={confirmPassword}
                            placeholder="••••••••"
                            class="bg-white/50 backdrop-blur-sm"
                        />
                    </div>

                    <Button
                        variant="outline"
                        onclick={changePassword}
                        disabled={isChangingPassword}
                    >
                        {#if isChangingPassword}
                            <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                        {:else}
                            <Shield class="mr-2 h-4 w-4" />
                        {/if}
                        Update Password
                    </Button>
                </div>
            </div>

            <!-- Active Sessions -->
            <div class="glass-card rounded-xl p-6">
                <div class="mb-6 flex items-center justify-between">
                    <div class="flex items-center gap-3">
                        <div
                            class="flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary"
                        >
                            <Monitor size={20} />
                        </div>
                        <h3 class="text-lg font-bold">Active Sessions</h3>
                    </div>
                    <Button
                        variant="ghost"
                        size="sm"
                        onclick={logoutAllDevices}
                    >
                        <LogOut class="mr-1 h-4 w-4" />
                        Logout All
                    </Button>
                </div>

                {#if isLoadingSessions}
                    <div class="flex items-center justify-center py-8">
                        <Loader2 class="animate-spin text-primary" />
                    </div>
                {:else if sessions.length > 0}
                    <div class="space-y-3">
                        {#each sessions as session}
                            <div
                                class="flex items-center justify-between rounded-lg bg-muted/50 p-4"
                            >
                                <div class="flex items-start gap-3">
                                    {#if session.current}
                                        <div
                                            class="mt-1 flex size-6 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30"
                                        >
                                            <Check
                                                size={14}
                                                class="text-green-600"
                                            />
                                        </div>
                                    {:else}
                                        <Monitor
                                            size={20}
                                            class="mt-0.5 text-muted-foreground"
                                        />
                                    {/if}
                                    <div>
                                        <p class="font-medium">
                                            {session.device}
                                            {#if session.current}
                                                <span class="ml-2 text-xs text-green-600"
                                                    >(Current)</span
                                                >
                                            {/if}
                                        </p>
                                        <div class="flex items-center gap-3 text-xs text-muted-foreground">
                                            <span class="flex items-center gap-1">
                                                <Clock size={12} />
                                                {session.last_active}
                                            </span>
                                            <span>{session.location}</span>
                                        </div>
                                    </div>
                                </div>
                                {#if !session.current}
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        onclick={() =>
                                            revokeSession(session.id)}
                                    >
                                        <LogOut size={16} />
                                    </Button>
                                {/if}
                            </div>
                        {/each}
                    </div>
                {:else}
                    <p class="py-8 text-center text-muted-foreground">
                        No active sessions
                    </p>
                {/if}
            </div>

            <!-- Danger Zone -->
            <div class="rounded-xl border border-red-200 bg-red-50 p-6 dark:border-red-900/50 dark:bg-red-900/20">
                <div class="mb-4 flex items-center gap-3">
                    <AlertCircle size={20} class="text-red-600" />
                    <h3 class="text-lg font-bold text-red-700 dark:text-red-400">
                        Danger Zone
                    </h3>
                </div>
                <p class="mb-4 text-sm text-red-600 dark:text-red-500">
                    Once you delete your account, there is no going back. Please
                    be certain.
                </p>
                <Button variant="destructive" size="sm">
                    <Trash2 class="mr-2 h-4 w-4" />
                    Delete Account
                </Button>
            </div>
        </div>
    </div>
</div>
