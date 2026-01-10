import type { PageLoad } from './$types';
import { searchApi } from '$lib/api/search';

export const load: PageLoad = async ({ url }) => {
    const from = url.searchParams.get('from') || '';
    const to = url.searchParams.get('to') || '';
    const date = url.searchParams.get('date') || new Date().toISOString().split('T')[0];
    const type = url.searchParams.get('type') || 'bus';

    try {
        const response = await searchApi.searchTrips({ from, to, date, type });
        return {
            results: response.results || [],
            total: response.total || 0,
            searchParams: { from, to, date, type }
        };
    } catch (e) {
        console.error("Search failed:", e);
        return {
            results: [],
            total: 0,
            searchParams: { from, to, date, type },
            error: "Failed to load trips."
        };
    }
};
