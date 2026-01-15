<script lang="ts">
    import { MapPin, Calendar, Clock, Tag } from "@lucide/svelte";
    import { Button } from "$lib/components/ui/button";
    import type { EventSearchResult } from "$lib/api/events";

    let { result } = $props<{ result: EventSearchResult }>();

    // Fallback image if none provided
    const defaultImage =
        "https://images.unsplash.com/photo-1501281668745-f7f57925c3b4?auto=format&fit=crop&q=80&w=300&h=200";

    function formatTime(isoString: string) {
        return new Date(isoString).toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    function formatDate(isoString: string) {
        return new Date(isoString).toLocaleDateString([], {
            weekday: "short",
            month: "short",
            day: "numeric",
        });
    }
</script>

<div
    class="group relative overflow-hidden rounded-2xl border bg-white p-4 shadow-sm transition-all hover:shadow-md dark:bg-white/5 dark:border-white/10"
>
    <div class="flex flex-col md:flex-row gap-6">
        <!-- Event Image -->
        <div
            class="relative h-48 md:h-32 md:w-48 shrink-0 overflow-hidden rounded-xl bg-muted"
        >
            <img
                src={result.event.images?.[0] || defaultImage}
                alt={result.event.title}
                class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-110"
            />
            <div
                class="absolute top-2 left-2 rounded-md bg-black/60 px-2 py-1 text-xs font-bold text-white backdrop-blur-md"
            >
                {result.event.category}
            </div>
        </div>

        <!-- Event Details -->
        <div class="flex flex-1 flex-col justify-between py-1">
            <div>
                <h3
                    class="text-xl font-bold text-foreground group-hover:text-primary transition-colors"
                >
                    {result.event.title}
                </h3>
                <div
                    class="mt-2 flex flex-col gap-1 text-sm text-muted-foreground"
                >
                    <div class="flex items-center gap-2">
                        <MapPin size={14} />
                        <span>{result.venue.name}, {result.venue.city}</span>
                    </div>
                    <div class="flex items-center gap-2">
                        <Calendar size={14} />
                        <span
                            >{formatDate(result.event.start_time)} • {formatTime(
                                result.event.start_time,
                            )}</span
                        >
                    </div>
                </div>
            </div>

            <p class="mt-3 line-clamp-2 text-sm text-muted-foreground/80">
                {result.event.description}
            </p>
        </div>

        <!-- Action -->
        <div
            class="flex flex-col items-end justify-center gap-3 pt-4 md:pt-0 md:w-32 border-t md:border-t-0 border-border"
        >
            <div class="text-right">
                <span class="text-xs text-muted-foreground">Starts from</span>
                <p class="text-xl font-black text-primary">৳500</p>
            </div>
            <Button
                class="w-full rounded-lg font-bold shadow-lg shadow-primary/20 hover:shadow-primary/30 transition-all"
                size="sm"
            >
                Get Tickets
            </Button>
        </div>
    </div>
</div>
