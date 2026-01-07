package engine

import (
	"context"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Rule represents a pricing rule
type Rule struct {
	ID          string
	Name        string
	Condition   string  // expr expression
	Multiplier  float64 // e.g., 1.20 for 20% increase, 0.85 for 15% discount
	Priority    int
	compiledPrg *vm.Program
}

// Environment provides variables for rule evaluation
type Environment struct {
	BasePrice          int64   `expr:"base_price"`
	SeatClass          string  `expr:"seat_class"`
	DayOfWeek          string  `expr:"day_of_week"` // "Monday", "Tuesday", etc.
	DaysUntilDeparture int     `expr:"days_until_departure"`
	OccupancyRate      float64 `expr:"occupancy_rate"` // 0.0 to 1.0
	Quantity           int     `expr:"quantity"`
	IsHoliday          bool    `expr:"is_holiday"`
	Hour               int     `expr:"hour"`
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
			price *= rule.Multiplier
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
func CreateEnvironment(seatClass, date string, quantity int, occupancyRate float64) Environment {
	parsedDate, _ := time.Parse("2006-01-02", date)
	daysUntil := int(time.Until(parsedDate).Hours() / 24)
	if daysUntil < 0 {
		daysUntil = 0
	}

	return Environment{
		SeatClass:          seatClass,
		DayOfWeek:          parsedDate.Weekday().String(),
		DaysUntilDeparture: daysUntil,
		OccupancyRate:      occupancyRate,
		Quantity:           quantity,
		Hour:               time.Now().Hour(),
	}
}
