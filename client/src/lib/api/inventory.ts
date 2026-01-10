import { api } from './index';

export interface Seat {
    seat_id: string;
    seat_number: string;
    column: string;
    seat_type: string;
    seat_class: string;
    status: string; // "AVAILABLE", "BOOKED", "HELD"
    price_paisa: number;
}

export interface SeatRow {
    row_number: number;
    seats: Seat[];
}

export interface SeatMapResponse {
    rows: SeatRow[];
    legend: Record<string, string>; // status -> color mapping
}

export interface CheckAvailabilityResponse {
    is_available: boolean;
    available_seats: number;
    price_paisa: number;
    seats: {
        seat_id: string;
        seat_number: string;
        seat_class: string;
        seat_type: string;
        status: string;
    }[];
}

export interface HoldSeatsRequest {
    trip_id: string;
    from_station_id: string;
    to_station_id: string;
    seat_ids: string[];
    session_id: string;
}

export interface HoldSeatsResponse {
    hold_id: string;
    success: boolean;
    held_seat_ids: string[];
    failed_seat_ids: string[];
    expires_at: string;
    failure_reason: string;
}

export const inventoryApi = {
    getSeatMap: async (tripId: string, fromId: string, toId: string): Promise<SeatMapResponse> => {
        const query = new URLSearchParams({ from: fromId, to: toId });
        const response = await api.get<SeatMapResponse>(`/v1/trips/${tripId}/seatmap?${query.toString()}`);
        return response;
    },

    checkAvailability: async (tripId: string, fromId: string, toId: string, passengers: number): Promise<CheckAvailabilityResponse> => {
        const query = new URLSearchParams({ from: fromId, to: toId, passengers: passengers.toString() });
        const response = await api.get<CheckAvailabilityResponse>(`/v1/trips/${tripId}/availability?${query.toString()}`);
        return response;
    },

    holdSeats: async (payload: HoldSeatsRequest): Promise<HoldSeatsResponse> => {
        const response = await api.post<HoldSeatsResponse>('/v1/bookings/hold', payload);
        return response;
    },

    releaseHold: async (holdId: string): Promise<void> => {
        await api.post<void>(`/v1/bookings/release/${holdId}`, {});
    }
};
