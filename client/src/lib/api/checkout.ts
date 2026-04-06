/**
 * Checkout API client
 */

import { api } from './index';

export interface PassengerInfo {
    nid: string;
    name: string;
    seat_id: string;
    seat_number: string;
    date_of_birth: string;
    gender: string;
    phone: string;
    email: string;
    price?: number;
}

export interface CreateCheckoutRequest {
    trip_id: string;
    hold_id: string;
    from_station_id: string;
    to_station_id: string;
    passengers: PassengerInfo[];
    coupon_code?: string;
}

export interface CheckoutSession {
    id: string;
    trip_id: string;
    hold_id: string;
    passengers: PassengerInfo[];
    base_price_paisa: number;
    tax_paisa: number;
    booking_fee_paisa: number;
    discount_paisa: number;
    total_paisa: number;
    currency: string;
    coupon_code?: string;
    coupon_discount: number;
    coupon_validated: boolean;
    payment_method: string;
    status: 'pending' | 'in_progress' | 'confirmed' | 'failed' | 'expired';
    order_id?: string;
    created_at: string;
    expires_at: string;
}

export interface HoldStatus {
    hold_id: string;
    trip_id: string;
    from_station_id: string;
    to_station_id: string;
    held_seat_ids: string[];
    expires_at: string;
    total_paisa: number;
    status: string;
}

export const checkoutApi = {
    createCheckout: async (payload: CreateCheckoutRequest): Promise<CheckoutSession> => {
        const response = await api.post<CheckoutSession>('/v1/checkout', payload);
        return response;
    },

    getCheckout: async (checkoutId: string): Promise<CheckoutSession> => {
        const response = await api.get<CheckoutSession>(`/v1/checkout/${checkoutId}`);
        return response;
    },

    getCheckoutByHoldId: async (holdId: string): Promise<CheckoutSession> => {
        const response = await api.get<CheckoutSession>(`/v1/checkout/hold/${holdId}`);
        return response;
    },

    updateCheckout: async (checkoutId: string, payload: Partial<CreateCheckoutRequest>): Promise<CheckoutSession> => {
        const response = await api.patch<CheckoutSession>(`/v1/checkout/${checkoutId}`, payload);
        return response;
    },

    confirmCheckout: async (checkoutId: string): Promise<CheckoutSession> => {
        const response = await api.post<CheckoutSession>(`/v1/checkout/${checkoutId}/confirm`, {});
        return response;
    },

    listCheckouts: async (params?: { status?: string; limit?: number; offset?: number }): Promise<{ sessions: CheckoutSession[]; total: number }> => {
        const query = new URLSearchParams();
        if (params?.status) query.append('status', params.status);
        if (params?.limit) query.append('limit', params.limit.toString());
        if (params?.offset) query.append('offset', params.offset.toString());
        
        const response = await api.get<{ sessions: CheckoutSession[]; total: number }>(`/v1/checkout?${query.toString()}`);
        return response;
    },
};
