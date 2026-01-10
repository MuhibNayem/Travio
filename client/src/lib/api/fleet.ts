import { api } from './index';

export enum AssetType {
    BUS = "ASSET_TYPE_BUS",
    TRAIN = "ASSET_TYPE_TRAIN",
    LAUNCH = "ASSET_TYPE_LAUNCH",
}

export enum AssetStatus {
    ACTIVE = "ASSET_STATUS_ACTIVE",
    MAINTENANCE = "ASSET_STATUS_MAINTENANCE",
    RETIRED = "ASSET_STATUS_RETIRED",
}

export interface AssetConfig {
    layout_type?: string;
    features?: string;
}

export interface Asset {
    id: string;
    organization_id: string;
    type: string; // Enum string from backend
    name: string;
    license_plate: string;
    vin: string;
    make: string;
    model: string;
    year: number;
    status: string;
    config: AssetConfig;
    created_at: string;
    updated_at: string;
}

export interface ListAssetsResponse {
    assets: Asset[];
    total_count: number;
}

export interface RegisterAssetRequest {
    name: string;
    type: string;
    license_plate: string;
    vin: string;
    make: string;
    model: string;
    year: number;
    status: string;
    config: AssetConfig;
}

export const fleetApi = {
    getAssets: async (): Promise<Asset[]> => {
        // Gateway endpoint: /v1/fleet/assets
        // Gateway handler name: likely ListAssets if generated? 
        // Wait, Gateway needs to expose this.
        // Checking Gateway routes...
        const response = await api.get<ListAssetsResponse>('/fleet/assets');
        return response.assets || [];
    },

    registerAsset: async (req: RegisterAssetRequest): Promise<Asset> => {
        const response = await api.post<Asset>('/fleet/assets', req);
        return response;
    },

    getAsset: async (id: string): Promise<Asset> => {
        const response = await api.get<Asset>(`/fleet/assets/${id}`);
        return response;
    }
};
