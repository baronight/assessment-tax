package validators

import (
	"testing"

	"github.com/baronight/assessment-tax/models"
)

func assertIsNil(t *testing.T, obj interface{}) {
	t.Helper()
	if obj != nil {
		t.Error("expect this object should be null")
	}
}

func assertIsNotNil(t *testing.T, obj interface{}) {
	t.Helper()
	if obj == nil {
		t.Error("expect this object should not be null")
	}
}

func assertErrorMessage(t *testing.T, expect, got error) {
	t.Helper()
	if got != expect {
		t.Errorf("expect error is %s but got %s", expect.Error(), got.Error())
	}
}

func TestValidateTaxRequest(t *testing.T) {
	t.Run("given only income invalid should get error 'ErrTotalIncomeInvalid'", func(t *testing.T) {
		err := ValidateTaxRequest(models.TaxRequest{
			TotalIncome: -1,
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrTotalIncomeInvalid, err)
	})
	t.Run("given only wht invalid should get error 'ErrWhtInvalid'", func(t *testing.T) {
		err := ValidateTaxRequest(models.TaxRequest{
			Wht: -1,
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrWhtInvalid, err)
	})
	t.Run("given wht more than income should get error 'ErrWhtMoreThanIncome'", func(t *testing.T) {
		err := ValidateTaxRequest(models.TaxRequest{
			Wht:         30000.01,
			TotalIncome: 30000,
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrWhtMoreThanIncome, err)
	})
	t.Run("given allowance type is invalid should get error 'ErrAllowanceTypeInvalid'", func(t *testing.T) {
		err := ValidateTaxRequest(models.TaxRequest{
			Allowances: []models.Allowance{
				{Type: "Donation", Amount: -1},
			},
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrAllowanceTypeInvalid, err)
	})
	t.Run("given allowance amount is invalid should get error 'ErrAllowanceAmountInvalid'", func(t *testing.T) {
		err := ValidateTaxRequest(models.TaxRequest{
			Allowances: []models.Allowance{
				{Type: models.DonationSlug, Amount: 2000},
				{Type: models.DonationSlug, Amount: -1},
			},
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrAllowanceAmountInvalid, err)
	})
	t.Run("given valid tax request should not get error", func(t *testing.T) {
		err := ValidateTaxRequest(models.TaxRequest{
			Wht:         25000,
			TotalIncome: 500000,
			Allowances: []models.Allowance{
				{Type: models.DonationSlug, Amount: 2000},
				{Type: models.DonationSlug, Amount: 250},
			},
		})

		assertIsNil(t, err)
	})
}
