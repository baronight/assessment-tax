package models

const (
	DonationSlug = "donation"
	PersonalSlug = "personal"
	KReceiptSlug = "k-receipt"
)

type TaxRequest struct {
	TotalIncome float64     `json:"totalIncome" validate:"gte=0" example:"500000"`
	Wht         float64     `json:"wht,omitempty" validate:"omitempty,ltefield=totalIncome,gte=0"`
	Allowances  []Allowance `json:"allowances,omitempty" validate:"omitempty,dive"`
} //@Name TaxRequest

type Allowance struct {
	Type   string  `json:"allowanceType" validate:"required,oneof=donation k-receipt"`
	Amount float64 `json:"amount" validate:"gte=0"`
} //@Name Allowance

type TaxResponse struct {
	Tax       float64    `json:"tax"`
	TaxRefund float64    `json:"taxRefund,omitempty"`
	TaxLevel  []TaxLevel `json:"taxLevel"`
} //@Name TaxResponse

type TaxStep struct {
	MinIncome float64
	MaxIncome float64
	Rate      float64
} //@Name TaxStep

type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
} //@Name TaxLevel

type Deduction struct {
	Id        uint    `postgres:"id" json:"-"`
	Slug      string  `postgres:"slug" json:"slug"`
	Name      string  `postgres:"name" json:"name"`
	Amount    float64 `postgres:"amount" json:"amount"`
	MinAmount float64 `postgres:"minAmount" json:"-"`
	MaxAmount float64 `postgres:"maxAmount" json:"-"`
} //@Name Deduction

type TaxCsv struct {
	TotalIncome float64 `csv:"totalIncome"`
	Wht         float64 `csv:"wht"`
	Donation    float64 `csv:"donation"`
	KReceipt    float64 `csv:"k-receipt,omitempty"`
}

type TaxCsvResponse struct {
	Taxes []CsvCalculateResult `json:"taxes"`
} //@Name TaxCsvResponse

type CsvCalculateResult struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax"`
	TaxRefund   float64 `json:"taxRefund,omitempty"`
} //@Name CsvCalculateResult
