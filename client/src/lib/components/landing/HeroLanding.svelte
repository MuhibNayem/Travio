<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { ChevronDown } from "@lucide/svelte";
    import { auth } from "$lib/runes/auth.svelte";
</script>

<section
    class="relative min-h-[85vh] flex items-center justify-center overflow-hidden"
>
    <!-- Immersive Background -->
    <div class="absolute inset-0 z-0">
        <div
            class="absolute inset-0 bg-gradient-to-b from-black/30 via-black/10 to-[#101922] z-10"
        ></div>
        <!-- High quality travel video placeholder or image -->
        <img
            src="https://images.unsplash.com/photo-1476514525535-07fb3b4ae5f1?q=80&w=2070&auto=format&fit=crop"
            alt="Travel Hero"
            class="w-full h-full object-cover scale-105 animate-[ken-burns_20s_ease-in-out_infinite_alternate]"
        />
    </div>

    <div class="relative z-20 container px-4 mx-auto text-center">
        <!-- Floating Badge -->
        <div
            class="inline-flex items-center gap-2 px-4 py-2 mb-8 rounded-full glass-panel bg-white/10 border-white/20 backdrop-blur-md animate-fade-in-up"
        >
            <span class="relative flex h-3 w-3">
                <span
                    class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"
                ></span>
                <span
                    class="relative inline-flex rounded-full h-3 w-3 bg-green-500"
                ></span>
            </span>
            <span class="text-sm font-semibold text-white tracking-wide"
                >Live Updates 24/7</span
            >
        </div>

        <h1
            class="text-5xl md:text-8xl font-black mb-8 tracking-tight text-white drop-shadow-2xl leading-[1.1]"
        >
            Explore <br class="md:hidden" />
            <span
                class="text-transparent bg-clip-text bg-gradient-to-r from-blue-300 via-cyan-200 to-white"
                >Without Limits</span
            >
        </h1>

        <p
            class="text-lg md:text-2xl font-medium mb-12 max-w-2xl mx-auto text-blue-100/90 drop-shadow-lg leading-relaxed"
        >
            Your all-in-one gateway to trains, buses, launches, and flights.
            Experience seamless booking with military-grade security.
        </p>

        <div
            class="flex flex-col sm:flex-row gap-5 justify-center items-center"
        >
            <Button
                size="lg"
                class="h-14 px-10 text-lg font-bold rounded-2xl bg-[#137fec] hover:bg-[#0f6bc6] shadow-xl shadow-blue-500/30 hover:scale-105 transition-all duration-300"
                href={auth.isAuthenticated ? "/dashboard" : "/login"}
            >
                Book Your Trip
            </Button>
            <Button
                size="lg"
                variant="outline"
                class="glass-button h-14 px-10 text-lg font-bold border-white/20 bg-white/5 text-white hover:bg-white/10 backdrop-blur-md hover:scale-105 transition-all duration-300"
                href="#features"
            >
                Learn More
            </Button>
        </div>
    </div>

    <!-- Scroll Indicator -->
    <div
        class="absolute bottom-10 left-1/2 -translate-x-1/2 z-20 animate-bounce"
    >
        <ChevronDown class="text-white/50 w-10 h-10" />
    </div>
</section>

<style>
    @keyframes ken-burns {
        0% {
            transform: scale(1);
        }
        100% {
            transform: scale(1.1);
        }
    }
</style>
