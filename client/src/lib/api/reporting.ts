import { api } from './api';

export interface RevenueData {
    organization_id: string;
    date: string; // ISO timestamp
    order_count: number;
    total_revenue_paisa: number;
    avg_order_value: number;
    currency: string;
}

export interface RevenueReportResponse {
    data: RevenueData[];
    total_count: string; // int64 comes as string sometimes, handle safely
}

export interface OrganizationMetrics {
    organization_id: string;
    total_orders: string; // int64
    total_revenue: string; // int64
    avg_order_value: number;
    total_customers: string; // int64
    repeat_customer_rate: number;
    avg_bookings_per_day: number;
    cancellation_rate: number;
    refund_rate: number;
}

export interface BookingTrendData {
    period: string;
    booking_count: string;
    completed_count: string;
    cancelled_count: string;
    conversion_rate: number;
}

export interface BookingTrendsResponse {
    data: BookingTrendData[];
}

export interface TopRouteData {
    route_name: string;
    booking_count: string;
    revenue: string;
    avg_occupancy: number;
}

export const reportingApi = {
    getRevenueReport: (params: { startDate?: string; endDate?: string; limit?: number }) => {
        const query = new URLSearchParams();
        if (params.startDate) query.append('start_date', params.startDate);
        if (params.endDate) query.append('end_date', params.endDate);
        if (params.limit) query.append('limit', params.limit.toString());

        return api.get<RevenueReportResponse>(`/reports/revenue?${query.toString()}`);
    },

    getOrganizationMetrics: (params: { startDate?: string; endDate?: string }) => {
        const query = new URLSearchParams();
        if (params.startDate) query.append('start_date', params.startDate);
        if (params.endDate) query.append('end_date', params.endDate);

        return api.get<OrganizationMetrics>(`/reports/metrics?${query.toString()}`);
    },

    getBookingTrends: (params: { startDate?: string; endDate?: string; granularity?: 'day' | 'week' }) => {
        const query = new URLSearchParams();
        if (params.startDate) query.append('start_date', params.startDate);
        if (params.endDate) query.append('end_date', params.endDate);
        if (params.granularity) query.append('granularity', params.granularity);

        return api.get<BookingTrendsResponse>(`/reports/bookings?${query.toString()}`);
    },

    getTopRoutes: (params: { limit?: number; sortBy?: 'revenue' | 'bookings' }) => {
        const query = new URLSearchParams();
        if (params.limit) query.append('limit', params.limit.toString());
        if (params.sortBy) query.append('sort_by', params.sortBy);

        return api.get<{ data: TopRouteData[] }>(`/reports/routes?${query.toString()}`);
    }
};
