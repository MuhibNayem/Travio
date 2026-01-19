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
    nextPageToken = $state<string>("");
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
        this.nextPageToken = "";

        try {
            // Initial fetch
            const response = await catalogApi.getStations({ page_size: this.pageSize });
            const stations = response.stations || [];

            this.stations = stations; // Keep initial cache? Or just use visibleStations? 
            // Let's keep stations for "default" view, but visibleStations drives the UI
            this.visibleStations = stations;

            this.stationMap = stations.reduce<Record<string, Station>>((acc, s) => { acc[s.id] = s; return acc; }, {});

            this.nextPageToken = response.next_page_token || "";
            this.hasMore = !!this.nextPageToken;
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
        this.nextPageToken = "";
        // Note: We might want a separate "searching" state to avoid full page loader, 
        // but for now reusing 'loading' or we can add 'isSearching' if needed. 
        // SearchHero uses 'loading' prop on Combobox which shows a spinner.

        try {
            const response = await catalogApi.getStations({
                search_query: query,
                page_size: this.pageSize
            });
            const stations = response.stations || [];

            this.visibleStations = stations;
            this.nextPageToken = response.next_page_token || "";
            this.hasMore = !!this.nextPageToken;

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
        if (this.loadingMore || !this.hasMore) {
            return;
        }

        this.loadingMore = true;

        try {
            const response = await catalogApi.getStations({
                search_query: this.currentQuery,
                page_size: this.pageSize,
                page_token: this.nextPageToken
            });

            const moreStations = response.stations || [];

            if (moreStations.length > 0) {
                this.visibleStations = [...this.visibleStations, ...moreStations];
                this.currentPage++;
                this.nextPageToken = response.next_page_token || "";
                this.hasMore = !!this.nextPageToken;

                moreStations.forEach(s => this.stationMap[s.id] = s);
            } else {
                this.hasMore = !!response.next_page_token;
                this.nextPageToken = response.next_page_token || "";
                // If items 0 but token exists, we might still have more? 
                // Usually empty items means no more, but let's trust token.
            }
        } catch (err) {
            console.error("Load more failed", err);
        } finally {
            this.loadingMore = false;
        }
    }

    /**
     * Reset to default station list (clears search filter).
     * Called when a combobox closes to restore shared state.
     */
    resetToDefault() {
        this.currentQuery = "";
        this.visibleStations = this.stations;
        this.nextPageToken = "";
        this.hasMore = this.stations.length >= this.pageSize;
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
