package validators

import (
	"errors"
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
	if got.Error() != expect.Error() {
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
		// case 1 donation case sentitive
		err := ValidateTaxRequest(models.TaxRequest{
			Allowances: []models.Allowance{
				{Type: "Donation", Amount: -1},
			},
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrAllowanceTypeInvalid, err)

		// case 2 k-receipt case sentitive
		err = ValidateTaxRequest(models.TaxRequest{
			Allowances: []models.Allowance{
				{Type: "K-Receipt", Amount: -1},
			},
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrAllowanceTypeInvalid, err)

		// case 3 kReceipt (camel case)
		err = ValidateTaxRequest(models.TaxRequest{
			Allowances: []models.Allowance{
				{Type: "kReceipt", Amount: -1},
			},
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrAllowanceTypeInvalid, err)

		// case 4 k_receipt (snake case)
		err = ValidateTaxRequest(models.TaxRequest{
			Allowances: []models.Allowance{
				{Type: "k_receipt", Amount: -1},
			},
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrAllowanceTypeInvalid, err)
	})
	t.Run("given allowance amount is invalid should get error 'ErrAllowanceAmountInvalid'", func(t *testing.T) {
		// case 1 donation type
		err := ValidateTaxRequest(models.TaxRequest{
			Allowances: []models.Allowance{
				{Type: models.DonationSlug, Amount: 2000},
				{Type: models.DonationSlug, Amount: -1},
			},
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrAllowanceAmountInvalid, err)

		// case 2 k-receipt type
		err = ValidateTaxRequest(models.TaxRequest{
			Allowances: []models.Allowance{
				{Type: models.DonationSlug, Amount: 2000},
				{Type: models.KReceiptSlug, Amount: 2000},
				{Type: models.KReceiptSlug, Amount: -1},
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
				{Type: models.KReceiptSlug, Amount: 0},
				{Type: models.KReceiptSlug, Amount: 50000},
			},
		})

		assertIsNil(t, err)
	})
}

func TestValidateTaxCsv(t *testing.T) {
	t.Run("given only income invalid should get error 'ErrTotalIncomeInvalid'", func(t *testing.T) {
		err := ValidateTaxCsv(models.TaxCsv{
			TotalIncome: -1,
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrTotalIncomeInvalid, err)
	})
	t.Run("given only wht invalid should get error 'ErrWhtInvalid'", func(t *testing.T) {
		err := ValidateTaxCsv(models.TaxCsv{
			Wht: -1,
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrWhtInvalid, err)
	})
	t.Run("given wht more than income should get error 'ErrWhtMoreThanIncome'", func(t *testing.T) {
		err := ValidateTaxCsv(models.TaxCsv{
			Wht:         30000.01,
			TotalIncome: 30000,
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrWhtMoreThanIncome, err)
	})
	t.Run("given invalid donation should get error 'donation amount should be more than or equal 0'", func(t *testing.T) {
		err := ValidateTaxCsv(models.TaxCsv{
			Donation: -1,
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, errors.New("donation amount should be more than or equal 0"), err)
	})
	t.Run("given invalid k-receipt should get error 'k-receipt amount should be more than or equal 0'", func(t *testing.T) {
		err := ValidateTaxCsv(models.TaxCsv{
			KReceipt: -1,
		})

		assertIsNotNil(t, err)
		assertErrorMessage(t, errors.New("k-receipt amount should be more than or equal 0"), err)
	})
	t.Run("given valid csv data should not get error", func(t *testing.T) {
		err := ValidateTaxCsv(models.TaxCsv{
			Wht:         25000,
			TotalIncome: 500000,
			Donation:    20000,
			KReceipt:    0,
		})

		assertIsNil(t, err)
	})
}
