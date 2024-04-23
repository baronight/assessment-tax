package models

type TaxRequest struct {
	TotalIncome float32     `json:"totalIncome" validate:"gte=0"`
	Wht         float32     `json:"wht,omitempty" validate:"omitempty,ltefield=totalIncome,gte=0"`
	Allowances  []Allowance `json:"allowances,omitempty" validate:"omitempty,dive"`
}

type Allowance struct {
	Type   string  `json:"allowanceType" validate:"required,oneof=donation k-receipt"`
	Amount float32 `json:"amount" validate:"gte=0"`
}

type TaxResponse struct {
	Tax float32 `json:"tax"`
}

type TaxStep struct {
	MinIncome float32
	MaxIncome float32
	Rate      float32
}
