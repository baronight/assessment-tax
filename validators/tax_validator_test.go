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

	t.Run("given total income is equal 0 should get null error", func(t *testing.T) {
		err := ValidateTotalIncome(0)

		assertIsNil(t, err)
	})

	t.Run("given total income more than 0 should get null error", func(t *testing.T) {
		err := ValidateTotalIncome(500000)

		assertIsNil(t, err)
	})
}
