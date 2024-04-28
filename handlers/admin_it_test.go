//go:build integration
// +build integration

package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/baronight/assessment-tax/models"
)

func TestITPersonalDeduction(t *testing.T) {
	var got models.PersonalResponse

	res := clientITRequest(
		http.MethodPost,
		os.Getenv("API_URL")+"/admin/deductions/personal",
		io.NopCloser(
			strings.NewReader(
				`{
				"amount": 60000.0
			}`),
		),
		"aplication/json",
		"adminTax",
		"admin!",
	)

	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Errorf("expect response body to be valid json but got %s", res.Body.String())
	}
	assertHttpCode(t, http.StatusOK, res.StatusCode)

	var want = models.PersonalResponse{Amount: 60_000}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("expect %#v but got %#v", want, got)
	}
}

func TestITKReceiptDeduction(t *testing.T) {
	var got models.PersonalResponse

	res := clientITRequest(
		http.MethodPost,
		os.Getenv("API_URL")+"/admin/deductions/k-receipt",
		io.NopCloser(
			strings.NewReader(
				`{
				"amount": 70000.0
			}`),
		),
		"aplication/json",
		"adminTax",
		"admin!",
	)

	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Errorf("expect response body to be valid json but got %s", res.Body.String())
	}
	assertHttpCode(t, http.StatusOK, res.StatusCode)

	var want = models.KReceiptResponse{Amount: 70_000}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("expect %#v but got %#v", want, got)
	}
}
