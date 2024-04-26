package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/baronight/assessment-tax/models"
)

type PersonalTestSuite struct {
	name                string
	stub                StubAdminStorer
	want                models.Deduction
	params              models.DeductionRequest
	wantError           error
	updateDeductionCall bool
}

type DbDeductionResult struct {
	deduction models.Deduction
	err       error
}

type StubAdminStorer struct {
	getDeduction       models.Deduction
	updateDeduction    models.Deduction
	getDeductionErr    error
	updateDeductionErr error
	expectToCall       map[string]bool
	expectCallTimes    map[string]int
}

func (s *StubAdminStorer) GetDeduction(slug string) (models.Deduction, error) {
	s.expectToCall["GetDeduction"] = true
	s.expectCallTimes["GetDeduction"]++
	return s.getDeduction, s.getDeductionErr
}

func (s *StubAdminStorer) UpdateDeduction(slug string, amount float64) (models.Deduction, error) {
	s.expectToCall["UpdateDeduction"] = true
	s.expectCallTimes["UpdateDeduction"]++
	return s.updateDeduction, s.updateDeductionErr
}

func (s *StubAdminStorer) assertMethodWasCalled(t *testing.T, methodName string) {
	t.Helper()
	if !s.expectToCall[methodName] {
		t.Errorf("expect %s was called", methodName)
	}
}
func (s *StubAdminStorer) assertMethodWasNotCalled(t *testing.T, methodName string) {
	t.Helper()
	if s.expectToCall[methodName] {
		t.Errorf("expect %s was not called", methodName)
	}
}
func (s *StubAdminStorer) assertMethodCalledTime(t *testing.T, methodName string, times int) {
	t.Helper()
	if s.expectCallTimes[methodName] != times {
		t.Errorf("expect %s was called %d times but got %d", methodName, times, s.expectCallTimes[methodName])
	}
}

func setupAdminService(stub StubAdminStorer) *AdminService {
	service := NewAdminService(&stub)

	return service
}

func initStubAdminStorer(getDeduction, updateDeduction DbDeductionResult) StubAdminStorer {
	return StubAdminStorer{
		expectToCall:       map[string]bool{},
		expectCallTimes:    map[string]int{},
		getDeduction:       getDeduction.deduction,
		getDeductionErr:    getDeduction.err,
		updateDeduction:    updateDeduction.deduction,
		updateDeductionErr: updateDeduction.err,
	}
}

func TestValidateDeduction(t *testing.T) {
	t.Run("given amount is less than acceptable amount should return error 'amount should not be less than xxx'", func(t *testing.T) {
		stub := StubAdminStorer{
			getDeduction:    models.Deduction{MinAmount: 10_000},
			expectToCall:    map[string]bool{},
			expectCallTimes: map[string]int{},
		}
		service := setupAdminService(stub)

		err := service.ValidateDeductionRequest("xxx", 5_000)

		stub.assertMethodWasCalled(t, "GetDeduction")
		stub.assertMethodCalledTime(t, "GetDeduction", 1)
		want := errors.New("amount should not be less than 10,000.00")
		if err == nil {
			t.Fatalf("expect error should not null")
		}
		assertIsEqual(t, want.Error(), err.Error(), fmt.Sprintf("expect error %q but got %q", want.Error(), err.Error()))
	})
	t.Run("given amount is more than acceptable amount should return error 'amount should not be more than xxx'", func(t *testing.T) {
		stub := StubAdminStorer{
			getDeduction:    models.Deduction{MaxAmount: 100_000},
			expectToCall:    map[string]bool{},
			expectCallTimes: map[string]int{},
		}
		service := setupAdminService(stub)

		err := service.ValidateDeductionRequest("xxx", 100_001)

		stub.assertMethodWasCalled(t, "GetDeduction")
		stub.assertMethodCalledTime(t, "GetDeduction", 1)
		want := errors.New("amount should not be more than 100,000.00")
		if err == nil {
			t.Fatalf("expect error should not null")
		}
		assertIsEqual(t, want.Error(), err.Error(), fmt.Sprintf("expect error %q but got %q", want.Error(), err.Error()))
	})
	t.Run("given error on call 'GetDeduction' should return ErrDeductionInvalid", func(t *testing.T) {
		stub := StubAdminStorer{
			getDeductionErr: errors.New("xxx"),
			expectToCall:    map[string]bool{},
			expectCallTimes: map[string]int{},
		}
		service := setupAdminService(stub)

		err := service.ValidateDeductionRequest("xxx", 5_000)

		stub.assertMethodWasCalled(t, "GetDeduction")
		stub.assertMethodCalledTime(t, "GetDeduction", 1)
		want := ErrDeductionInvalid
		if err == nil {
			t.Fatalf("expect error should not null")
		}
		assertIsEqual(t, want.Error(), err.Error(), fmt.Sprintf("expect error %s but got %s", want.Error(), err.Error()))
	})
	t.Run("given amount is in range of acceptable amount should return null error", func(t *testing.T) {
		stub := StubAdminStorer{
			getDeduction:    models.Deduction{MinAmount: 10_000, MaxAmount: 100_000},
			expectToCall:    map[string]bool{},
			expectCallTimes: map[string]int{},
		}
		service := setupAdminService(stub)

		err := service.ValidateDeductionRequest("xxx", 50_000)

		stub.assertMethodWasCalled(t, "GetDeduction")
		stub.assertMethodCalledTime(t, "GetDeduction", 1)

		if err != nil {
			t.Fatalf("expect error should be null but got %q", err)
		}
	})
}

func TestUpdateDeductionConfig(t *testing.T) {
	testSuites := []PersonalTestSuite{
		{
			name: "given error when call \"getDeduction\" should return error",
			stub: initStubAdminStorer(
				DbDeductionResult{
					deduction: models.Deduction{},
					err:       errors.New("error xxx occured"),
				},
				DbDeductionResult{},
			),
			want:                models.Deduction{},
			params:              models.DeductionRequest{},
			wantError:           ErrDeductionInvalid,
			updateDeductionCall: false,
		},
		{
			name: "given amount is less than minimum acceptable amount should return error 'amount should not be less than xxx'",
			stub: initStubAdminStorer(
				DbDeductionResult{
					deduction: models.Deduction{MinAmount: 10_000},
				},
				DbDeductionResult{},
			),
			params:              models.DeductionRequest{Amount: 5_000},
			wantError:           errors.New("amount should not be less than 10,000.00"),
			updateDeductionCall: false,
		},
		{
			name: "given amount is more than maximum acceptable amount should return error 'amount should not be more than xxx'",
			stub: initStubAdminStorer(
				DbDeductionResult{
					deduction: models.Deduction{MaxAmount: 100_000},
				},
				DbDeductionResult{},
			),
			params:              models.DeductionRequest{Amount: 200_000},
			wantError:           errors.New("amount should not be more than 100,000.00"),
			updateDeductionCall: false,
		},
		{
			name: "given error on called 'updateDeduction' should return error",
			stub: initStubAdminStorer(
				DbDeductionResult{
					deduction: models.Deduction{MaxAmount: 100_000},
				},
				DbDeductionResult{
					err: errors.New("error xxx occured"),
				},
			),
			params:              models.DeductionRequest{Amount: 50_000},
			wantError:           errors.New("error xxx occured"),
			updateDeductionCall: true,
		},
		{
			name: "given amount is in range of minimum and maximum acceptable amount should return updated amount",
			stub: initStubAdminStorer(
				DbDeductionResult{
					deduction: models.Deduction{MaxAmount: 100_000},
				},
				DbDeductionResult{
					deduction: models.Deduction{Amount: 50_000},
				},
			),
			params:              models.DeductionRequest{Amount: 50_000},
			want:                models.Deduction{Amount: 50_000},
			updateDeductionCall: true,
		},
	}

	for _, tc := range testSuites {
		t.Run(tc.name, func(t *testing.T) {
			service := setupAdminService(tc.stub)

			result, err := service.UpdateDeductionConfig(tc.params)

			// verify get deduction was called
			tc.stub.assertMethodWasCalled(t, "GetDeduction")
			tc.stub.assertMethodCalledTime(t, "GetDeduction", 1)
			// verify update deduction was called or not
			if tc.updateDeductionCall {
				tc.stub.assertMethodWasCalled(t, "UpdateDeduction")
				tc.stub.assertMethodCalledTime(t, "UpdateDeduction", 1)
			} else {
				tc.stub.assertMethodWasNotCalled(t, "")
			}
			if tc.wantError != nil {
				if err == nil {
					t.Fatalf("expect error should not null")
				}
				assertIsEqual(t, tc.wantError.Error(), err.Error(), fmt.Sprintf("expect error %s but got %s", tc.wantError.Error(), err.Error()))
			} else {
				assertIsNil(t, err, expectNilErrMsg)
				assertObjectIsEqual(t, tc.want, result)
			}
		})
	}
}
