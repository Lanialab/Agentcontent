package brand

import "time"

type Blueprint struct {
	Love        string    `json:"love"`
	GoodAt      string    `json:"goodAt"`
	WorldNeeds  string    `json:"worldNeeds"`
	PaidFor     string    `json:"paidFor"`
	Positioning string    `json:"positioning"`
	Voice       string    `json:"voice"`
	Pillars     []string  `json:"pillars"`
	Topics      []string  `json:"topics"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (b Blueprint) Intersection() string {
	if b.Love == "" || b.GoodAt == "" || b.WorldNeeds == "" || b.PaidFor == "" {
		return ""
	}
	return "Ikigai: " + b.Love + " × " + b.GoodAt + " × " + b.WorldNeeds + " × " + b.PaidFor
}
