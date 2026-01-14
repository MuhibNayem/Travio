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
};
