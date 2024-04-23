package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/baronight/assessment-tax/models"
	"github.com/labstack/echo/v4"
)

type stubTaxCalculate struct {
	expectToCall    map[string]bool
	expectCallTimes map[string]int
	err             error
	response        models.TaxResponse
}

func (s *stubTaxCalculate) TaxCalculate(tax models.TaxRequest) (models.TaxResponse, error) {
	s.expectToCall["TaxCalculate"] = true
	s.expectCallTimes["TaxCalculate"]++
	return s.response, s.err
}

func assertHttpCode(t *testing.T, expect int, got int) {
	t.Helper()
	if expect != got {
		t.Errorf("expect status code %d but got %d", expect, got)
	}
}
func assertErrorMessage(t *testing.T, expect, got string) {
	t.Helper()
	if expect != got {
		t.Errorf("expect error message '%s' but got '%s'", expect, got)
	}
}
func (s *stubTaxCalculate) assertMethodWasCalled(t *testing.T, methodName string) {
	t.Helper()
	if !s.expectToCall[methodName] {
		t.Errorf("expect %s was called", methodName)
	}
}
func (s *stubTaxCalculate) assertMethodWasNotCalled(t *testing.T, methodName string) {
	t.Helper()
	if s.expectToCall[methodName] {
		t.Errorf("expect %s was not called", methodName)
	}
}
func (s *stubTaxCalculate) assertMethodCalledTime(t *testing.T, methodName string, times int) {
	t.Helper()
	if s.expectCallTimes[methodName] != times {
		t.Errorf("expect %s was called %d times but got %d", methodName, times, s.expectCallTimes[methodName])
	}
}

func decodeTaxResponse(t *testing.T, res *httptest.ResponseRecorder) (got models.TaxResponse) {
	t.Helper()
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Errorf("expect response body to be valid json but got %s", res.Body.String())
	}
	return
}
func decodeErrorResponse(t *testing.T, res *httptest.ResponseRecorder) (got models.ErrorResponse) {
	t.Helper()
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Errorf("expect response body to be valid json but got %s", res.Body.String())
	}
	return
}

func setup(method, url string, body io.Reader) (res *httptest.ResponseRecorder, c echo.Context, h *TaxHandlers, stub *stubTaxCalculate) {
	e := echo.New()
	// e.Validator = &models.CustomValidator{Validator: validator.New()}
	req := httptest.NewRequest(method, url, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)
	stub = &stubTaxCalculate{
		expectToCall:    make(map[string]bool),
		expectCallTimes: make(map[string]int),
	}
	h = NewTaxHandlers(stub)
	return
}
func TestTaxCalculateHandler(t *testing.T) {
	t.Run("given not valid request body should return status 400 with validate message", func(t *testing.T) {
		body, _ := json.Marshal(models.TaxRequest{TotalIncome: -1})
		res, c, h, stub := setup(http.MethodPost, "/tax/calculations", strings.NewReader(string(body)))

		h.TaxCalculateHandler(c)

		stub.assertMethodWasNotCalled(t, "TaxCalculate")
		assertHttpCode(t, http.StatusBadRequest, res.Code)
		got := decodeErrorResponse(t, res)
		if got.Message == "" {
			t.Errorf("expect error message should not be empty")
		}
	})

	t.Run("given valid total income should return status 200 with tax value", func(t *testing.T) {
		body, _ := json.Marshal(models.TaxRequest{
			TotalIncome: 500000.0,
			Wht:         0.0,
			Allowances: []models.Allowance{
				{
					Type:   "donation",
					Amount: 0.0,
				},
			},
		})
		res, c, h, stub := setup(http.MethodPost, "/tax/calculations", strings.NewReader(string(body)))
		stub.response = models.TaxResponse{
			Tax: 29000.0,
		}

		h.TaxCalculateHandler(c)

		stub.assertMethodWasCalled(t, "TaxCalculate")
		stub.assertMethodCalledTime(t, "TaxCalculate", 1)
		assertHttpCode(t, http.StatusOK, res.Code)
		got := decodeTaxResponse(t, res)
		if got.Tax != stub.response.Tax {
			t.Errorf("expect tax should be %.2f but got %.2f", stub.response.Tax, got.Tax)
		}
	})

	t.Run("given error from service should return 500 with error message", func(t *testing.T) {
		body, _ := json.Marshal(models.TaxRequest{
			TotalIncome: 500000.0,
			Wht:         0.0,
			Allowances: []models.Allowance{
				{
					Type:   "donation",
					Amount: 0.0,
				},
			},
		})
		res, c, h, stub := setup(http.MethodPost, "/tax/calculations", strings.NewReader(string(body)))
		stub.err = ErrInternalServer

		h.TaxCalculateHandler(c)

		stub.assertMethodWasCalled(t, "TaxCalculate")
		stub.assertMethodCalledTime(t, "TaxCalculate", 1)
		assertHttpCode(t, http.StatusInternalServerError, res.Code)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, ErrInternalServer.Error(), got.Message)
	})
}
