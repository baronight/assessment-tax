package services

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/baronight/assessment-tax/models"
)

type TaxTestSuite struct {
	name      string
	stub      StubTaxStore
	want      models.TaxResponse
	params    models.TaxRequest
	wantError error
}

type StubTaxStore struct {
	deductions      []models.Deduction
	err             error
	expectToCall    map[string]bool
	expectCallTimes map[string]int
}

func (s *StubTaxStore) GetDeductions() ([]models.Deduction, error) {
	s.expectToCall["GetDeductions"] = true
	s.expectCallTimes["GetDeductions"]++
	return s.deductions, s.err
}

func (s *StubTaxStore) assertMethodWasCalled(t *testing.T, methodName string) {
	t.Helper()
	if !s.expectToCall[methodName] {
		t.Errorf("expect %s was called", methodName)
	}
}
func (s *StubTaxStore) assertMethodCalledTime(t *testing.T, methodName string, times int) {
	t.Helper()
	if s.expectCallTimes[methodName] != times {
		t.Errorf("expect %s was called %d times but got %d", methodName, times, s.expectCallTimes[methodName])
	}
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
func setupTaxService(stub StubTaxStore) *TaxService {
	service := NewTaxService(&stub)

	return service
}

var expectNilErrMsg = "unexpect error should be null"

func expectTaxValueMsg(want, got float64) string {
	return fmt.Sprintf("expect tax should be %.2f, but got %.2f", want, got)
}
func expectTaxRefundValueMsg(want, got float64) string {
	return fmt.Sprintf("expect tax refund should be %.2f, but got %.2f", want, got)
}
func assertObjectIsEqual(t *testing.T, want, got interface{}) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("expect object should be %#v, but got %#v", want, got)
	}
}

func initStub(deductions []models.Deduction, err error) StubTaxStore {
	return StubTaxStore{
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
				name:   "when total income is 560_000.01 then tax should be 35_000",
				want:   models.TaxResponse{Tax: 35_000.00},
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

				tc.stub.assertMethodWasCalled(t, "GetDeductions")
				tc.stub.assertMethodCalledTime(t, "GetDeductions", 1)
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

				tc.stub.assertMethodWasCalled(t, "GetDeductions")
				tc.stub.assertMethodCalledTime(t, "GetDeductions", 1)
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

				tc.stub.assertMethodWasCalled(t, "GetDeductions")
				tc.stub.assertMethodCalledTime(t, "GetDeductions", 1)
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
	t.Run("given input k-receipt allowances", func(t *testing.T) {
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
				name: "when no limit k-receipt deduction it should subtract all k-receipt from tax",
				stub: initStub([]models.Deduction{}, nil),
				want: models.TaxResponse{Tax: 3_200},
				params: models.TaxRequest{
					TotalIncome: 500_000,
					Wht:         25_000,
					Allowances: []models.Allowance{
						{
							Type:   models.KReceiptSlug,
							Amount: 5_000,
						},
						{
							Type:   models.KReceiptSlug,
							Amount: 3_000,
						},
					},
				},
			},
			{
				name: "when k-receipt deduction has limit it should subtract with no over limit from tax",
				stub: initStub(
					[]models.Deduction{
						{Slug: models.DonationSlug, Amount: 5_000, Name: "Donation"},
						{Slug: models.PersonalSlug, Amount: 50_000, Name: "PersonalDeduction"},
						{Slug: models.KReceiptSlug, Amount: 5_000, Name: "kReceipt"},
					},
					nil,
				),
				params: models.TaxRequest{
					TotalIncome: 500_000,
					Wht:         25_000,
					Allowances: []models.Allowance{
						{Type: models.KReceiptSlug, Amount: 4_000},
						{Type: models.KReceiptSlug, Amount: 2_000},
					},
				},
				want: models.TaxResponse{Tax: 4_500},
			},
		}

		for _, tc := range testSuites {
			t.Run(tc.name, func(t *testing.T) {
				service := setupTaxService(tc.stub)

				result, err := service.TaxCalculate(tc.params)

				tc.stub.assertMethodWasCalled(t, "GetDeductions")
				tc.stub.assertMethodCalledTime(t, "GetDeductions", 1)
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
	t.Run("given input all donation and k-receipt allowances", func(t *testing.T) {
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
				name: "when no limit donation it should subtract all donation from tax",
				stub: initStub(
					[]models.Deduction{
						{Slug: models.DonationSlug, Amount: 0, Name: "Donation"},
						{Slug: models.PersonalSlug, Amount: 50_000, Name: "PersonalDeduction"},
						{Slug: models.KReceiptSlug, Amount: 5_000, Name: "kReceipt"},
					},
					nil,
				),
				want: models.TaxResponse{TaxRefund: 10_500},
				params: models.TaxRequest{
					TotalIncome: 500_000,
					Wht:         25_000,
					Allowances: []models.Allowance{
						{
							Type:   models.DonationSlug,
							Amount: 150_000,
						},
						{
							Type:   models.KReceiptSlug,
							Amount: 150_000,
						},
					},
				},
			},
			{
				name: "when no limit k-receipt deduction it should subtract all k-receipt from tax",
				stub: initStub(
					[]models.Deduction{
						{Slug: models.DonationSlug, Amount: 5_000, Name: "Donation"},
						{Slug: models.PersonalSlug, Amount: 50_000, Name: "PersonalDeduction"},
						{Slug: models.KReceiptSlug, Amount: 0, Name: "kReceipt"},
					},
					nil,
				),
				want: models.TaxResponse{TaxRefund: 10_500},
				params: models.TaxRequest{
					TotalIncome: 500_000,
					Wht:         25_000,
					Allowances: []models.Allowance{
						{
							Type:   models.DonationSlug,
							Amount: 150_000,
						},
						{
							Type:   models.KReceiptSlug,
							Amount: 150_000,
						},
					},
				},
			},
			{
				name: "when donation and k-receipt deduction has limit it should subtract with no over limit from tax",
				stub: initStub(
					[]models.Deduction{
						{Slug: models.DonationSlug, Amount: 3_000, Name: "Donation"},
						{Slug: models.PersonalSlug, Amount: 50_000, Name: "PersonalDeduction"},
						{Slug: models.KReceiptSlug, Amount: 2_000, Name: "kReceipt"},
					},
					nil,
				),
				params: models.TaxRequest{
					TotalIncome: 500_000,
					Wht:         25_000,
					Allowances: []models.Allowance{
						{Type: models.DonationSlug, Amount: 4_000},
						{Type: models.KReceiptSlug, Amount: 2_000},
						{Type: models.KReceiptSlug, Amount: 1_000},
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

				tc.stub.assertMethodWasCalled(t, "GetDeductions")
				tc.stub.assertMethodCalledTime(t, "GetDeductions", 1)
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

		stub.assertMethodWasCalled(t, "GetDeductions")
		stub.assertMethodCalledTime(t, "GetDeductions", 1)
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

func TestFuncExtractCsv(t *testing.T) {
	openCsvFile := func(t *testing.T, filePath string) io.Reader {
		t.Helper()
		dir, _ := os.Getwd()
		fileData, err := os.Open(filepath.Join(dir, filePath))
		if err != nil {
			t.Fatal(err)
		}
		return fileData
	}
	// missing column -> missing required header field
	t.Run("when csv is missing required colum should return error 'missing required header field'", func(t *testing.T) {
		stub := initStub(nil, nil)
		s := setupTaxService(stub)
		fileData := openCsvFile(t, "../testdata/missing-column-taxes.csv")

		result, err := s.ExtractCsv(fileData)

		if len(result) != 0 {
			t.Errorf("expect result should no data")
		}
		if err == nil {
			t.Fatal("error should not be null")
		}
		expect := errors.New("missing required header field")
		assertObjectIsEqual(t, expect, err)
	})
	// missing field -> valud should not be null
	t.Run("when csv is missing value on required field should return error 'value should not be empty'", func(t *testing.T) {
		stub := initStub(nil, nil)
		s := setupTaxService(stub)
		fileData := openCsvFile(t, "../testdata/missing-value-on-required-field-taxes.csv")

		result, err := s.ExtractCsv(fileData)

		if len(result) != 0 {
			t.Errorf("expect result should no data")
		}
		if err == nil {
			t.Fatal("error should not be null")
		}
		expect := errors.New("value should not be empty")
		assertObjectIsEqual(t, expect, err)
	})
	t.Run("when invalid csv field should return error message", func(t *testing.T) {
		stub := initStub(nil, nil)
		s := setupTaxService(stub)
		fileData := openCsvFile(t, "../testdata/invalid-taxes.csv")

		result, err := s.ExtractCsv(fileData)

		if len(result) != 0 {
			t.Errorf("expect result should no data")
		}
		if err == nil {
			t.Fatal("error should not be null")
		}
	})
	// valid -> extract complete and transform to tax request array
	t.Run("when valid csv field should return array of tax csv data", func(t *testing.T) {
		stub := initStub(nil, nil)
		s := setupTaxService(stub)
		fileData := openCsvFile(t, "../testdata/valid-taxes.csv")

		result, err := s.ExtractCsv(fileData)

		assertIsNil(t, err, expectNilErrMsg)
		if len(result) != 3 {
			t.Errorf("expect result length should be 3 but have %d", len(result))
		}
		expect := []models.TaxCsv{
			{
				TotalIncome: 500_000, Wht: 0, Donation: 0, KReceipt: 0,
			},
			{
				TotalIncome: 600_000, Wht: 40_000, Donation: 20_000, KReceipt: 0,
			},
			{
				TotalIncome: 750_000, Wht: 50_000, Donation: 15_000, KReceipt: 0,
			},
		}
		assertObjectIsEqual(t, expect, result)
	})
	// unorder -> extract complete and transform to tax request array
	t.Run("when csv data is unorder field should return array of tax csv data", func(t *testing.T) {
		stub := initStub(nil, nil)
		s := setupTaxService(stub)
		fileData := openCsvFile(t, "../testdata/unorder-column-taxes.csv")

		result, err := s.ExtractCsv(fileData)

		assertIsNil(t, err, expectNilErrMsg)
		if len(result) != 3 {
			t.Errorf("expect result length should be 3 but have %d", len(result))
		}
		expect := []models.TaxCsv{
			{
				TotalIncome: 500_000, Wht: 0, Donation: 0, KReceipt: 0,
			},
			{
				TotalIncome: 600_000, Wht: 40_000, Donation: 20_000, KReceipt: 0,
			},
			{
				TotalIncome: 750_000, Wht: 50_000, Donation: 15_000, KReceipt: 0,
			},
		}
		assertObjectIsEqual(t, expect, result)
	})
	// over column -> extract complete and transform to tax request array
	t.Run("when csv field have more than expected should return array of tax csv data", func(t *testing.T) {
		stub := initStub(nil, nil)
		s := setupTaxService(stub)
		fileData := openCsvFile(t, "../testdata/over-column-taxes.csv")

		result, err := s.ExtractCsv(fileData)

		t.Log(result, err)
		assertIsNil(t, err, expectNilErrMsg)
		if len(result) != 3 {
			t.Errorf("expect result length should be 3 but have %d", len(result))
		}
		expect := []models.TaxCsv{
			{
				TotalIncome: 500_000, Wht: 0, Donation: 0, KReceipt: 0,
			},
			{
				TotalIncome: 600_000, Wht: 40_000, Donation: 20_000, KReceipt: 10_000,
			},
			{
				TotalIncome: 750_000, Wht: 50_000, Donation: 15_000, KReceipt: 10_000,
			},
		}
		assertObjectIsEqual(t, expect, result)
	})
}

func TestTransformCsvToRequest(t *testing.T) {
	t.Run("given tax csv model should return tax request model", func(t *testing.T) {
		var csv models.TaxCsv
		var expect models.TaxRequest

		t.Run("when csv have only total income", func(t *testing.T) {
			csv.TotalIncome = 500_000

			output := TransformTaxCsvToTaxRequest(csv)

			expect.TotalIncome = 500_000
			expect.Allowances = []models.Allowance{
				{Type: models.DonationSlug, Amount: 0},
				{Type: models.KReceiptSlug, Amount: 0},
			}
			assertObjectIsEqual(t, expect, output)
		})
		t.Run("when csv have total income and wht", func(t *testing.T) {
			csv.TotalIncome = 500_000
			csv.Wht = 25_000

			output := TransformTaxCsvToTaxRequest(csv)

			expect.TotalIncome = 500_000
			expect.Wht = 25_000
			expect.Allowances = []models.Allowance{
				{Type: models.DonationSlug, Amount: 0},
				{Type: models.KReceiptSlug, Amount: 0},
			}
			assertObjectIsEqual(t, expect, output)
		})
		t.Run("when csv have total income and donation", func(t *testing.T) {
			csv.TotalIncome = 500_000
			csv.Donation = 20_000

			output := TransformTaxCsvToTaxRequest(csv)

			expect.TotalIncome = 500_000
			expect.Allowances = []models.Allowance{
				{Type: models.DonationSlug, Amount: 20_000},
				{Type: models.KReceiptSlug, Amount: 0},
			}
			assertObjectIsEqual(t, expect, output)
		})
		t.Run("when csv have total income, wht and donation", func(t *testing.T) {
			csv.TotalIncome = 500_000
			csv.Wht = 25_000
			csv.Donation = 20_000

			output := TransformTaxCsvToTaxRequest(csv)

			expect.TotalIncome = 500_000
			expect.Wht = 25_000
			expect.Allowances = []models.Allowance{
				{Type: models.DonationSlug, Amount: 20_000},
				{Type: models.KReceiptSlug, Amount: 0},
			}
			assertObjectIsEqual(t, expect, output)
		})
		t.Run("when csv have all income, wht, donation and k-receipt", func(t *testing.T) {
			csv.TotalIncome = 500_000
			csv.Wht = 25_000
			csv.Donation = 20_000
			csv.KReceipt = 10_000

			output := TransformTaxCsvToTaxRequest(csv)

			expect.TotalIncome = 500_000
			expect.Wht = 25_000
			expect.Allowances = []models.Allowance{
				{Type: models.DonationSlug, Amount: 20_000},
				{Type: models.KReceiptSlug, Amount: 10_000},
			}
			assertObjectIsEqual(t, expect, output)
		})
	})
}

func TestFuncCalculateTaxCsv(t *testing.T) {
	t.Run("given error on get deduction config should return error", func(t *testing.T) {
		stub := StubTaxStore{
			err:             errors.New("error 'xxx' occured"),
			expectToCall:    map[string]bool{},
			expectCallTimes: map[string]int{},
		}
		s := NewTaxService(&stub)
		input := []models.TaxCsv{
			{TotalIncome: 500_000, Wht: 0, Donation: 0},
			{TotalIncome: 600_000, Wht: 40_000, Donation: 20_000},
			{TotalIncome: 750_000, Wht: 50_000, Donation: 15_000},
		}

		_, err := s.CalculateTaxCsv(input)

		stub.assertMethodWasCalled(t, "GetDeductions")
		stub.assertMethodCalledTime(t, "GetDeductions", 1)
		if err == nil {
			t.Fatalf("expect error should not null")
		}
		assertIsEqual(t, stub.err.Error(), err.Error(), fmt.Sprintf("expect error %s but got %s", stub.err.Error(), err.Error()))
	})
	t.Run("given list of csv data should return list of csv response", func(t *testing.T) {
		stub := StubTaxStore{
			deductions: []models.Deduction{
				{Slug: models.DonationSlug, Amount: 100_000},
				{Slug: models.PersonalSlug, Amount: 60_000},
				{Slug: models.KReceiptSlug, Amount: 50_000},
			},
			expectToCall:    map[string]bool{},
			expectCallTimes: map[string]int{},
		}
		s := NewTaxService(&stub)
		input := []models.TaxCsv{
			{TotalIncome: 500_000, Wht: 0, Donation: 0},
			{TotalIncome: 600_000, Wht: 40_000, Donation: 20_000},
			{TotalIncome: 750_000, Wht: 50_000, Donation: 15_000},
		}

		result, err := s.CalculateTaxCsv(input)

		assertIsNil(t, err, expectNilErrMsg)
		stub.assertMethodWasCalled(t, "GetDeductions")
		stub.assertMethodCalledTime(t, "GetDeductions", 1)
		expect := models.TaxCsvResponse{
			Taxes: []models.CsvCalculateResult{
				{
					TotalIncome: 500_000,
					Tax:         29_000,
				},
				{
					TotalIncome: 600_000,
					TaxRefund:   2_000,
				},
				{
					TotalIncome: 750_000,
					Tax:         11_250,
				},
			},
		}
		assertObjectIsEqual(t, expect, result)
	})
	t.Run("given list of csv data with k-receipt should return list of csv response", func(t *testing.T) {
		stub := StubTaxStore{
			deductions: []models.Deduction{
				{Slug: models.DonationSlug, Amount: 100_000},
				{Slug: models.PersonalSlug, Amount: 60_000},
				{Slug: models.KReceiptSlug, Amount: 50_000},
			},
			expectToCall:    map[string]bool{},
			expectCallTimes: map[string]int{},
		}
		s := NewTaxService(&stub)
		input := []models.TaxCsv{
			{TotalIncome: 500_000, Wht: 0, Donation: 0, KReceipt: 0},
			{TotalIncome: 600_000, Wht: 40_000, Donation: 20_000, KReceipt: 10_000},
			{TotalIncome: 750_000, Wht: 50_000, Donation: 15_000, KReceipt: 10_000},
			{TotalIncome: 500_000, Wht: 0, Donation: 100_000, KReceipt: 200_000},
		}

		result, err := s.CalculateTaxCsv(input)

		assertIsNil(t, err, expectNilErrMsg)
		stub.assertMethodWasCalled(t, "GetDeductions")
		stub.assertMethodCalledTime(t, "GetDeductions", 1)
		expect := models.TaxCsvResponse{
			Taxes: []models.CsvCalculateResult{
				{
					TotalIncome: 500_000,
					Tax:         29_000,
				},
				{
					TotalIncome: 600_000,
					TaxRefund:   3_500,
				},
				{
					TotalIncome: 750_000,
					Tax:         9_750,
				},
				{
					TotalIncome: 500_000,
					Tax:         14_000,
				},
			},
		}
		assertObjectIsEqual(t, expect, result)
	})
}

func TestGetDeductionConfig(t *testing.T) {
	stub := initStub(nil, nil)
	s := NewTaxService(&stub)
	t.Run("given get no row error from database should return default value of each deduction", func(t *testing.T) {
		stub.err = sql.ErrNoRows
		stub.deductions = nil

		personal, donation, kReceipt, err := s.GetDeductionConfig()

		assertIsNil(t, err, expectNilErrMsg)
		assertIsEqual(t, personal.Amount, DefaultPersonalDeduction, fmt.Sprintf("expect personal deduction is %.2f but got %.2f", DefaultPersonalDeduction, personal.Amount))
		assertIsEqual(t, donation.Amount, DefaultDonationDeduction, fmt.Sprintf("expect donation deduction is %.2f but got %.2f", DefaultDonationDeduction, donation.Amount))
		assertIsEqual(t, kReceipt.Amount, DefaultKReceiptDeduction, fmt.Sprintf("expect k-receipt deduction is %.2f but got %.2f", DefaultKReceiptDeduction, kReceipt.Amount))
	})
	t.Run("given get error that is not 'no row' should return error", func(t *testing.T) {
		stub.err = errors.New("error 'xxx' occured")
		stub.deductions = nil

		_, _, _, err := s.GetDeductionConfig()

		if err == nil {
			t.Fatal("expect error should not be null")
		}
		assertIsEqual(t, stub.err, err, fmt.Sprintf("expect error %q but got %q", stub.err, err))
	})
	t.Run("given get some deduction from db should return value from db and default for other that no data", func(t *testing.T) {
		stub.err = nil
		var personal, donation, kReceipt models.Deduction
		var err error
		// case 1 have personal
		stub.deductions = []models.Deduction{
			{Slug: models.PersonalSlug, Amount: 100},
		}

		personal, donation, kReceipt, err = s.GetDeductionConfig()

		assertIsNil(t, err, expectNilErrMsg)
		assertIsEqual(t, personal.Amount, 100.0, fmt.Sprintf("expect personal deduction is %.2f but got %.2f", 100.0, personal.Amount))
		assertIsEqual(t, donation.Amount, DefaultDonationDeduction, fmt.Sprintf("expect donation deduction is %.2f but got %.2f", DefaultDonationDeduction, donation.Amount))
		assertIsEqual(t, kReceipt.Amount, DefaultKReceiptDeduction, fmt.Sprintf("expect k-receipt deduction is %.2f but got %.2f", DefaultKReceiptDeduction, kReceipt.Amount))

		// case 2 have donation
		stub.deductions = []models.Deduction{
			{Slug: models.DonationSlug, Amount: 100},
		}

		personal, donation, kReceipt, err = s.GetDeductionConfig()

		assertIsNil(t, err, expectNilErrMsg)
		assertIsEqual(t, personal.Amount, DefaultPersonalDeduction, fmt.Sprintf("expect personal deduction is %.2f but got %.2f", DefaultPersonalDeduction, personal.Amount))
		assertIsEqual(t, donation.Amount, 100.0, fmt.Sprintf("expect donation deduction is %.2f but got %.2f", 100.0, donation.Amount))
		assertIsEqual(t, kReceipt.Amount, DefaultKReceiptDeduction, fmt.Sprintf("expect k-receipt deduction is %.2f but got %.2f", DefaultKReceiptDeduction, kReceipt.Amount))

		// case 3 have k-receipt
		stub.deductions = []models.Deduction{
			{Slug: models.KReceiptSlug, Amount: 100},
		}

		personal, donation, kReceipt, err = s.GetDeductionConfig()

		assertIsNil(t, err, expectNilErrMsg)
		assertIsEqual(t, personal.Amount, DefaultPersonalDeduction, fmt.Sprintf("expect personal deduction is %.2f but got %.2f", DefaultPersonalDeduction, personal.Amount))
		assertIsEqual(t, donation.Amount, DefaultDonationDeduction, fmt.Sprintf("expect donation deduction is %.2f but got %.2f", DefaultDonationDeduction, donation.Amount))
		assertIsEqual(t, kReceipt.Amount, 100.0, fmt.Sprintf("expect k-receipt deduction is %.2f but got %.2f", 100.0, kReceipt.Amount))

		// case 4 have donation and k-receipt
		stub.deductions = []models.Deduction{
			{Slug: models.DonationSlug, Amount: 100},
			{Slug: models.KReceiptSlug, Amount: 200},
		}

		personal, donation, kReceipt, err = s.GetDeductionConfig()

		assertIsNil(t, err, expectNilErrMsg)
		assertIsEqual(t, personal.Amount, DefaultPersonalDeduction, fmt.Sprintf("expect personal deduction is %.2f but got %.2f", DefaultPersonalDeduction, personal.Amount))
		assertIsEqual(t, donation.Amount, 100.0, fmt.Sprintf("expect donation deduction is %.2f but got %.2f", 100.0, donation.Amount))
		assertIsEqual(t, kReceipt.Amount, 200.0, fmt.Sprintf("expect k-receipt deduction is %.2f but got %.2f", 200.0, kReceipt.Amount))

		// case 5 have donation and personal
		stub.deductions = []models.Deduction{
			{Slug: models.PersonalSlug, Amount: 100},
			{Slug: models.DonationSlug, Amount: 200},
		}

		personal, donation, kReceipt, err = s.GetDeductionConfig()

		assertIsNil(t, err, expectNilErrMsg)
		assertIsEqual(t, personal.Amount, 100.0, fmt.Sprintf("expect personal deduction is %.2f but got %.2f", 100.0, personal.Amount))
		assertIsEqual(t, donation.Amount, 200.0, fmt.Sprintf("expect donation deduction is %.2f but got %.2f", 200.0, donation.Amount))
		assertIsEqual(t, kReceipt.Amount, DefaultKReceiptDeduction, fmt.Sprintf("expect k-receipt deduction is %.2f but got %.2f", DefaultKReceiptDeduction, kReceipt.Amount))

		// case 6 have personal and k-receipt
		stub.deductions = []models.Deduction{
			{Slug: models.PersonalSlug, Amount: 100},
			{Slug: models.KReceiptSlug, Amount: 200},
		}

		personal, donation, kReceipt, err = s.GetDeductionConfig()

		assertIsNil(t, err, expectNilErrMsg)
		assertIsEqual(t, personal.Amount, 100.0, fmt.Sprintf("expect personal deduction is %.2f but got %.2f", 100.0, personal.Amount))
		assertIsEqual(t, donation.Amount, DefaultDonationDeduction, fmt.Sprintf("expect donation deduction is %.2f but got %.2f", DefaultDonationDeduction, donation.Amount))
		assertIsEqual(t, kReceipt.Amount, 200.0, fmt.Sprintf("expect k-receipt deduction is %.2f but got %.2f", 200.0, kReceipt.Amount))

		// case 7 have all
		stub.deductions = []models.Deduction{
			{Slug: models.PersonalSlug, Amount: 100},
			{Slug: models.DonationSlug, Amount: 200},
			{Slug: models.KReceiptSlug, Amount: 300},
		}

		personal, donation, kReceipt, err = s.GetDeductionConfig()

		assertIsNil(t, err, expectNilErrMsg)
		assertIsEqual(t, personal.Amount, 100.0, fmt.Sprintf("expect personal deduction is %.2f but got %.2f", 100.0, personal.Amount))
		assertIsEqual(t, donation.Amount, 200.0, fmt.Sprintf("expect donation deduction is %.2f but got %.2f", 200.0, donation.Amount))
		assertIsEqual(t, kReceipt.Amount, 300.0, fmt.Sprintf("expect k-receipt deduction is %.2f but got %.2f", 300.0, kReceipt.Amount))

	})
}
