package services

import (
	"fmt"
	"testing"

	"github.com/baronight/assessment-tax/models"
)

type TestCaseTaxSuite struct {
	name   string
	stub   StubStore
	want   models.TaxResponse
	params models.TaxRequest
}

type StubStore struct {
}

func assertIsNil(t *testing.T, obj interface{}, message string) {
	t.Helper()
	if obj != nil {
		t.Error(message)
	}
}
func assertIsEqual(t *testing.T, want, got interface{}, message string) {
	t.Helper()
	if want != got {
		t.Error(message)
	}
}

func setupTaxService(stub StubStore) *TaxService {
	service := NewTaxService(stub)

	return service
}

func TestTaxCalculate(t *testing.T) {
	t.Run("given input only total income and personal deduction is 60000", func(t *testing.T) {
		expectNilErrMsg := "unexpect error should be null"
		expectTaxValueMsg := func(want, got float32) string {
			return fmt.Sprintf("expect tax should be %.2f, but got %.2f", want, got)
		}
		testSuites := []TestCaseTaxSuite{
			{
				name:   "when total income is lower than 150_000 then tax should be 0",
				want:   models.TaxResponse{Tax: 0},
				params: models.TaxRequest{TotalIncome: 40_000},
			},
			{
				name:   "when total income is 210_000 then tax should be 0",
				want:   models.TaxResponse{Tax: 0},
				params: models.TaxRequest{TotalIncome: 210_000},
			},
			{
				name:   "when total income is 500_000 then tax should be 29000",
				want:   models.TaxResponse{Tax: 29_000},
				params: models.TaxRequest{TotalIncome: 500_000},
			},
			{
				name:   "when total income is 560_000 then tax should be 35_000",
				want:   models.TaxResponse{Tax: 35_000},
				params: models.TaxRequest{TotalIncome: 560_000},
			},
			{
				name:   "when total income is 560_000.01 then tax should be 35_000.0015",
				want:   models.TaxResponse{Tax: 35_000.0015},
				params: models.TaxRequest{TotalIncome: 560_000.01},
			},
			{
				name:   "when total income is 560_001 then tax should be 35_000.15",
				want:   models.TaxResponse{Tax: 35_000.15},
				params: models.TaxRequest{TotalIncome: 560_001},
			},
			{
				name:   "when total income is 1_060_000 then tax should be 110_000",
				want:   models.TaxResponse{Tax: 110_000},
				params: models.TaxRequest{TotalIncome: 1_060_000},
			},
			{
				name:   "when total income is 1_100_000 then tax should be 118_000",
				want:   models.TaxResponse{Tax: 118_000},
				params: models.TaxRequest{TotalIncome: 1_100_000},
			},
			{
				name:   "when total income is 2_060_000 then tax should be 310_000",
				want:   models.TaxResponse{Tax: 310_000},
				params: models.TaxRequest{TotalIncome: 2_060_000},
			},
			{
				name:   "when total income is over 2_060_001 then tax should be 310_000.35",
				want:   models.TaxResponse{Tax: 310_000.35},
				params: models.TaxRequest{TotalIncome: 2_060_001},
			},
		}

		for _, tc := range testSuites {
			t.Run(tc.name, func(t *testing.T) {
				service := setupTaxService(tc.stub)

				result, err := service.TaxCalculate(tc.params)

				assertIsNil(t, err, expectNilErrMsg)
				assertIsEqual(t, tc.want.Tax, result.Tax, expectTaxValueMsg(tc.want.Tax, result.Tax))
			})
		}
	})
}
