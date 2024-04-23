package validators

import (
	"errors"

	"github.com/baronight/assessment-tax/models"
)

type AllowanceType string

const (
	Donation AllowanceType = "donation"
	KReceipt AllowanceType = "k-receipt"
)

var (
	ErrTotalIncomeInvalid     = errors.New("total income should be more than or equal 0")
	ErrWhtInvalid             = errors.New("wht should be more than or equal 0")
	ErrWhtMoreThanIncome      = errors.New("wht should not more than income")
	ErrAllowanceTypeInvalid   = errors.New("allowance type should be one of 'donation', 'k-receipt'")
	ErrAllowanceAmountInvalid = errors.New("allowance amount should be more than or equal 0")
)

func ValidateTaxRequest(tax models.TaxRequest) error {
	if err := ValidateTotalIncome(tax.TotalIncome); err != nil {
		return err
	}
	if err := ValidateWht(tax.Wht, tax.TotalIncome); err != nil {
		return err
	}
	for _, v := range tax.Allowances {
		if v.Type != string(Donation) && v.Type != string(KReceipt) {
			return ErrAllowanceTypeInvalid
		}
		if v.Amount < 0 {
			return ErrAllowanceAmountInvalid
		}
	}
	return nil
}

func ValidateTotalIncome(totalIncome float32) error {
	if totalIncome < 0 {
		return ErrTotalIncomeInvalid
	}
	return nil
}

func ValidateWht(wht, totalIncome float32) error {
	if wht < 0 {
		return ErrWhtInvalid
	}
	if wht > totalIncome {
		return ErrWhtMoreThanIncome
	}
	return nil
}
