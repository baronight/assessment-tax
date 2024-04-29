//go:build !integration
// +build !integration

package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/baronight/assessment-tax/middlewares"
	"github.com/baronight/assessment-tax/models"
	"github.com/baronight/assessment-tax/utils"
	"github.com/labstack/echo/v4"
)

type AdminRequestConfig struct {
	method string
	url    string
	user   string
	pass   string
	body   io.Reader
}

type StubAdminServicer struct {
	expectToCall    map[string]bool
	expectCallTimes map[string]int
	err             error
	deduction       models.Deduction
	errValidate     error
}

func (s *StubAdminServicer) ValidateDeductionRequest(slug string, amount float64) error {
	s.expectToCall["ValidateDeductionRequest"] = true
	s.expectCallTimes["ValidateDeductionRequest"]++
	return s.errValidate
}
func (s *StubAdminServicer) UpdateDeductionConfig(slug string, deduction models.DeductionRequest) (models.Deduction, error) {
	s.expectToCall["UpdateDeductionConfig"] = true
	s.expectCallTimes["UpdateDeductionConfig"]++
	return s.deduction, s.err
}

func (s *StubAdminServicer) assertMethodWasCalled(t *testing.T, methodName string) {
	t.Helper()
	if !s.expectToCall[methodName] {
		t.Errorf("expect %s was called", methodName)
	}
}
func (s *StubAdminServicer) assertMethodWasNotCalled(t *testing.T, methodName string) {
	t.Helper()
	if s.expectToCall[methodName] {
		t.Errorf("expect %s was not called", methodName)
	}
}
func (s *StubAdminServicer) assertMethodCalledTime(t *testing.T, methodName string, times int) {
	t.Helper()
	if s.expectCallTimes[methodName] != times {
		t.Errorf("expect %s was called %d times but got %d", methodName, times, s.expectCallTimes[methodName])
	}
}

func setupAdminHandler(config AdminRequestConfig) (res *httptest.ResponseRecorder, c echo.Context, h *AdminHandlers, stub *StubAdminServicer, mw echo.MiddlewareFunc) {
	e := echo.New()
	// e.Validator = &models.CustomValidator{Validator: validator.New()}
	req := httptest.NewRequest(config.method, config.url, config.body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth(config.user, config.pass)

	res = httptest.NewRecorder()
	mw = middlewares.BasicAuthMiddleware()
	e.Use(mw)
	c = e.NewContext(req, res)
	stub = &StubAdminServicer{
		expectToCall:    make(map[string]bool),
		expectCallTimes: make(map[string]int),
	}
	h = NewAdminHandlers(stub)
	return
}

func TestPersonalDeductionConfigHandler(t *testing.T) {
	os.Setenv("ADMIN_USERNAME", "adminTax")
	os.Setenv("ADMIN_PASSWORD", "admin!")
	url := "/admin/deductions/personal"
	t.Run("given invalid authentication should return status 401", func(t *testing.T) {
		body, _ := json.Marshal(models.DeductionRequest{Amount: 60_000})
		res, c, h, stub, mw := setupAdminHandler(
			AdminRequestConfig{
				method: http.MethodPost,
				url:    url,
				user:   "hello",
				pass:   "world",
				body:   strings.NewReader(string(body)),
			},
		)

		err := mw(func(c echo.Context) error {
			return h.PersonalDeductionConfigHandler(c)
		})(c)

		var statusCode int
		statusCode = res.Code
		if err != nil {
			statusCode = err.(*echo.HTTPError).Code
		}
		stub.assertMethodWasNotCalled(t, "ValidateDeductionRequest")
		stub.assertMethodWasNotCalled(t, "UpdateDeductionConfig")
		assertHttpCode(t, http.StatusUnauthorized, statusCode)
	})
	t.Run("given amount is invalid should return 400 with error message from validate function", func(t *testing.T) {
		body, _ := json.Marshal(models.DeductionRequest{Amount: 9_000})
		res, c, h, stub, mw := setupAdminHandler(
			AdminRequestConfig{
				method: http.MethodPost,
				url:    url,
				user:   "adminTax",
				pass:   "admin!",
				body:   strings.NewReader(string(body)),
			},
		)
		stub.errValidate = errors.New("error 'xxx' occured")

		err := mw(func(c echo.Context) error {
			return h.PersonalDeductionConfigHandler(c)
		})(c)

		var statusCode int
		statusCode = res.Code
		if err != nil {
			statusCode = err.(*echo.HTTPError).Code
		}
		stub.assertMethodWasCalled(t, "ValidateDeductionRequest")
		stub.assertMethodCalledTime(t, "ValidateDeductionRequest", 1)
		stub.assertMethodWasNotCalled(t, "UpdateDeductionConfig")
		assertHttpCode(t, http.StatusBadRequest, statusCode)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, stub.errValidate.Error(), got.Message)
	})
	t.Run("given no found personal data in database should return 404 with message 'data not found'", func(t *testing.T) {
		body, _ := json.Marshal(models.DeductionRequest{Amount: 60_000})
		res, c, h, stub, mw := setupAdminHandler(
			AdminRequestConfig{
				method: http.MethodPost,
				url:    url,
				user:   "adminTax",
				pass:   "admin!",
				body:   strings.NewReader(string(body)),
			},
		)
		stub.err = sql.ErrNoRows

		err := mw(func(c echo.Context) error {
			return h.PersonalDeductionConfigHandler(c)
		})(c)

		var statusCode int
		statusCode = res.Code
		if err != nil {
			statusCode = err.(*echo.HTTPError).Code
		}
		stub.assertMethodWasCalled(t, "ValidateDeductionRequest")
		stub.assertMethodWasCalled(t, "UpdateDeductionConfig")
		stub.assertMethodCalledTime(t, "UpdateDeductionConfig", 1)
		assertHttpCode(t, http.StatusNotFound, statusCode)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, "data not found", got.Message)
	})
	t.Run("given error on call 'UpdateDeductionConfig' should return status 500 with message 'internal server error'", func(t *testing.T) {
		body, _ := json.Marshal(models.DeductionRequest{Amount: 60_000})
		res, c, h, stub, mw := setupAdminHandler(
			AdminRequestConfig{
				method: http.MethodPost,
				url:    url,
				user:   "adminTax",
				pass:   "admin!",
				body:   strings.NewReader(string(body)),
			},
		)
		stub.err = errors.New("error 'xxx' occured")

		err := mw(func(c echo.Context) error {
			return h.PersonalDeductionConfigHandler(c)
		})(c)

		var statusCode int
		statusCode = res.Code
		if err != nil {
			statusCode = err.(*echo.HTTPError).Code
		}
		stub.assertMethodWasCalled(t, "ValidateDeductionRequest")
		stub.assertMethodWasCalled(t, "UpdateDeductionConfig")
		stub.assertMethodCalledTime(t, "UpdateDeductionConfig", 1)
		assertHttpCode(t, http.StatusInternalServerError, statusCode)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, utils.ErrInternalServer.Error(), got.Message)
	})
	t.Run("given valid amount should return 200 with updated personal deduction amount", func(t *testing.T) {
		body, _ := json.Marshal(models.DeductionRequest{Amount: 60_000})
		res, c, h, stub, mw := setupAdminHandler(
			AdminRequestConfig{
				method: http.MethodPost,
				url:    url,
				user:   "adminTax",
				pass:   "admin!",
				body:   strings.NewReader(string(body)),
			},
		)
		stub.deduction = models.Deduction{Amount: 60_000}

		err := mw(func(c echo.Context) error {
			return h.PersonalDeductionConfigHandler(c)
		})(c)

		var statusCode int
		statusCode = res.Code
		if err != nil {
			statusCode = err.(*echo.HTTPError).Code
		}
		stub.assertMethodWasCalled(t, "ValidateDeductionRequest")
		stub.assertMethodWasCalled(t, "UpdateDeductionConfig")
		stub.assertMethodCalledTime(t, "UpdateDeductionConfig", 1)
		assertHttpCode(t, http.StatusOK, statusCode)
		want := models.PersonalResponse{Amount: 60_000}
		var got models.PersonalResponse
		if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
			t.Errorf("expect response body to be valid json but got %s", res.Body.String())
		}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("expect %#v but got %#v", want, got)
		}
	})
}

func TestKReceiptDeductionConfigHandler(t *testing.T) {
	os.Setenv("ADMIN_USERNAME", "adminTax")
	os.Setenv("ADMIN_PASSWORD", "admin!")
	url := "/admin/deductions/k-receipt"
	t.Run("given invalid authentication should return status 401", func(t *testing.T) {
		body, _ := json.Marshal(models.DeductionRequest{Amount: 60_000})
		res, c, h, stub, mw := setupAdminHandler(
			AdminRequestConfig{
				method: http.MethodPost,
				url:    url,
				user:   "hello",
				pass:   "world",
				body:   strings.NewReader(string(body)),
			},
		)

		err := mw(func(c echo.Context) error {
			return h.KReceiptDeductionConfigHandler(c)
		})(c)

		var statusCode int
		statusCode = res.Code
		if err != nil {
			statusCode = err.(*echo.HTTPError).Code
		}
		stub.assertMethodWasNotCalled(t, "ValidateDeductionRequest")
		stub.assertMethodWasNotCalled(t, "UpdateDeductionConfig")
		assertHttpCode(t, http.StatusUnauthorized, statusCode)
	})
	t.Run("given amount is invalid should return 400 with error message from validate function", func(t *testing.T) {
		body, _ := json.Marshal(models.DeductionRequest{Amount: 900_000})
		res, c, h, stub, mw := setupAdminHandler(
			AdminRequestConfig{
				method: http.MethodPost,
				url:    url,
				user:   "adminTax",
				pass:   "admin!",
				body:   strings.NewReader(string(body)),
			},
		)
		stub.errValidate = errors.New("error 'xxx' occured")

		err := mw(func(c echo.Context) error {
			return h.KReceiptDeductionConfigHandler(c)
		})(c)

		var statusCode int
		statusCode = res.Code
		if err != nil {
			statusCode = err.(*echo.HTTPError).Code
		}
		stub.assertMethodWasCalled(t, "ValidateDeductionRequest")
		stub.assertMethodCalledTime(t, "ValidateDeductionRequest", 1)
		stub.assertMethodWasNotCalled(t, "UpdateDeductionConfig")
		assertHttpCode(t, http.StatusBadRequest, statusCode)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, stub.errValidate.Error(), got.Message)
	})
	t.Run("given no found k-receipt data in database should return 404 with message 'data not found'", func(t *testing.T) {
		body, _ := json.Marshal(models.DeductionRequest{Amount: 60_000})
		res, c, h, stub, mw := setupAdminHandler(
			AdminRequestConfig{
				method: http.MethodPost,
				url:    url,
				user:   "adminTax",
				pass:   "admin!",
				body:   strings.NewReader(string(body)),
			},
		)
		stub.err = sql.ErrNoRows

		err := mw(func(c echo.Context) error {
			return h.KReceiptDeductionConfigHandler(c)
		})(c)

		var statusCode int
		statusCode = res.Code
		if err != nil {
			statusCode = err.(*echo.HTTPError).Code
		}
		stub.assertMethodWasCalled(t, "ValidateDeductionRequest")
		stub.assertMethodWasCalled(t, "UpdateDeductionConfig")
		stub.assertMethodCalledTime(t, "UpdateDeductionConfig", 1)
		assertHttpCode(t, http.StatusNotFound, statusCode)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, "data not found", got.Message)
	})
	t.Run("given error on call 'UpdateDeductionConfig' should return status 500 with message 'internal server error'", func(t *testing.T) {
		body, _ := json.Marshal(models.DeductionRequest{Amount: 60_000})
		res, c, h, stub, mw := setupAdminHandler(
			AdminRequestConfig{
				method: http.MethodPost,
				url:    url,
				user:   "adminTax",
				pass:   "admin!",
				body:   strings.NewReader(string(body)),
			},
		)
		stub.err = errors.New("error 'xxx' occured")

		err := mw(func(c echo.Context) error {
			return h.KReceiptDeductionConfigHandler(c)
		})(c)

		var statusCode int
		statusCode = res.Code
		if err != nil {
			statusCode = err.(*echo.HTTPError).Code
		}
		stub.assertMethodWasCalled(t, "ValidateDeductionRequest")
		stub.assertMethodWasCalled(t, "UpdateDeductionConfig")
		stub.assertMethodCalledTime(t, "UpdateDeductionConfig", 1)
		assertHttpCode(t, http.StatusInternalServerError, statusCode)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, utils.ErrInternalServer.Error(), got.Message)
	})
	t.Run("given valid amount should return 200 with updated k-receipt deduction amount", func(t *testing.T) {
		body, _ := json.Marshal(models.DeductionRequest{Amount: 60_000})
		res, c, h, stub, mw := setupAdminHandler(
			AdminRequestConfig{
				method: http.MethodPost,
				url:    url,
				user:   "adminTax",
				pass:   "admin!",
				body:   strings.NewReader(string(body)),
			},
		)
		stub.deduction = models.Deduction{Amount: 60_000}

		err := mw(func(c echo.Context) error {
			return h.KReceiptDeductionConfigHandler(c)
		})(c)

		var statusCode int
		statusCode = res.Code
		if err != nil {
			statusCode = err.(*echo.HTTPError).Code
		}
		stub.assertMethodWasCalled(t, "ValidateDeductionRequest")
		stub.assertMethodWasCalled(t, "UpdateDeductionConfig")
		stub.assertMethodCalledTime(t, "UpdateDeductionConfig", 1)
		assertHttpCode(t, http.StatusOK, statusCode)
		want := models.KReceiptResponse{Amount: 60_000}
		var got models.KReceiptResponse
		if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
			t.Errorf("expect response body to be valid json but got %s", res.Body.String())
		}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("expect %#v but got %#v", want, got)
		}
	})
}
