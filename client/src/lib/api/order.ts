import { api } from './index';

export interface Passenger {
    nid: string;
    name: string;
    seat_id: string; // The seat ID they are occupying
    date_of_birth: string; // YYYY-MM-DD
    gender: string;
    age: number;
}

export interface PaymentMethod {
    type: string;
    token?: string;
}

export interface CreateOrderRequest {
    trip_id: string;
    from_station_id: string;
    to_station_id: string;
    hold_id: string;
    passengers: Passenger[];
    payment_method: PaymentMethod;
    contact_email: string;
    contact_phone: string;
    idempotency_key: string;
}

export interface Order {
    id: string;
    trip_id: string;
    route_id: string;
    from_station_id: string;
    to_station_id: string;
    status: string; // "PENDING", "CONFIRMED", "CANCELLED"
    passengers: any[]; // mapped passenger structure
    subtotal_paisa: number;
    tax_paisa: number;
    booking_fee_paisa: number;
    discount_paisa: number;
    total_paisa: number;
    currency: string;
    payment_status: string;
    contact_email: string;
    contact_phone: string;
    created_at: string;
    expires_at: string;
}

export const orderApi = {
    createOrder: async (payload: CreateOrderRequest): Promise<Order> => {
        const response = await api.post<Order>('/v1/orders', payload);
        return response;
    },

    getOrder: async (orderId: string): Promise<Order> => {
        const response = await api.get<Order>(`/v1/orders/${orderId}`);
        return response;
    },

    listOrders: async (pageToken?: string): Promise<{ orders: Order[], next_page: string, total: number }> => {
        const query = pageToken ? `?page_token=${pageToken}` : '';
        const response = await api.get<{ orders: Order[], next_page: string, total: number }>(`/v1/orders${query}`);
        return response;
    },

    cancelOrder: async (orderId: string, reason: string): Promise<{ success: boolean, order: Order, refund?: any }> => {
        const response = await api.post<{ success: boolean, order: Order, refund?: any }>(`/v1/orders/${orderId}/cancel`, { reason });
        return response;
    }
};
