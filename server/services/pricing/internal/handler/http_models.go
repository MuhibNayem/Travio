package handler

type CalculatePriceRequest struct {
	TripID         string  `json:"trip_id"`
	SeatClass      string  `json:"seat_class"`
	SeatCategory   string  `json:"seat_category"`
	Date           string  `json:"date"`
	Quantity       int32   `json:"quantity"`
	BasePricePaisa int64   `json:"base_price_paisa"`
	OccupancyRate  float64 `json:"occupancy_rate"`
	OrganizationID string  `json:"organization_id"`
	DepartureTime  int64   `json:"departure_time"`
	RouteID        string  `json:"route_id"`
	ScheduleID     string  `json:"schedule_id"`
	FromStationID  string  `json:"from_station_id"`
	ToStationID    string  `json:"to_station_id"`
	VehicleType    string  `json:"vehicle_type"`
	VehicleClass   string  `json:"vehicle_class"`
	PromoCode      string  `json:"promo_code"`
}

type CalculatePriceResponse struct {
	FinalPricePaisa int64         `json:"final_price_paisa"`
	BasePricePaisa  int64         `json:"base_price_paisa"`
	AppliedRules    []AppliedRule `json:"applied_rules"`
}

type AppliedRule struct {
	RuleID     string  `json:"rule_id"`
	RuleName   string  `json:"rule_name"`
	Multiplier float64 `json:"multiplier"`
}
