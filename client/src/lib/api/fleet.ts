import { api } from './index';

// ========== ENUMS ==========

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

export enum TrainCoachClass {
    UNSPECIFIED = "TRAIN_COACH_CLASS_UNSPECIFIED",
    AC_FIRST = "TRAIN_COACH_CLASS_AC_FIRST",
    AC_SECOND = "TRAIN_COACH_CLASS_AC_SECOND",
    AC_CHAIR = "TRAIN_COACH_CLASS_AC_CHAIR",
    SLEEPER = "TRAIN_COACH_CLASS_SLEEPER",
    S_CHAIR = "TRAIN_COACH_CLASS_S_CHAIR",
    GENERAL = "TRAIN_COACH_CLASS_GENERAL",
}

export enum BerthType {
    UNSPECIFIED = "BERTH_TYPE_UNSPECIFIED",
    TWO_TIER = "BERTH_TYPE_TWO_TIER",
    THREE_TIER = "BERTH_TYPE_THREE_TIER",
    CHAIR = "BERTH_TYPE_CHAIR",
}

export enum LaunchDeckType {
    UNSPECIFIED = "LAUNCH_DECK_TYPE_UNSPECIFIED",
    ECONOMY = "LAUNCH_DECK_TYPE_ECONOMY",
    BUSINESS = "LAUNCH_DECK_TYPE_BUSINESS",
    VIP_CABIN = "LAUNCH_DECK_TYPE_VIP_CABIN",
}

// ========== BUS CONFIG ==========

export interface SeatCategory {
    name: string;           // "Economy", "Business", "VIP"
    price_paisa: number;
    seat_ids: string[];     // Which seats belong to this category
}

export interface BusConfig {
    rows: number;
    seats_per_row: number;      // e.g., 4 for 2+2 layout
    aisle_after_seat: number;   // Position of aisle (2 means after seat 2)
    has_toilet: boolean;
    has_sleeper: boolean;
    categories?: SeatCategory[];
}

// ========== TRAIN CONFIG ==========

export interface BerthConfiguration {
    type: BerthType;
    berths_per_compartment: number;
    has_side_berths: boolean;
}

export interface TrainCoach {
    id: string;             // e.g., "S1", "AC1"
    name: string;           // "Shatabdi Chair Car"
    class: TrainCoachClass;
    rows: number;
    seats_per_row: number;  // 4 for 2+2, 6 for 3+3
    has_berths: boolean;
    berth_config?: BerthConfiguration;
    price_paisa: number;
}

export interface TrainConfig {
    coaches: TrainCoach[];
}

// ========== LAUNCH CONFIG ==========

export interface LaunchCabin {
    id: string;
    name: string;
    beds: number;
    price_paisa: number;
    is_suite: boolean;
}

export interface LaunchDeck {
    id: string;
    name: string;
    type: LaunchDeckType;
    rows?: number;          // For open seating
    cols?: number;
    seat_price_paisa?: number;
    cabins?: LaunchCabin[]; // For VIP cabin type
}

export interface LaunchConfig {
    decks: LaunchDeck[];
}

// ========== UNIFIED CONFIG ==========

export interface AssetConfig {
    bus?: BusConfig;
    train?: TrainConfig;
    launch?: LaunchConfig;
    features?: string[];    // ["AC", "WiFi", "Recliner"]
}

// ========== ASSET ==========

export interface Asset {
    id: string;
    organization_id: string;
    type: AssetType | string;
    name: string;
    license_plate: string;
    vin: string;
    make: string;
    model: string;
    year: number;
    status: AssetStatus | string;
    config: AssetConfig;
    created_at: string;
    updated_at: string;
}

// ========== API TYPES ==========

export interface ListAssetsResponse {
    assets: Asset[];
    total_count: number;
}

export interface RegisterAssetRequest {
    name: string;
    type: AssetType | string;
    license_plate: string;
    vin?: string;
    make?: string;
    model?: string;
    year?: number;
    status?: AssetStatus | string;
    config: AssetConfig;
}

export interface UpdateAssetRequest {
    id: string;
    name?: string;
    status?: AssetStatus | string;
    config?: AssetConfig;
}

// ========== API METHODS ==========

const normalizeBusConfig = (raw: Partial<BusConfig> | undefined): BusConfig | undefined => {
    if (!raw) return undefined;
    return {
        rows: raw.rows ?? 0,
        seats_per_row:
            (raw as { seats_per_row?: number }).seats_per_row ??
            (raw as { seatsPerRow?: number }).seatsPerRow ??
            0,
        aisle_after_seat:
            (raw as { aisle_after_seat?: number }).aisle_after_seat ??
            (raw as { aisleAfterSeat?: number }).aisleAfterSeat ??
            0,
        has_toilet:
            (raw as { has_toilet?: boolean }).has_toilet ??
            (raw as { hasToilet?: boolean }).hasToilet ??
            false,
        has_sleeper:
            (raw as { has_sleeper?: boolean }).has_sleeper ??
            (raw as { hasSleeper?: boolean }).hasSleeper ??
            false,
        categories: raw.categories,
    };
};

const normalizeAssetConfig = (raw: AssetConfig | undefined): AssetConfig => {
    if (!raw) return {};
    return {
        bus: normalizeBusConfig(raw.bus),
        train: raw.train,
        launch: raw.launch,
        features: raw.features,
    };
};

const normalizeAsset = (raw: Partial<Asset> & Record<string, unknown>): Asset => {
    return {
        id: (raw.id as string) || '',
        organization_id:
            (raw.organization_id as string) ||
            (raw.organizationId as string) ||
            '',
        type: (raw.type as AssetType | string) || AssetType.BUS,
        name: (raw.name as string) || '',
        license_plate:
            (raw.license_plate as string) ||
            (raw.licensePlate as string) ||
            '',
        vin: (raw.vin as string) || '',
        make: (raw.make as string) || '',
        model: (raw.model as string) || '',
        year: (raw.year as number) || 0,
        status: (raw.status as AssetStatus | string) || '',
        config: normalizeAssetConfig(raw.config as AssetConfig | undefined),
        created_at:
            (raw.created_at as string) ||
            (raw.createdAt as string) ||
            '',
        updated_at:
            (raw.updated_at as string) ||
            (raw.updatedAt as string) ||
            '',
    };
};

export const fleetApi = {
    getAssets: async (): Promise<Asset[]> => {
        const response = await api.get<ListAssetsResponse>('/v1/fleet/assets');
        const assets = response?.assets ?? [];
        return assets.map((asset) => normalizeAsset(asset));
    },

    registerAsset: async (req: RegisterAssetRequest): Promise<Asset> => {
        const response = await api.post<Asset>('/v1/fleet/assets', req);
        return normalizeAsset(response as unknown as Asset);
    },

    getAsset: async (id: string): Promise<Asset> => {
        const response = await api.get<Asset>(`/v1/fleet/assets/${id}`);
        return normalizeAsset(response as unknown as Asset);
    },

    updateAsset: async (req: UpdateAssetRequest): Promise<Asset> => {
        const response = await api.put<Asset>(`/v1/fleet/assets/${req.id}`, req);
        return normalizeAsset(response as unknown as Asset);
    },
};
