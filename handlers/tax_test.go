//go:build !integration
// +build !integration

package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/baronight/assessment-tax/models"
	"github.com/baronight/assessment-tax/utils"
	"github.com/baronight/assessment-tax/validators"
	"github.com/labstack/echo/v4"
)

type stubTaxCalculate struct {
	expectToCall    map[string]bool
	expectCallTimes map[string]int
	err             error
	response        models.TaxResponse
	extractResult   []models.TaxCsv
	extractErr      error
	csvResponse     models.TaxCsvResponse
}

func (s *stubTaxCalculate) TaxCalculate(tax models.TaxRequest) (models.TaxResponse, error) {
	s.expectToCall["TaxCalculate"] = true
	s.expectCallTimes["TaxCalculate"]++
	return s.response, s.err
}
func (s *stubTaxCalculate) ExtractCsv(reader io.Reader) ([]models.TaxCsv, error) {
	s.expectToCall["ExtractCsv"] = true
	s.expectCallTimes["ExtractCsv"]++
	return s.extractResult, s.extractErr
}
func (s *stubTaxCalculate) CalculateTaxCsv(taxes []models.TaxCsv) (models.TaxCsvResponse, error) {
	s.expectToCall["CalculateTaxCsv"] = true
	s.expectCallTimes["CalculateTaxCsv"]++
	return s.csvResponse, s.err
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

func setupTaxHandler(method, url string, body io.Reader, contentType string) (res *httptest.ResponseRecorder, c echo.Context, h *TaxHandlers, stub *stubTaxCalculate) {
	e := echo.New()
	// e.Validator = &models.CustomValidator{Validator: validator.New()}
	req := httptest.NewRequest(method, url, body)
	req.Header.Set(echo.HeaderContentType, contentType)
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
		t.Run("when total income is not valid should get error message ErrTotalIncomeInvalid", func(t *testing.T) {
			body, _ := json.Marshal(models.TaxRequest{TotalIncome: -1})
			res, c, h, stub := setupTaxHandler(http.MethodPost, "/tax/calculations", strings.NewReader(string(body)), echo.MIMEApplicationJSON)

			h.TaxCalculateHandler(c)

			stub.assertMethodWasNotCalled(t, "TaxCalculate")
			assertHttpCode(t, http.StatusBadRequest, res.Code)
			got := decodeErrorResponse(t, res)
			assertErrorMessage(t, validators.ErrTotalIncomeInvalid.Error(), got.Message)
		})
		t.Run("when wht is not valid should get error message ErrWhtInvalid", func(t *testing.T) {
			body, _ := json.Marshal(models.TaxRequest{Wht: -1})
			res, c, h, stub := setupTaxHandler(http.MethodPost, "/tax/calculations", strings.NewReader(string(body)), echo.MIMEApplicationJSON)

			h.TaxCalculateHandler(c)

			stub.assertMethodWasNotCalled(t, "TaxCalculate")
			assertHttpCode(t, http.StatusBadRequest, res.Code)
			got := decodeErrorResponse(t, res)
			assertErrorMessage(t, validators.ErrWhtInvalid.Error(), got.Message)
		})
		t.Run("when wht is more than income should get error message ErrWhtMoreThanIncome", func(t *testing.T) {
			body, _ := json.Marshal(models.TaxRequest{TotalIncome: 200_000, Wht: 300_000})
			res, c, h, stub := setupTaxHandler(http.MethodPost, "/tax/calculations", strings.NewReader(string(body)), echo.MIMEApplicationJSON)

			h.TaxCalculateHandler(c)

			stub.assertMethodWasNotCalled(t, "TaxCalculate")
			assertHttpCode(t, http.StatusBadRequest, res.Code)
			got := decodeErrorResponse(t, res)
			assertErrorMessage(t, validators.ErrWhtMoreThanIncome.Error(), got.Message)
		})
	})

	t.Run("given valid total income should return status 200 with tax response", func(t *testing.T) {
		body, _ := json.Marshal(models.TaxRequest{
			TotalIncome: 500000.0,
			Wht:         0.0,
			Allowances: []models.Allowance{
				{
					Type:   models.DonationSlug,
					Amount: 0.0,
				},
			},
		})
		res, c, h, stub := setupTaxHandler(http.MethodPost, "/tax/calculations", strings.NewReader(string(body)), echo.MIMEApplicationJSON)
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
					Type:   models.DonationSlug,
					Amount: 0.0,
				},
			},
		})
		res, c, h, stub := setupTaxHandler(http.MethodPost, "/tax/calculations", strings.NewReader(string(body)), echo.MIMEApplicationJSON)
		stub.err = utils.ErrInternalServer

		h.TaxCalculateHandler(c)

		stub.assertMethodWasCalled(t, "TaxCalculate")
		stub.assertMethodCalledTime(t, "TaxCalculate", 1)
		assertHttpCode(t, http.StatusInternalServerError, res.Code)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, utils.ErrInternalServer.Error(), got.Message)
	})
}

func TestTaxUploadCsvHandler(t *testing.T) {
	csvMimeType := "text/csv"
	uploadUrl := "/tax/calculations/upload-csv"
	createCustomPartFromFile := func(w *multipart.Writer, name, filename, contentType string) (io.Writer, error) {
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, name, filename))
		h.Set("Content-Type", contentType)
		return w.CreatePart(h)
	}
	initBody := func(t *testing.T, filePath, contentType string) (*bytes.Buffer, *multipart.Writer) {
		body := new(bytes.Buffer)
		dir, _ := os.Getwd()
		fileData, err := os.Open(filepath.Join(dir, filePath))
		if err != nil {
			t.Fatal(err)
		}
		defer fileData.Close()

		writer := multipart.NewWriter(body)
		part, err := createCustomPartFromFile(writer, "taxFile", "taxes.csv", contentType)
		if err != nil {
			t.Fatal(err)
		}

		if written, err := io.Copy(part, fileData); err != nil {
			t.Fatal(err)
		} else if written == 0 {
			t.Fatal("no content data")
		}
		return body, writer
	}
	decodeTaxCsvResponse := func(res *httptest.ResponseRecorder) models.TaxCsvResponse {
		var got models.TaxCsvResponse
		if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
			t.Fatalf("expect json body valid but got %q", res.Body.String())
		}
		return got
	}
	t.Run("given correct csv file should return 200 with list of tax data", func(t *testing.T) {
		body, writer := initBody(t, "../testdata/valid-taxes.csv", csvMimeType)
		writer.Close()
		res, c, h, stub := setupTaxHandler(http.MethodPost, uploadUrl, body, writer.FormDataContentType())
		stub.csvResponse = models.TaxCsvResponse{
			Taxes: []models.CsvCalculateResult{
				{
					TotalIncome: 500000,
					Tax:         29000,
				},
				{
					TotalIncome: 600000,
					TaxRefund:   2000,
				},
				{
					TotalIncome: 750000,
					Tax:         11250,
				},
			},
		}

		h.TaxUploadCsvHandler(c)

		stub.assertMethodWasCalled(t, "ExtractCsv")
		stub.assertMethodCalledTime(t, "ExtractCsv", 1)
		stub.assertMethodWasCalled(t, "CalculateTaxCsv")
		stub.assertMethodCalledTime(t, "CalculateTaxCsv", 1)
		assertHttpCode(t, http.StatusOK, res.Code)
		got := decodeTaxCsvResponse(res)
		if !reflect.DeepEqual(stub.csvResponse, got) {
			t.Errorf("expected %#v but got %#v", stub.csvResponse, got)
		}
	})
	t.Run("given missing upload file should return 400 with error message", func(t *testing.T) {
		res, c, h, stub := setupTaxHandler(http.MethodPost, uploadUrl, nil, echo.MIMEMultipartForm)

		h.TaxUploadCsvHandler(c)

		assertHttpCode(t, http.StatusBadRequest, res.Code)
		stub.assertMethodWasNotCalled(t, "ExtractCsv")
		got := decodeErrorResponse(t, res)
		if got.Message == "" {
			t.Errorf("expect error message should not empty")
		}
	})
	t.Run("given upload non csv file should return 400 with error message 'support only csv file'", func(t *testing.T) {
		body, writer := initBody(t, "../testdata/taxes.txt", "text/plain")
		writer.Close()
		res, c, h, stub := setupTaxHandler(http.MethodPost, uploadUrl, body, writer.FormDataContentType())

		h.TaxUploadCsvHandler(c)

		stub.assertMethodWasNotCalled(t, "ExtractCsv")
		assertHttpCode(t, http.StatusBadRequest, res.Code)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, "support only csv file", got.Message)
	})
	t.Run("given error on extract csv should return 400 with error message from extract function", func(t *testing.T) {
		body, writer := initBody(t, "../testdata/missing-value-on-required-field-taxes.csv", csvMimeType)
		writer.Close()
		res, c, h, stub := setupTaxHandler(http.MethodPost, uploadUrl, body, writer.FormDataContentType())
		stub.extractErr = errors.New("error 'xxx' occured")

		h.TaxUploadCsvHandler(c)

		stub.assertMethodWasCalled(t, "ExtractCsv")
		stub.assertMethodCalledTime(t, "ExtractCsv", 1)
		stub.assertMethodWasNotCalled(t, "CalculateTaxCsv")
		assertHttpCode(t, http.StatusBadRequest, res.Code)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, stub.extractErr.Error(), got.Message)
	})
	t.Run("given error on calculate function should return 500 with error 'internal server error'", func(t *testing.T) {
		body, writer := initBody(t, "../testdata/valid-taxes.csv", csvMimeType)
		writer.Close()
		res, c, h, stub := setupTaxHandler(http.MethodPost, uploadUrl, body, writer.FormDataContentType())
		stub.err = errors.New("error 'xxx' occured")

		h.TaxUploadCsvHandler(c)

		stub.assertMethodWasCalled(t, "ExtractCsv")
		stub.assertMethodCalledTime(t, "ExtractCsv", 1)
		stub.assertMethodWasCalled(t, "CalculateTaxCsv")
		stub.assertMethodCalledTime(t, "CalculateTaxCsv", 1)
		assertHttpCode(t, http.StatusInternalServerError, res.Code)
		got := decodeErrorResponse(t, res)
		assertErrorMessage(t, utils.ErrInternalServer.Error(), got.Message)
	})
}
