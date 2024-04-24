package services

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/baronight/assessment-tax/models"
)

type TaxTestSuite struct {
	name      string
	stub      StubStore
	want      models.TaxResponse
	params    models.TaxRequest
	wantError error
}

type StubStore struct {
	deductions      []models.Deduction
	err             error
	expectToCall    map[string]bool
	expectCallTimes map[string]int
}

func (s *StubStore) GetDeductions() ([]models.Deduction, error) {
	s.expectToCall["GetDeductions"] = true
	s.expectCallTimes["GetDeductions"]++
	return s.deductions, s.err
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
	service := NewTaxService(&stub)

	return service
}

var expectNilErrMsg = "unexpect error should be null"

func expectTaxValueMsg(want, got float32) string {
	return fmt.Sprintf("expect tax should be %.2f, but got %.2f", want, got)
}
func expectTaxRefundValueMsg(want, got float32) string {
	return fmt.Sprintf("expect tax refund should be %.2f, but got %.2f", want, got)
}
func assertObjectIsEqual(t *testing.T, want, got interface{}) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("expect object should be %#v, but got %#v", want, got)
	}
}

func initStub(deductions []models.Deduction, err error) StubStore {
	return StubStore{
		expectToCall:    map[string]bool{},
		expectCallTimes: map[string]int{},
		deductions:      deductions,
		err:             err,
	}
}

func TestTaxCalculate(t *testing.T) {
	t.Run("given input only total income and personal deduction is 60000", func(t *testing.T) {
		testSuites := []TaxTestSuite{
			{
				name:   "when total income is lower than 150_000 then tax should be 0",
				want:   models.TaxResponse{Tax: 0},
				params: models.TaxRequest{TotalIncome: 40_000},
				stub:   initStub([]models.Deduction{}, nil),
			},
			{
				name:   "when total income is 210_000 then tax should be 0",
				want:   models.TaxResponse{Tax: 0},
				params: models.TaxRequest{TotalIncome: 210_000},
				stub:   initStub([]models.Deduction{}, nil),
			},
			{
				name:   "when total income is 500_000 then tax should be 29000",
				want:   models.TaxResponse{Tax: 29_000},
				params: models.TaxRequest{TotalIncome: 500_000},
				stub:   initStub([]models.Deduction{}, nil),
			},
			{
				name:   "when total income is 560_000 then tax should be 35_000",
				want:   models.TaxResponse{Tax: 35_000},
				params: models.TaxRequest{TotalIncome: 560_000},
				stub:   initStub([]models.Deduction{}, nil),
			},
			{
				name:   "when total income is 560_000.01 then tax should be 35_000.0015",
				want:   models.TaxResponse{Tax: 35_000.0015},
				params: models.TaxRequest{TotalIncome: 560_000.01},
				stub:   initStub([]models.Deduction{}, nil),
			},
			{
				name:   "when total income is 560_001 then tax should be 35_000.15",
				want:   models.TaxResponse{Tax: 35_000.15},
				params: models.TaxRequest{TotalIncome: 560_001},
				stub:   initStub([]models.Deduction{}, nil),
			},
			{
				name:   "when total income is 1_060_000 then tax should be 110_000",
				want:   models.TaxResponse{Tax: 110_000},
				params: models.TaxRequest{TotalIncome: 1_060_000},
				stub:   initStub([]models.Deduction{}, nil),
			},
			{
				name:   "when total income is 1_100_000 then tax should be 118_000",
				want:   models.TaxResponse{Tax: 118_000},
				params: models.TaxRequest{TotalIncome: 1_100_000},
				stub:   initStub([]models.Deduction{}, nil),
			},
			{
				name:   "when total income is 2_060_000 then tax should be 310_000",
				want:   models.TaxResponse{Tax: 310_000},
				params: models.TaxRequest{TotalIncome: 2_060_000},
				stub:   initStub([]models.Deduction{}, nil),
			},
			{
				name:   "when total income is over 2_060_001 then tax should be 310_000.35",
				want:   models.TaxResponse{Tax: 310_000.35},
				params: models.TaxRequest{TotalIncome: 2_060_001},
				stub:   initStub([]models.Deduction{}, nil),
			},
		}

		for _, tc := range testSuites {
			t.Run(tc.name, func(t *testing.T) {
				service := setupTaxService(tc.stub)

				result, err := service.TaxCalculate(tc.params)

				assertIsNil(t, err, expectNilErrMsg)
				assertIsEqual(t, tc.want.Tax, result.Tax, expectTaxValueMsg(tc.want.Tax, result.Tax))
				assertIsEqual(t, tc.want.TaxRefund, result.TaxRefund, expectTaxRefundValueMsg(tc.want.TaxRefund, result.TaxRefund))
			})
		}
	})
}

func TestTaxWithWHT(t *testing.T) {
	t.Run("given input with income and wht with personal deduction is 60_000", func(t *testing.T) {
		testSuites := []TaxTestSuite{
			{
				name:   "when input wht = 25_000 and income = 500_000 then tax should be 4_000",
				want:   models.TaxResponse{Tax: 4_000},
				params: models.TaxRequest{TotalIncome: 500_000, Wht: 25_000},
				stub:   initStub([]models.Deduction{}, nil),
			},
			{
				name:   "when input wht = 30_000 and income = 500_000 then tax should be 0 and taxRefund should be 1_000",
				want:   models.TaxResponse{TaxRefund: 1_000},
				params: models.TaxRequest{TotalIncome: 500_000, Wht: 30_000},
				stub:   initStub([]models.Deduction{}, nil),
			},
		}

		for _, tc := range testSuites {
			t.Run(tc.name, func(t *testing.T) {
				service := setupTaxService(tc.stub)

				result, err := service.TaxCalculate(tc.params)

				assertIsNil(t, err, expectNilErrMsg)
				assertIsEqual(t, tc.want.Tax, result.Tax, expectTaxValueMsg(tc.want.Tax, result.Tax))
				assertIsEqual(t, tc.want.TaxRefund, result.TaxRefund, expectTaxRefundValueMsg(tc.want.TaxRefund, result.TaxRefund))
			})
		}
	})
}

func TestTaxWithAllowance(t *testing.T) {

	t.Run("given input donation allowances", func(t *testing.T) {
		testSuites := []TaxTestSuite{
			{
				name:      "when error on get deduction config it should return error",
				wantError: errors.New("error xxx happend"),
				stub:      initStub([]models.Deduction{}, errors.New("error xxx happend")),
			},
			{
				name:   "when error on get deduction is 'ErrNoRows' it should return tax",
				want:   models.TaxResponse{Tax: 29_000},
				params: models.TaxRequest{TotalIncome: 500_000, Allowances: []models.Allowance{}},
				stub:   initStub([]models.Deduction{}, sql.ErrNoRows),
			},
			{
				name: "when no limit donation deduction it should subtract all donation from tax",
				stub: initStub([]models.Deduction{}, nil),
				want: models.TaxResponse{Tax: 3_200},
				params: models.TaxRequest{
					TotalIncome: 500_000,
					Wht:         25_000,
					Allowances: []models.Allowance{
						{
							Type:   models.DonationSlug,
							Amount: 5_000,
						},
						{
							Type:   models.DonationSlug,
							Amount: 3_000,
						},
					},
				},
			},
			{
				name: "when donation deduction has limit it should subtract with no over limit from tax",
				stub: initStub(
					[]models.Deduction{
						{Slug: models.DonationSlug, Amount: 5_000, Name: "Donation"},
						{Slug: models.PersonalSlug, Amount: 50_000, Name: "PersonalDeduction"},
					},
					nil,
				),
				params: models.TaxRequest{
					TotalIncome: 500_000,
					Wht:         25_000,
					Allowances: []models.Allowance{
						{Type: models.DonationSlug, Amount: 4_000},
						{Type: models.DonationSlug, Amount: 2_000},
					},
				},
				want: models.TaxResponse{Tax: 4_500},
			},
		}

		for _, tc := range testSuites {
			t.Run(tc.name, func(t *testing.T) {
				service := setupTaxService(tc.stub)

				result, err := service.TaxCalculate(tc.params)

				if tc.wantError != nil {
					if err == nil {
						t.Fatalf("expect error should not null")
					}
					assertIsEqual(t, tc.wantError.Error(), err.Error(), fmt.Sprintf("expect error %s but got %s", tc.wantError.Error(), err.Error()))
				} else {
					assertIsNil(t, err, expectNilErrMsg)
					assertIsEqual(t, tc.want.Tax, result.Tax, expectTaxValueMsg(tc.want.Tax, result.Tax))
					assertIsEqual(t, tc.want.TaxRefund, result.TaxRefund, expectTaxRefundValueMsg(tc.want.TaxRefund, result.TaxRefund))
				}
			})
		}
	})
}

func TestTaxLevel(t *testing.T) {
	t.Run("given valid tax request should return response with tax level", func(t *testing.T) {
		stub := initStub([]models.Deduction{}, nil)
		params := models.TaxRequest{
			TotalIncome: 500000.0,
			Wht:         0.0,
			Allowances: []models.Allowance{
				{
					Type:   models.DonationSlug,
					Amount: 200000.0,
				},
			},
		}
		service := setupTaxService(stub)

		result, err := service.TaxCalculate(params)

		want := models.TaxResponse{
			Tax: 19_000,
			TaxLevel: []models.TaxLevel{
				{
					Level: "0-150,000",
					Tax:   0.0,
				},
				{
					Level: "150,001-500,000",
					Tax:   19000.0,
				},
				{
					Level: "500,001-1,000,000",
					Tax:   0.0,
				},
				{
					Level: "1,000,001-2,000,000",
					Tax:   0.0,
				},
				{
					Level: "2,000,001 ขึ้นไป",
					Tax:   0.0,
				},
			},
		}
		assertIsNil(t, err, expectNilErrMsg)
		assertObjectIsEqual(t, want, result)
	})
}
