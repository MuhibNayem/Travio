import { api } from "./index";

// --- Types ---

export interface Venue {
    id: string;
    organization_id: string;
    name: string;
    address: string;
    city: string;
    country: string;
    capacity: number;
    type: VenueType;
    sections: SeatingSection[];
    map_image_url: string;
    created_at: string;
    updated_at: string;
}

export enum VenueType {
    VENUE_TYPE_UNSPECIFIED = 0,
    VENUE_TYPE_STADIUM = 1,
    VENUE_TYPE_AUDITORIUM = 2,
    VENUE_TYPE_CONFERENCE_HALL = 3,
    VENUE_TYPE_OUTDOOR_GROUND = 4,
}

export interface SeatingSection {
    id: string;
    name: string;
    capacity: number;
    rows: number;
    seats_per_row: number;
    type: string;
}

export interface CreateVenueRequest {
    organization_id: string;
    name: string;
    address: string;
    city: string;
    country: string;
    type: string; // Enum string in API
    sections: SeatingSection[];
}

export interface UpdateVenueRequest {
    id: string;
    organization_id: string;
    name: string;
    type: string;
}

export interface Event {
    id: string;
    organization_id: string;
    venue_id: string;
    title: string;
    description: string;
    category: string;
    images: string[];
    start_time: string;
    end_time: string;
    status: EventStatus;
    venue?: Venue; // Enriched in search/get
    created_at: string;
    updated_at: string;
}

export enum EventStatus {
    EVENT_STATUS_UNSPECIFIED = 0,
    EVENT_STATUS_DRAFT = 1,
    EVENT_STATUS_PUBLISHED = 2,
    EVENT_STATUS_CANCELLED = 3,
    EVENT_STATUS_COMPLETED = 4,
}

export interface CreateEventRequest {
    organization_id: string;
    venue_id: string;
    title: string;
    description: string;
    category: string;
    start_time: string;
    end_time: string;
}

export interface UpdateEventRequest {
    id: string;
    organization_id: string;
    title: string;
    description: string;
    start_time: string;
    end_time: string;
}

export interface TicketType {
    id: string;
    event_id: string;
    name: string;
    description: string;
    price_paisa: number;
    total_quantity: number;
    available_quantity: number;
    max_per_user: number;
    sales_start_time: string;
    sales_end_time: string;
}

export interface CreateTicketTypeRequest {
    event_id: string;
    organization_id: string;
    name: string;
    price_paisa: number;
    total_quantity: number;
    sales_start_time: string;
    sales_end_time: string;
}

export interface ListEventsResponse {
    events: Event[];
    next_page_token: string;
    total_count: number;
}

export interface ListVenuesResponse {
    venues: Venue[];
    next_page_token: string;
}

export interface SearchEventsResponse {
    results: EventSearchResult[];
    total_count: number;
}

export interface EventSearchResult {
    event: Event;
    venue: Venue;
}


// --- API Methods ---
export const eventsApi = {
    // --- Venues ---
    getVenues: (orgId: string) => api.get<ListVenuesResponse>(`/v1/venues?organization_id=${orgId}`).then(r => r.venues || []),
    createVenue: (data: CreateVenueRequest) => api.post<Venue>("/v1/venues", data),
    updateVenue: (id: string, data: UpdateVenueRequest) => api.put<Venue>(`/v1/events/venues/${id}`, data),
    getVenue: (id: string) => api.get<Venue>(`/v1/venues/${id}`),

    // --- Events ---
    getEvents: (orgId: string) => api.get<ListEventsResponse>(`/v1/events?organization_id=${orgId}`).then(r => r.events || []),
    createEvent: (data: CreateEventRequest) => api.post<Event>("/v1/events", data),
    updateEvent: (id: string, data: UpdateEventRequest) => api.put<Event>(`/v1/events/${id}`, data),
    getEvent: (id: string) => api.get<Event>(`/v1/events/${id}`),
    publishEvent: (id: string) => api.post<Event>(`/v1/events/${id}/publish`, {}),

    // --- Search ---
    searchEvents: (query: string, city?: string, category?: string) => {
        const params = new URLSearchParams({ q: query });
        if (city) params.append("city", city);
        if (category) params.append("category", category);
        return api.get<SearchEventsResponse>(`/v1/search/events?${params.toString()}`);
    },

    // --- Tickets ---
    getTicketTypes: (eventId: string) => api.get<any>(`/v1/events/${eventId}/tickets`).then(r => r.ticket_types || []),
    createTicketType: (data: CreateTicketTypeRequest) => api.post<TicketType>("/v1/tickets/types", data),
};
