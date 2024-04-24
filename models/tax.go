package models

type TaxRequest struct {
	TotalIncome float32     `json:"totalIncome" validate:"gte=0" example:"500000"`
	Wht         float32     `json:"wht,omitempty" validate:"omitempty,ltefield=totalIncome,gte=0"`
	Allowances  []Allowance `json:"allowances,omitempty" validate:"omitempty,dive"`
} //@Name TaxRequest

type Allowance struct {
	Type   string  `json:"allowanceType" validate:"required,oneof=donation k-receipt"`
	Amount float32 `json:"amount" validate:"gte=0"`
} //@Name Allowance

type TaxResponse struct {
	Tax       float32 `json:"tax"`
	TaxRefund float32 `json:"taxRefund,omitempty"`
} //@Name TaxResponse

type TaxStep struct {
	MinIncome float32
	MaxIncome float32
	Rate      float32
}

type Deduction struct {
	Id        uint    `postgres:"id" json:"-"`
	Slug      string  `postgres:"slug" json:"slug"`
	Name      string  `postgres:"name" json:"name"`
	Amount    float32 `postgres:"amount" json:"amount"`
	MinAmount float32 `postgres:"minAmount" json:"-"`
	MaxAmount float32 `postgres:"maxAmount" json:"-"`
}

const (
	DonationSlug = "donation"
	PersonalSlug = "personal"
)
