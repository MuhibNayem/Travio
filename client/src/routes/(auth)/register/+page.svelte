<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";

    let name = $state("");
    let email = $state("");
    let password = $state("");

    function handleRegister() {
        alert("Registration not implemented yet");
    }
</script>

<div class="glass-panel w-full p-8 relative overflow-hidden">
    <div
        class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-purple-500 to-pink-500"
    ></div>

    <div class="flex flex-col gap-6 text-center">
        <div>
            <h1 class="text-3xl font-black tracking-tight mb-2">
                Join TicketNation
            </h1>
            <p class="text-muted-foreground">Create your account today</p>
        </div>

        <div class="flex flex-col gap-4 text-left">
            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    >Full Name</label
                >
                <Input
                    type="text"
                    bind:value={name}
                    class="bg-white/50 backdrop-blur-sm"
                />
            </div>
            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    >Email</label
                >
                <Input
                    type="email"
                    bind:value={email}
                    class="bg-white/50 backdrop-blur-sm"
                />
            </div>
            <div class="space-y-2">
                <label
                    class="text-sm font-bold text-gray-700 dark:text-gray-300"
                    >Password</label
                >
                <Input
                    type="password"
                    bind:value={password}
                    class="bg-white/50 backdrop-blur-sm"
                />
            </div>
        </div>

        <Button
            class="w-full h-12 text-lg font-bold shadow-lg shadow-purple-500/20 bg-purple-600 hover:bg-purple-700"
            onclick={handleRegister}
        >
            Create Account
        </Button>

        <p class="text-sm text-gray-500">
            Already have an account? <a
                href="/login"
                class="font-bold text-primary hover:underline">Sign In</a
            >
        </p>
    </div>
</div>
