import { api } from "./index";

// --- Types ---

export interface PaymentMethod {
    id: string; // card, bkash, nagad
    name: string;
    enabled: boolean;
}

export interface GetPaymentMethodsResponse {
    methods: PaymentMethod[];
}

export interface UpdatePaymentConfigRequest {
    organization_id: string;
    gateway: string;
    credentials: Record<string, string>;
    is_active: boolean;
}

export interface UpdatePaymentConfigResponse {
    success: boolean;
    message: string;
}

export interface GetPaymentConfigResponse {
    organization_id: string;
    gateway: string;
    is_active: boolean;
}

// --- API Methods ---

export const paymentApi = {
    getPaymentMethods: () => api.get<GetPaymentMethodsResponse>("/v1/payments/methods"),

    // Process generic payment (standalone, usually CreateOrder handles this for booking)
    processPayment: (data: any) => api.post<any>("/v1/payments", data),

    getPaymentStatus: (orderId: string) => api.get<any>(`/v1/payments/${orderId}`),

    // Org Config
    updatePaymentConfig: (orgId: string, data: UpdatePaymentConfigRequest) => api.put<UpdatePaymentConfigResponse>(`/v1/organizations/${orgId}/payment-config`, data),

    getPaymentConfig: (orgId: string) => api.get<GetPaymentConfigResponse>(`/v1/organizations/${orgId}/payment-config`),
};
