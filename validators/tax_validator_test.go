package validators

import (
	"testing"
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

func TestValidateTotalIncome(t *testing.T) {
	t.Run("given total income is less than 0 should get error 'ErrTotalIncomeInvalid'", func(t *testing.T) {
		err := ValidateTotalIncome(-1)

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrTotalIncomeInvalid, err)
	})

	t.Run("given total income is equal 0 should not get error", func(t *testing.T) {
		err := ValidateTotalIncome(0)

		assertIsNil(t, err)
	})

	t.Run("given total income more than 0 should not get error", func(t *testing.T) {
		err := ValidateTotalIncome(500000)

		assertIsNil(t, err)
	})
}

func TestValidateWht(t *testing.T) {
	t.Run("given wht is less than 0 should get error 'ErrWhtInvalid'", func(t *testing.T) {
		err := ValidateWht(-1, 0)

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrWhtInvalid, err)
	})

	t.Run("given wht is more than income should get error 'ErrWhtMoreThanIncome", func(t *testing.T) {
		err := ValidateWht(0.1, 0)

		assertIsNotNil(t, err)
		assertErrorMessage(t, ErrWhtMoreThanIncome, err)
	})

	t.Run("given wht is less than income should not get error", func(t *testing.T) {
		err := ValidateWht(0, 200)

		assertIsNil(t, err)
	})

	t.Run("given wht is equal income should not get error", func(t *testing.T) {
		err := ValidateWht(200, 200)

		assertIsNil(t, err)
	})
}
