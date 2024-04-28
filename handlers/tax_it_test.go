// go:build integration
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

func clientITRequest(method, url string, body io.Reader, contentType, authUser, authPass string) *Response {
	req, _ := http.NewRequest(method, url, body)
	req.SetBasicAuth(authUser, authPass)
	req.Header.Add("Content-Type", contentType)
	client := http.Client{}
	res, err := client.Do(req)
	return &Response{res, err}
}

func (r *Response) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(v)
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
		"aplication/json",
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
	var got []models.CsvCalculateResult
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

	res := clientITRequest(
		http.MethodPost,
		os.Getenv("API_URL")+"/tax/calculations",
		body,
		writer.FormDataContentType(),
		"",
		"",
	)

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
