import { api } from './index';

export interface TripSearchResult {
    id: string;
    route_id: string;
    type: string;
    operator: string;
    vehicle_name: string;
    departure_time: string;
    arrival_time: string;
    price: number;
    class: string;
    available_seats: number;
    total_seats: number;
    from: string;
    from_city: string;
    to: string;
    to_city: string;
    duration: number;
    distance: number;
}

export interface SearchResponse {
    results: TripSearchResult[];
    total: number;
    next_page: string;
}

export interface SearchParams {
    from: string;
    to: string;
    date: string; // YYYY-MM-DD
    type?: string;
}

export const searchApi = {
    searchTrips: async (params: SearchParams): Promise<SearchResponse> => {
        const query = new URLSearchParams({
            from: params.from,
            to: params.to,
            date: params.date,
        });
        if (params.type) {
            query.append('type', params.type);
        }

        const response = await api.get<SearchResponse>(`/search/trips?${query.toString()}`);
        return response;
    },

    getTrip: async (id: string): Promise<TripSearchResult> => {
        const response = await api.get<TripSearchResult>(`/trips/${id}`);
        return response;
    }
};
