import { api } from './api';

export interface PlanFeature {
    [key: string]: string;
}

export interface Plan {
    id: string;
    name: string;
    description: string;
    price: number;
    billing_cycle: string;
    features: PlanFeature;
    max_users: number;
    is_active: boolean;
}

export interface ListPlansResponse {
    plans: Plan[];
}

export interface CreateSubscriptionResponse {
    subscription_id: string;
    status: string;
}

export const subscriptionApi = {
    listPlans: async (): Promise<Plan[]> => {
        const response = await api.get<ListPlansResponse>('/plans');
        return response.plans;
    },

    createSubscription: async (organizationId: string, planId: string): Promise<CreateSubscriptionResponse> => {
        return api.post<CreateSubscriptionResponse>('/subscriptions', {
            organization_id: organizationId,
            plan_id: planId
        });
    },

    getSubscription: async (organizationId: string): Promise<Subscription> => {
        return api.get<Subscription>(`/subscriptions/${organizationId}`);
    },

    cancelSubscription: async (organizationId: string): Promise<any> => {
        return api.post(`/subscriptions/${organizationId}/cancel`, {});
    }
};

export interface Subscription {
    id: string;
    organization_id: string;
    plan_id: string;
    status: string; // active, canceled, etc.
    start_date: string;
    end_date: string;
}
