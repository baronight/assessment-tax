//go:build integration
// +build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/baronight/assessment-tax/models"
)

type Response struct {
	*http.Response
	err error
}

func assertHttpCode(t *testing.T, expect int, got int) {
	t.Helper()
	if expect != got {
		t.Errorf("expect status code %d but got %d", expect, got)
	}
}

func clientITRequest(method, url string, body io.Reader, contentType, authUser, authPass string) *Response {
	req, _ := http.NewRequest(method, url, body)
	req.SetBasicAuth(authUser, authPass)
	req.Header.Add("Content-Type", contentType)
	client := http.Client{}
	res, err := client.Do(req)
	return &Response{res, err}
}

func (r *Response) Decode(v interface{}) error {
	var body []byte
	body, r.err = io.ReadAll(r.Body)
	if r.err != nil {
		return r.err
	}

	r.err = json.Unmarshal(body, v)
	if r.err != nil {
		return r.err
	}
	return nil
}

func TestITTaxCalculations(t *testing.T) {
	var got models.TaxResponse

	res := clientITRequest(
		http.MethodPost,
		os.Getenv("API_URL")+"/tax/calculations",
		strings.NewReader(
			`{
			"totalIncome": 500000.0,
			"wht": 0.0,
			"allowances": [
				{
					"allowanceType": "donation",
					"amount": 0.0
				}
			]
		}`),
		"application/json;charset=UTF-8",
		"",
		"",
	)

	err := res.Decode(&got)
	if err != nil {
		t.Errorf("expect response body to be valid json but got %q", err)
	}
	assertHttpCode(t, http.StatusOK, res.StatusCode)

	var want = models.TaxResponse{
		Tax: 29_000,
		TaxLevel: []models.TaxLevel{
			{
				Level: "0-150,000",
				Tax:   0.0,
			},
			{
				Level: "150,001-500,000",
				Tax:   29000.0,
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

	if !reflect.DeepEqual(want, got) {
		t.Errorf("expect %#v but got %#v", want, got)
	}
}

func TestITCsvCalculate(t *testing.T) {
	var got models.TaxCsvResponse
	body := new(bytes.Buffer)
	dir, _ := os.Getwd()
	fileData, err := os.Open(filepath.Join(dir, "../testdata/valid-taxes.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer fileData.Close()

	writer := multipart.NewWriter(body)
	mimeHeader := make(textproto.MIMEHeader)
	mimeHeader.Set("Content-Disposition", `form-data; name="taxFile"; filename="taxes.csv"`)
	mimeHeader.Set("Content-Type", "text/csv")
	part, err := writer.CreatePart(mimeHeader)
	if err != nil {
		t.Fatal(err)
	}

	if written, err := io.Copy(part, fileData); err != nil {
		t.Fatal(err)
	} else if written == 0 {
		t.Fatal("no content data")
	}
	writer.Close()

	res := clientITRequest(
		http.MethodPost,
		os.Getenv("API_URL")+"/tax/calculations/upload-csv",
		body,
		writer.FormDataContentType(),
		"",
		"",
	)
	// var errResp models.ErrorResponse
	// respBody, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	t.Errorf("sss %+v", err)
	// }
	// t.Logf("%q", string(respBody))

	err = res.Decode(&got)
	if err != nil {
		t.Errorf("expect response body to be valid json but got %q", err)
	}
	assertHttpCode(t, http.StatusOK, res.StatusCode)

	var want = models.TaxCsvResponse{
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

	if !reflect.DeepEqual(want, got) {
		t.Errorf("expect %#v but got %#v", want, got)
	}
}
