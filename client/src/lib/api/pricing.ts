import { api } from './index';

export interface PricingRule {
    id: string;
    organization_id?: string;
    name: string;
    description?: string;
    condition: string;
    multiplier: number;
    adjustment_type: string;
    adjustment_value: number;
    priority: number;
    is_active: boolean;
    created_at?: string;
    updated_at?: string;
}

export interface ListPricingRulesResponse {
    rules: PricingRule[];
}

export interface CreatePricingRuleRequest {
    name: string;
    description?: string;
    condition: string;
    multiplier: number;
    adjustment_type: string;
    adjustment_value: number;
    priority: number;
}

export interface UpdatePricingRuleRequest {
    name: string;
    description?: string;
    condition: string;
    multiplier: number;
    adjustment_type: string;
    adjustment_value: number;
    priority: number;
    is_active: boolean;
}

export const pricingApi = {
    listRules: async (includeInactive = false): Promise<PricingRule[]> => {
        const response = await api.get<ListPricingRulesResponse>(
            `/v1/pricing/rules?include_inactive=${includeInactive}`,
        );
        return response?.rules ?? [];
    },

    createRule: async (payload: CreatePricingRuleRequest): Promise<PricingRule> => {
        const response = await api.post<{ rule: PricingRule }>('/v1/pricing/rules', payload);
        return response?.rule ?? response;
    },

    updateRule: async (id: string, payload: UpdatePricingRuleRequest): Promise<PricingRule> => {
        const response = await api.put<{ rule: PricingRule }>(`/v1/pricing/rules/${id}`, payload);
        return response?.rule ?? response;
    },

    deleteRule: async (id: string): Promise<void> => {
        await api.delete(`/v1/pricing/rules/${id}`);
    },

    // Promotions
    getPromotions: async (activeOnly = true): Promise<Promotion[]> => {
        const response = await api.get<{ promotions: Promotion[] }>(
            `/v1/pricing/promotions?active_only=${activeOnly}`,
        );
        return response?.promotions ?? [];
    },

    createPromotion: async (payload: CreatePromotionRequest): Promise<Promotion> => {
        const response = await api.post<{ promotion: Promotion }>('/v1/pricing/promotions', payload);
        return response?.promotion ?? response;
    },
};

export interface Promotion {
    id: string;
    code: string;
    description?: string;
    discount_type: string; // "percentage", "fixed"
    discount_value: number;
    max_usage: number;
    current_usage: number;
    valid_from?: string;
    valid_until?: string;
    min_order_amount_paisa: number;
    is_active: boolean;
}

export interface CreatePromotionRequest {
    code: string;
    description?: string;
    discount_type: string;
    discount_value: number;
    max_usage: number;
    valid_from?: string;
    valid_until?: string;
    min_order_amount_paisa: number;
    organization_id?: string;
}
