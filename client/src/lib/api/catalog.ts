import { api } from './index';

export interface Station {
    id: string;
    code: string;
    name: string;
    city: string;
    state: string;
    country: string;
    latitude: number;
    longitude: number;
    timezone: string;
    amenities: string[];
}

export interface ListStationsResponse {
    stations: Station[];
    total: number;
}

export const catalogApi = {
    getStations: async (): Promise<Station[]> => {
        const response = await api.get<ListStationsResponse>('/stations');
        return response.stations;
    },

    getStation: async (id: string): Promise<Station> => {
        const response = await api.get<Station>(`/stations/${id}`);
        return response;
    }
};
