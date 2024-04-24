package services

import (
	"database/sql"

	"github.com/baronight/assessment-tax/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type TaxService struct {
	Db TaxStorer
}

type TaxStorer interface {
	GetDeductions() ([]models.Deduction, error)
}

var (
	DefaultPersonalDeduction float32 = 60_000
	DefaultDonationDeduction float32 = 100_000
)

var TaxStep []models.TaxStep = []models.TaxStep{
	{MinIncome: -1, MaxIncome: 150_000, Rate: 0},
	{MinIncome: 150_000, MaxIncome: 500_000, Rate: 0.1},
	{MinIncome: 500_000, MaxIncome: 1_000_000, Rate: 0.15},
	{MinIncome: 1_000_000, MaxIncome: 2_000_000, Rate: 0.2},
	{MinIncome: 2_000_000, MaxIncome: 0, Rate: 0.35},
}

func NewTaxService(db TaxStorer) *TaxService {
	return &TaxService{
		Db: db,
	}
}

func (ts *TaxService) GetDeductionConfig() (personal, donation models.Deduction, err error) {
	var deductions map[string]models.Deduction = map[string]models.Deduction{}
	ds, err := ts.Db.GetDeductions()
	if err != nil && err != sql.ErrNoRows {
		return personal, donation, err
	}

	for _, v := range ds {
		deductions[v.Slug] = v
		switch v.Slug {
		case models.DonationSlug:
			donation = v
		case models.PersonalSlug:
			personal = v
		}
	}

	if personal.Amount == 0 {
		personal.Amount = DefaultPersonalDeduction
	}

	if donation.Amount == 0 {
		donation.Amount = DefaultDonationDeduction
	}

	return personal, donation, nil
}

func CalculateDonation(allowances []models.Allowance, donation models.Deduction) (amount float32) {
	for _, allowance := range allowances {
		if allowance.Type == models.DonationSlug {
			amount += allowance.Amount
		}
	}

	if donation.Amount != 0 && amount > donation.Amount {
		amount = donation.Amount
	}
	return
}

func (ts *TaxService) TaxCalculate(tax models.TaxRequest) (models.TaxResponse, error) {
	personal, donation, err := ts.GetDeductionConfig()
	if err != nil {
		return models.TaxResponse{}, err
	}

	netIncome := tax.TotalIncome - personal.Amount - CalculateDonation(tax.Allowances, donation)
	var result models.TaxResponse
	result.TaxLevel = []models.TaxLevel{}
	for _, v := range TaxStep {
		var taxStep float32
		p := message.NewPrinter(language.English)
		level := p.Sprintf("%.0f-%.0f", v.MinIncome+1, v.MaxIncome)
		overflowStep := netIncome - v.MaxIncome
		if v.MaxIncome <= 0 {
			// that mean unlimit ceiling income
			overflowStep = 0
			level = p.Sprintf("%.0f ขึ้นไป", v.MinIncome+1)
		}
		if overflowStep > 0 {
			// calculate full tax rate on this step
			taxStep = (v.MaxIncome - v.MinIncome) * v.Rate
		} else {
			// calculate remain tax
			remain := netIncome - v.MinIncome
			if remain < 0 {
				remain = 0
			}
			taxStep = remain * v.Rate
		}

		result.Tax += taxStep
		result.TaxLevel = append(result.TaxLevel, models.TaxLevel{Level: level, Tax: taxStep})
	}

	if tax.Wht > result.Tax {
		// over payment tax should refund
		result.TaxRefund = tax.Wht - result.Tax
		result.Tax = 0
	} else {
		result.Tax = result.Tax - tax.Wht
	}

	return result, nil
}
