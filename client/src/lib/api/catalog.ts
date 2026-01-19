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
    next_page_token?: string;
}

export const catalogApi = {
    getStations: async (params: { search_query?: string; page_size?: number; page_token?: string } = {}): Promise<ListStationsResponse> => {
        const query = new URLSearchParams();
        if (params.search_query) query.append('search_query', params.search_query);
        if (params.page_size) query.append('page_size', params.page_size.toString());
        if (params.page_token) query.append('page_token', params.page_token);

        const response = await api.get<ListStationsResponse>(`/v1/stations?${query.toString()}`);
        return response ?? { stations: [], total: 0 };
    },

    getStation: async (id: string): Promise<Station> => {
        const response = await api.get<Station>(`/v1/stations/${id}`);
        return response;
    },

    // Routes
    getRoutes: async (): Promise<Route[]> => {
        const response = await api.get<ListRoutesResponse>('/v1/routes');
        return response?.routes ?? [];
    },

    createRoute: async (route: CreateRouteRequest): Promise<Route> => {
        const response = await api.post<Route>('/v1/routes', route);
        return response;
    },

    // Trip Instances
    listTripInstances: async (params: ListTripInstancesParams = {}): Promise<TripInstanceResult[]> => {
        const query = new URLSearchParams();
        if (params.schedule_id) query.append('schedule_id', params.schedule_id);
        if (params.route_id) query.append('route_id', params.route_id);
        if (params.start_date) query.append('start_date', params.start_date);
        if (params.end_date) query.append('end_date', params.end_date);
        if (params.status) query.append('status', params.status);

        const response = await api.get<ListTripInstancesResponse>(`/v1/trip-instances?${query.toString()}`);
        return response?.results ?? [];
    },

    getTripInstance: async (id: string): Promise<TripInstanceResult> => {
        const response = await api.get<TripInstanceResult>(`/v1/trip-instances/${id}`);
        return response;
    },

    // Schedules
    createSchedule: async (schedule: CreateScheduleRequest): Promise<Schedule> => {
        const response = await api.post<Schedule>('/v1/schedules', schedule);
        return response;
    },

    bulkCreateSchedules: async (schedules: CreateScheduleRequest[]): Promise<Schedule[]> => {
        const response = await api.post<BulkCreateSchedulesResponse>('/v1/schedules/bulk', { schedules });
        return response?.schedules ?? [];
    },

    listSchedules: async (params: ListSchedulesParams = {}): Promise<Schedule[]> => {
        const query = new URLSearchParams();
        if (params.route_id) query.append('route_id', params.route_id);
        if (params.status) query.append('status', params.status);
        const response = await api.get<ListSchedulesResponse>(`/v1/schedules?${query.toString()}`);
        return response?.schedules ?? [];
    },

    generateTripInstances: async (scheduleId: string, startDate: string, endDate: string): Promise<GenerateTripInstancesResponse> => {
        const query = new URLSearchParams();
        if (startDate) query.append('start_date', startDate);
        if (endDate) query.append('end_date', endDate);
        const response = await api.post<GenerateTripInstancesResponse>(`/v1/schedules/${scheduleId}/generate?${query.toString()}`, {});
        return response;
    },
};

export interface Trip {
    id: string;
    organization_id: string;
    schedule_id?: string;
    service_date?: string;
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
    tax_paisa: number;
    booking_fee_paisa: number;
    currency: string;
    class_prices: Record<string, number>;
    seat_category_prices: Record<string, number>;
    segment_prices: SegmentPricing[];
}

export interface SegmentPricing {
    from_station_id: string;
    to_station_id: string;
    base_price_paisa: number;
    class_prices: Record<string, number>;
    seat_category_prices: Record<string, number>;
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

export interface TripInstanceResult {
    trip: Trip;
    route: Route;
    origin_station: Station;
    destination_station: Station;
    operator_name?: string;
}

export interface ListTripInstancesResponse {
    results: TripInstanceResult[];
    next_page_token: string;
    total_count: number;
}

export interface ListTripInstancesParams {
    schedule_id?: string;
    route_id?: string;
    start_date?: string;
    end_date?: string;
    status?: string;
}

export interface Schedule {
    id: string;
    organization_id: string;
    route_id: string;
    vehicle_id: string;
    vehicle_type: string;
    vehicle_class: string;
    total_seats: number;
    pricing: TripPricing;
    departure_minutes: number;
    arrival_offset_minutes: number;
    timezone: string;
    start_date: string;
    end_date: string;
    days_of_week: number;
    status: string;
    created_at: number;
    updated_at: number;
    version: number;
}

export interface CreateScheduleRequest {
    route_id: string;
    vehicle_id: string;
    vehicle_type: string;
    vehicle_class: string;
    total_seats: number;
    pricing: TripPricing;
    departure_minutes: number;
    arrival_offset_minutes: number;
    timezone: string;
    start_date: string;
    end_date: string;
    days_of_week: number;
    status?: string;
}

export interface ListSchedulesParams {
    route_id?: string;
    status?: string;
}

export interface ListSchedulesResponse {
    schedules: Schedule[];
    next_page_token: string;
    total_count: number;
}

export interface GenerateTripInstancesResponse {
    trips: Trip[];
    created_count: number;
}

export interface BulkCreateSchedulesResponse {
    schedules: Schedule[];
    created_count: number;
}
