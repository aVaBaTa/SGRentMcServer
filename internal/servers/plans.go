package servers

import "fmt"

type Plan struct {
	Name       string
	RAMMb      int64
	CPUCores   float64
	MaxSlots   int
	Free       bool
	PriceCents int // prix mensuel en cents USD (0 = gratuit)
}

var plans = map[string]Plan{
	"free":     {Name: "free", RAMMb: 1024, CPUCores: 1.0, MaxSlots: 5, Free: true, PriceCents: 0},
	"starter":  {Name: "starter", RAMMb: 2048, CPUCores: 1.0, MaxSlots: 20, PriceCents: 300},
	"standard": {Name: "standard", RAMMb: 4096, CPUCores: 2.0, MaxSlots: 50, PriceCents: 700},
	"pro":      {Name: "pro", RAMMb: 8192, CPUCores: 4.0, MaxSlots: 100, PriceCents: 1400},
	"extreme":  {Name: "extreme", RAMMb: 16384, CPUCores: 6.0, MaxSlots: 200, PriceCents: 2500},
}

// PriceString retourne le prix formaté en dollars (ex: "7.00").
func (p Plan) PriceString() string {
	return fmt.Sprintf("%d.%02d", p.PriceCents/100, p.PriceCents%100)
}

func GetPlan(name string) (Plan, error) {
	p, ok := plans[name]
	if !ok {
		return Plan{}, fmt.Errorf("unknown plan %q", name)
	}
	return p, nil
}
