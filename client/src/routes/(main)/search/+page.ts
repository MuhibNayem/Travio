import type { PageLoad } from './$types';
import { searchApi } from '$lib/api/search';
import { eventsApi } from '$lib/api/events';

export const load: PageLoad = async ({ url }) => {
    const from = url.searchParams.get('from') || '';
    const to = url.searchParams.get('to') || '';
    const date = url.searchParams.get('date') || new Date().toISOString().split('T')[0];
    const type = url.searchParams.get('type') || 'bus';
    const query = url.searchParams.get('q') || '';
    const city = url.searchParams.get('city') || '';

    try {
        if (type === 'events') {
            // For events, 'from' might be treated as city if passed, or use explicit 'city' param
            // SearchHero might pass city name in 'from' if we hack it, but let's prefer 'city' param.
            const searchCity = city || (from && !from.includes('-') ? from : '');

            const response = await eventsApi.searchEvents(query, searchCity);
            return {
                mode: 'events',
                results: response.results || [],
                total: response.total_count || 0,
                searchParams: { from, to, date, type, query, city }
            };
        } else {
            const response = await searchApi.searchTrips({ from, to, date, type });
            return {
                mode: 'trips',
                results: response.results || [],
                total: response.total || 0,
                searchParams: { from, to, date, type }
            };
        }
    } catch (e) {
        console.error("Search failed:", e);
        return {
            mode: type === 'events' ? 'events' : 'trips',
            results: [],
            total: 0,
            searchParams: { from, to, date, type },
            error: "Failed to load results."
        };
    }
};
