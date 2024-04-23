package services

import (
	"github.com/baronight/assessment-tax/models"
)

type TaxService struct {
	Db TaxStorer
}

type TaxStorer interface {
}

const (
	personal = 60000
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

func (ts *TaxService) TaxCalculate(tax models.TaxRequest) (models.TaxResponse, error) {
	netIncome := tax.TotalIncome - personal
	var result models.TaxResponse
	for _, v := range TaxStep {
		overflowStep := netIncome - v.MaxIncome
		if v.MaxIncome == 0 {
			// that mean unlimit ceiling income
			overflowStep = 0
		}
		if overflowStep > 0 {
			// calculate full tax rate on this step
			result.Tax += (v.MaxIncome - v.MinIncome) * v.Rate
		} else {
			// calculate remain tax
			remain := netIncome - v.MinIncome
			if remain < 0 {
				remain = 0
			}
			result.Tax += remain * v.Rate
		}
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
