package saas

type Plan struct {
	ID              string
	Name            string
	MaxTrips        int
	MaxUsers        int
	MaxScheduleDays int // How far in future they can create trips
	PriceMonthly    float64
}

var (
	PlanStarter = Plan{
		ID:              "plan_starter",
		Name:            "Starter",
		MaxTrips:        10,
		MaxUsers:        5,
		MaxScheduleDays: 10, // User requirement: Min 10 days
		PriceMonthly:    29.00,
	}
	PlanGrowth = Plan{
		ID:              "plan_growth",
		Name:            "Growth",
		MaxTrips:        100,
		MaxUsers:        20,
		MaxScheduleDays: 30, // 1 Month
		PriceMonthly:    99.00,
	}
	PlanEnterprise = Plan{
		ID:              "plan_enterprise",
		Name:            "Enterprise",
		MaxTrips:        10000,
		MaxUsers:        1000,
		MaxScheduleDays: 60, // User requirement: Max 2 months
		PriceMonthly:    499.00,
	}
)

func GetPlan(id string) Plan {
	switch id {
	case PlanStarter.ID:
		return PlanStarter
	case PlanGrowth.ID:
		return PlanGrowth
	case PlanEnterprise.ID:
		return PlanEnterprise
	default:
		return PlanStarter // Default to starter
	}
}
