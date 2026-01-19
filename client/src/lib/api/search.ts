import { api } from './index';
import type { TripInstanceResult } from './catalog';

export interface TripSearchResult {
    trip_id: string;
    vehicle_type: string;
    vehicle_class: string;
    departure_time: number;    // Unix timestamp
    arrival_time: number;      // Unix timestamp
    price_paisa: number;
    total_seats: number;
    available_seats: number;
    from_station_id: string;
    from_station_name: string;
    from_city: string;
    to_station_id: string;
    to_station_name: string;
    to_city: string;
    date: string;              // YYYY-MM-DD
    status: string;
    route_id: string;
    organization_id: string;
    operator_name?: string;
    route_name?: string;
}

export interface SearchResponse {
    results: TripSearchResult[];
    total: number;
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

        const response = await api.get<SearchResponse>(`/v1/search/trips?${query.toString()}`);
        return response;
    },

    getTripInstance: async (id: string): Promise<TripInstanceResult> => {
        const response = await api.get<TripInstanceResult>(`/v1/trip-instances/${id}`);
        return response;
    },
};
