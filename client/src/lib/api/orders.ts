import { api } from "./index";

// --- Types ---

export interface Order {
    id: string;
    organization_id: string;
    user_id: string;
    trip_id: string;
    from_station_id: string;
    to_station_id: string;
    passengers: Passenger[];
    subtotal_paisa: number;
    tax_paisa: number;
    booking_fee_paisa: number;
    discount_paisa: number;
    total_paisa: number;
    currency: string;
    payment_id: string;
    payment_status: PaymentStatus;
    booking_id: string;
    seats: BookedSeat[];
    status: OrderStatus;
    created_at: string;
    updated_at: string;
    contact_email: string;
    contact_phone: string;
}

export interface Passenger {
    nid: string;
    name: string;
    seat_id: string;
    seat_number: string;
    seat_class: string;
    gender: string;
    age: number;
    nid_verified: boolean;
}

export interface BookedSeat {
    seat_id: string;
    seat_number: string;
    seat_class: string;
    ticket_id: string;
    price_paisa: number;
}

export enum OrderStatus {
    ORDER_STATUS_UNSPECIFIED = 0,
    ORDER_STATUS_PENDING = 1,
    ORDER_STATUS_CONFIRMED = 2,
    ORDER_STATUS_FAILED = 3,
    ORDER_STATUS_CANCELLED = 4,
    ORDER_STATUS_EXPIRED = 5,
    ORDER_STATUS_REFUND_PENDING = 6,
    ORDER_STATUS_REFUNDED = 7,
}

export enum PaymentStatus {
    PAYMENT_STATUS_UNSPECIFIED = 0,
    PAYMENT_STATUS_PENDING = 1,
    PAYMENT_STATUS_AUTHORIZED = 2,
    PAYMENT_STATUS_CAPTURED = 3,
    PAYMENT_STATUS_FAILED = 4,
    PAYMENT_STATUS_REFUNDED = 5,
}

export interface CreateOrderRequest {
    organization_id: string;
    user_id: string;
    trip_id: string;
    from_station_id: string;
    to_station_id: string;
    hold_id?: string;
    passengers: PassengerRequest[];
    payment_method: PaymentMethodRequest;
    contact_email: string;
    contact_phone: string;
    coupon_code?: string;
    idempotency_key?: string;
}

export interface PassengerRequest {
    nid?: string;
    name: string;
    seat_id: string;
    gender?: string;
    age?: number;
    date_of_birth?: string;
}

export interface PaymentMethodRequest {
    type: string; // card, bkash, nagad, bank, cash
    token?: string;
    card_last_four?: string;
    card_brand?: string;
}

export interface CreateOrderResponse {
    order: Order;
    payment_redirect_url: string;
    requires_action: boolean;
}

export interface ListOrdersResponse {
    orders: Order[];
    next_page_token: string;
    total_count: number;
}

// --- API Methods ---

export const ordersApi = {
    createOrder: (data: CreateOrderRequest) => api.post<CreateOrderResponse>("/v1/orders", data),

    getOrder: (orderId: string) => api.get<Order>(`/v1/orders/${orderId}`),

    listOrders: (userId: string) => api.get<ListOrdersResponse>(`/v1/orders?user_id=${userId}`),

    cancelOrder: (orderId: string, reason: string) => api.post<{ success: boolean }>(`/v1/orders/${orderId}/cancel`, { reason }),
};
