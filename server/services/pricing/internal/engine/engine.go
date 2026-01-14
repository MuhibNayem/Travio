package engine

import (
	"context"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Rule represents a pricing rule
type Rule struct {
	ID              string
	Name            string
	Condition       string  // expr expression
	Multiplier      float64 // e.g., 1.20 for 20% increase, 0.85 for 15% discount
	AdjustmentType  string  // multiplier, additive, override
	AdjustmentValue float64
	Priority        int
	compiledPrg     *vm.Program
}

// Environment provides variables for rule evaluation
type Environment struct {
	BasePrice          int64   `expr:"base_price"`
	SeatClass          string  `expr:"seat_class"`
	SeatCategory       string  `expr:"seat_category"`
	DayOfWeek          string  `expr:"day_of_week"` // "Monday", "Tuesday", etc.
	DaysUntilDeparture int     `expr:"days_until_departure"`
	OccupancyRate      float64 `expr:"occupancy_rate"` // 0.0 to 1.0
	Quantity           int     `expr:"quantity"`
	IsHoliday          bool    `expr:"is_holiday"`
	Hour               int     `expr:"hour"`
	Minute             int     `expr:"minute"`
	TripID             string  `expr:"trip_id"`
	RouteID            string  `expr:"route_id"`
	ScheduleID         string  `expr:"schedule_id"`
	FromStationID      string  `expr:"from_station_id"`
	ToStationID        string  `expr:"to_station_id"`
	VehicleType        string  `expr:"vehicle_type"`
	VehicleClass       string  `expr:"vehicle_class"`
	PromoCode          string  `expr:"promo_code"`
}

// AppliedRule represents a rule that was applied during calculation
type AppliedRule struct {
	RuleID     string
	RuleName   string
	Multiplier float64
}

// RulesEngine evaluates pricing rules
type RulesEngine struct {
	rules []*Rule
}

// NewRulesEngine creates a new rules engine with pre-compiled rules
func NewRulesEngine(rules []*Rule) (*RulesEngine, error) {
	for _, rule := range rules {
		prg, err := expr.Compile(rule.Condition, expr.Env(Environment{}), expr.AsBool())
		if err != nil {
			return nil, err
		}
		rule.compiledPrg = prg
	}
	return &RulesEngine{rules: rules}, nil
}

// Evaluate calculates the final price by applying all matching rules
func (e *RulesEngine) Evaluate(ctx context.Context, basePrice int64, env Environment) (int64, []AppliedRule, error) {
	env.BasePrice = basePrice

	price := float64(basePrice)
	var applied []AppliedRule

	for _, rule := range e.rules {
		if rule.compiledPrg == nil {
			continue
		}

		result, err := expr.Run(rule.compiledPrg, env)
		if err != nil {
			continue // Skip rule on error
		}

		if match, ok := result.(bool); ok && match {
			adjustmentType := rule.AdjustmentType
			if adjustmentType == "" {
				adjustmentType = "multiplier"
			}
			switch adjustmentType {
			case "override":
				if rule.AdjustmentValue > 0 {
					price = rule.AdjustmentValue
				}
			case "additive":
				price += rule.AdjustmentValue
			default:
				multiplier := rule.Multiplier
				if multiplier == 0 {
					multiplier = 1
				}
				price *= multiplier
			}
			applied = append(applied, AppliedRule{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				Multiplier: rule.Multiplier,
			})
		}
	}

	return int64(price), applied, nil
}

// CreateEnvironment creates an environment from request parameters
func CreateEnvironment(params EnvironmentParams) Environment {
	parsedDate, _ := time.Parse("2006-01-02", params.Date)
	daysUntil := int(time.Until(parsedDate).Hours() / 24)
	if daysUntil < 0 {
		daysUntil = 0
	}

	hour := time.Now().Hour()
	minute := time.Now().Minute()
	if params.DepartureTime > 0 {
		departure := time.Unix(params.DepartureTime, 0)
		hour = departure.Hour()
		minute = departure.Minute()
	}

	return Environment{
		SeatClass:          params.SeatClass,
		SeatCategory:       params.SeatCategory,
		DayOfWeek:          parsedDate.Weekday().String(),
		DaysUntilDeparture: daysUntil,
		OccupancyRate:      params.OccupancyRate,
		Quantity:           params.Quantity,
		Hour:               hour,
		Minute:             minute,
		TripID:             params.TripID,
		RouteID:            params.RouteID,
		ScheduleID:         params.ScheduleID,
		FromStationID:      params.FromStationID,
		ToStationID:        params.ToStationID,
		VehicleType:        params.VehicleType,
		VehicleClass:       params.VehicleClass,
		PromoCode:          params.PromoCode,
	}
}

type EnvironmentParams struct {
	SeatClass     string
	SeatCategory  string
	Date          string
	Quantity      int
	OccupancyRate float64
	TripID        string
	RouteID       string
	ScheduleID    string
	FromStationID string
	ToStationID   string
	VehicleType   string
	VehicleClass  string
	PromoCode     string
	DepartureTime int64
}
