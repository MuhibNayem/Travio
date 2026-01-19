import { browser } from '$app/environment';
import { catalogApi, type Station } from '$lib/api/catalog';

/**
 * Stations Store - Svelte 5 Runes Pattern
 * Provides reactive state management for stations data
 */
class StationsStore {
    stations = $state<Station[]>([]);
    stationMap = $state<Record<string, Station>>({});

    // Pagination & Filtering State
    visibleStations = $state<Station[]>([]);
    currentQuery = $state("");
    currentPage = $state(1);
    pageSize = 50;
    loading = $state<boolean>(false);
    loadingMore = $state<boolean>(false);
    hasMore = $state(true);

    error = $state<string | null>(null);
    private cacheKey = 'stations:byId';

    /**
     * Load stations from API (initial load)
     */
    async load(forceRefresh = false): Promise<Station[]> {
        if (this.stations.length > 0 && !forceRefresh) {
            return this.stations; // Return cached initial set
        }

        this.loading = true;
        this.error = null;
        this.currentPage = 1;
        this.hasMore = true;
        this.currentQuery = "";

        try {
            // Initial fetch
            const stations = await catalogApi.getStations({ page_size: this.pageSize });

            this.stations = stations; // Keep initial cache? Or just use visibleStations? 
            // Let's keep stations for "default" view, but visibleStations drives the UI
            this.visibleStations = stations;

            this.stationMap = stations.reduce<Record<string, Station>>((acc, s) => { acc[s.id] = s; return acc; }, {});

            this.hasMore = stations.length === this.pageSize;
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
     * specific search method that hits the API
     */
    async handleSearch(query: string) {
        this.currentQuery = query;
        this.currentPage = 1;
        this.loading = true;
        // Note: We might want a separate "searching" state to avoid full page loader, 
        // but for now reusing 'loading' or we can add 'isSearching' if needed. 
        // SearchHero uses 'loading' prop on Combobox which shows a spinner.

        try {
            const stations = await catalogApi.getStations({
                search_query: query,
                page_size: this.pageSize
            });

            this.visibleStations = stations;
            this.hasMore = stations.length === this.pageSize;

            // Merge into stationMap so GetById still works for these new ones
            stations.forEach(s => this.stationMap[s.id] = s);

        } catch (err) {
            console.error("Search failed", err);
            this.error = "Search failed";
        } finally {
            this.loading = false;
        }
    }

    /**
     * Load next page of data
     */
    async loadMore() {
        if (this.loadingMore || !this.hasMore) return;

        this.loadingMore = true;

        try {
            // We need to implement page_token support in Store or just use offset.
            // The API supports page_size and page_token (which is offset as string).
            // Current page is 1-based. Next offset = currentPage * pageSize
            const offset = this.currentPage * this.pageSize;

            const moreStations = await catalogApi.getStations({
                search_query: this.currentQuery,
                page_size: this.pageSize,
                page_token: offset.toString()
            });

            if (moreStations.length > 0) {
                this.visibleStations = [...this.visibleStations, ...moreStations];
                this.currentPage++;

                moreStations.forEach(s => this.stationMap[s.id] = s);

                this.hasMore = moreStations.length === this.pageSize;
            } else {
                this.hasMore = false;
            }
        } catch (err) {
            console.error("Load more failed", err);
        } finally {
            this.loadingMore = false;
        }
    }

    // updateVisibleStations is no longer needed/used in this pattern, removed logic.
    updateVisibleStations(query: string = "", reset: boolean = false) {
        // Legacy/Unused stub to satisfy any lingering calls until cleanup? 
        // Actually best to remove it, but check callsites. 
        // handleSearch replaced it. load() calls it? No, I removed that call in load() above.
    }

    /**
     * Resolve a station by ID with in-memory + localStorage caching.
     */
    /**
     * Resolve a station by ID.
     */
    async getStationById(id: string): Promise<Station | null> {
        if (!id) return null;

        const existing = this.stationMap[id];
        if (existing) return existing;

        try {
            const station = await catalogApi.getStation(id);
            this.stationMap = { ...this.stationMap, [id]: station };
            return station;
        } catch {
            return null;
        }
    }

    /**
     * Clear store (reset)
     */
    clear(): void {
        this.stations = [];
        this.stationMap = {};
        this.visibleStations = [];
        this.loading = false;
        this.error = null;
    }


}

// Export singleton instance
export const stationsStore = new StationsStore();
