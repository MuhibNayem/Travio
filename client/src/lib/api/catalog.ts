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
    },

    // Routes
    getRoutes: async (): Promise<Route[]> => {
        const response = await api.get<ListRoutesResponse>('/routes');
        return response.routes;
    },

    createRoute: async (route: CreateRouteRequest): Promise<Route> => {
        const response = await api.post<Route>('/routes', route);
        return response;
    },

    // Trips
    getTrips: async (): Promise<Trip[]> => {
        // Operator view of trips (management)
        // Gateway endpoint: /trips?organization_id=... (handled by backend or context)
        // Check proto: ListTripsRequest
        // We might need a specific management endpoint or use the general list with filters.
        // Assuming /trips is available for operators via Gateway Catalog Handler override or same endpoint?
        // Gateway: r.Post("/trips", catalogHandler.CreateTrip) protected.
        // Gateway: r.Get("/trips/search", ...).
        // Gateway: r.Get("/trips/{tripId}", ...).
        // MISSING: r.Get("/trips") for listing!
        // I need to add that to Gateway first? Or does it exist?
        // Checking Main.go: No "ListTrips" exposed except Search.
        // I will add ListTrips to Gateway first.
        // For now, I'll add the method here and fix Gateway next.
        const response = await api.get<ListTripsResponse>('/trips');
        return response.trips;
    },

    createTrip: async (trip: CreateTripRequest): Promise<Trip> => {
        const response = await api.post<Trip>('/trips', trip);
        return response;
    }
};

export interface Trip {
    id: string;
    organization_id: string;
    route_id: string;
    vehicle_id: string;
    vehicle_type: string;
    vehicle_class: string;
    departure_time: number; // Unix timestamp
    arrival_time: number;
    total_seats: number;
    available_seats: number;
    pricing: TripPricing;
    status: string;
}

export interface TripPricing {
    base_price_paisa: number;
    currency: string;
    class_prices: Record<string, number>;
}

export interface CreateTripRequest {
    route_id: string;
    vehicle_id: string;
    vehicle_type: string;
    vehicle_class: string;
    departure_time: number;
    total_seats: number;
    pricing: TripPricing;
}

export interface ListTripsResponse {
    trips: Trip[];
    total: number;
}

export interface Route {
    id: string;
    organization_id: string;
    code: string;
    name: string;
    origin_station_id: string;
    destination_station_id: string;
    intermediate_stops: RouteStop[];
    distance_km: number;
    estimated_duration_minutes: number;
    status: string;
}

export interface RouteStop {
    station_id: string;
    sequence: number;
    arrival_offset_minutes: number;
    departure_offset_minutes: number;
    distance_from_origin_km: number;
}

export interface ListRoutesResponse {
    routes: Route[];
    total: number;
}

export interface CreateRouteRequest {
    code: string;
    name: string;
    origin_station_id: string;
    destination_station_id: string;
    distance_km: number;
    estimated_duration_minutes: number;
    intermediate_stops?: RouteStop[];
}
