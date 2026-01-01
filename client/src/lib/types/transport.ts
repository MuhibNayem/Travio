export interface Station {
    id: string;
    name: string;
    code: string;
    city: string;
}

export interface Route {
    id: string;
    name: string;
    originId: string;
    destinationId: string;
    distance: number; // km
}

export type TransportType = "bus" | "train" | "launch";

export interface Trip {
    id: string;
    routeId: string;
    type: TransportType;
    operator: string;
    vehicleName: string; // e.g., "Green Line 101"
    departureTime: string; // ISO String
    arrivalTime: string; // ISO String
    price: number;
    class: string; // e.g., "AC Business", "Shovan"
    availableSeats: number;
    totalSeats: number;
}

export type SeatStatus = "available" | "booked" | "held" | "selected";

export interface Seat {
    id: string;
    label: string; // e.g., "A1"
    status: SeatStatus;
    price: number;
}
