import { catalogApi, type Station } from '$lib/api/catalog';

/**
 * Stations Store - Svelte 5 Runes Pattern
 * Provides reactive state management for stations data
 */
class StationsStore {
    stations = $state<Station[]>([]);
    loading = $state<boolean>(false);
    error = $state<string | null>(null);

    /**
     * Load stations from API (with in-memory caching)
     */
    async load(forceRefresh = false): Promise<Station[]> {
        // Check if already loaded and not forcing refresh
        if (this.stations.length > 0 && !forceRefresh) {
            return this.stations;
        }

        this.loading = true;
        this.error = null;

        try {
            // Fetch from API
            const stations = await catalogApi.getStations();

            // Update state
            this.stations = stations;

            this.loading = false;
            return stations;
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Failed to load stations';
            this.error = errorMessage;
            this.loading = false;
            throw err;
        }
    }

    /**
     * Search stations by name, city, code, or division (client-side)
     */
    search(query: string): Station[] {
        if (!query) return this.stations;

        const lowerQuery = query.toLowerCase();
        return this.stations.filter(
            (station) =>
                station.name.toLowerCase().includes(lowerQuery) ||
                station.city.toLowerCase().includes(lowerQuery) ||
                station.code.toLowerCase().includes(lowerQuery) ||
                station.state?.toLowerCase().includes(lowerQuery)
        );
    }

    /**
     * Filter stations by division
     */
    filterByDivision(division: string): Station[] {
        if (!division) return this.stations;

        return this.stations.filter(
            (station) => station.state?.toLowerCase() === division.toLowerCase()
        );
    }

    /**
     * Get unique divisions
     */
    getDivisions(): string[] {
        const divisions = new Set(
            this.stations
                .map((s) => s.state)
                .filter((state): state is string => !!state)
        );
        return Array.from(divisions).sort();
    }

    /**
     * Clear store (reset)
     */
    clear(): void {
        this.stations = [];
        this.loading = false;
        this.error = null;
    }
}

// Export singleton instance
export const stationsStore = new StationsStore();
