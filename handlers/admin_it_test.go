//go:build integration
// +build integration

package handlers

import (
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
		strings.NewReader(
			`{
			"amount": 60000.0
		}`),
		"aplication/json",
		"adminTax",
		"admin!",
	)

	err := res.Decode(&got)
	if err != nil {
		t.Errorf("expect response body to be valid json but got %q", err)
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
		strings.NewReader(
			`{
			"amount": 70000.0
		}`),
		"aplication/json",
		"adminTax",
		"admin!",
	)

	err := res.Decode(&got)
	if err != nil {
		t.Errorf("expect response body to be valid json but got %q", err)
	}
	assertHttpCode(t, http.StatusOK, res.StatusCode)

	var want = models.KReceiptResponse{Amount: 70_000}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("expect %#v but got %#v", want, got)
	}
}
