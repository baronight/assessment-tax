package services

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"math"
	"slices"
	"strconv"

	"github.com/baronight/assessment-tax/models"
	"github.com/baronight/assessment-tax/validators"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type TaxInput struct {
	tax      models.TaxRequest
	personal models.Deduction
	donation models.Deduction
	kReceipt models.Deduction
}

type TaxService struct {
	Db TaxStorer
}

type TaxStorer interface {
	GetDeductions() ([]models.Deduction, error)
}

var (
	DefaultPersonalDeduction float64 = 60_000
	DefaultDonationDeduction float64 = 100_000
	DefaultKReceiptDeduction float64 = 50_000
)

var TaxStep []models.TaxStep = []models.TaxStep{
	{MinIncome: -1, MaxIncome: 150_000, Rate: 0},
	{MinIncome: 150_000, MaxIncome: 500_000, Rate: 0.1},
	{MinIncome: 500_000, MaxIncome: 1_000_000, Rate: 0.15},
	{MinIncome: 1_000_000, MaxIncome: 2_000_000, Rate: 0.2},
	{MinIncome: 2_000_000, MaxIncome: 0, Rate: 0.35},
}

func NewTaxService(db TaxStorer) *TaxService {
	return &TaxService{
		Db: db,
	}
}

func (ts *TaxService) GetDeductionConfig() (personal, donation, kReceipt models.Deduction, err error) {
	var deductions map[string]models.Deduction = map[string]models.Deduction{}
	ds, err := ts.Db.GetDeductions()
	if err != nil && err != sql.ErrNoRows {
		return personal, donation, kReceipt, err
	}

	for _, v := range ds {
		deductions[v.Slug] = v
		switch v.Slug {
		case models.DonationSlug:
			donation = v
		case models.PersonalSlug:
			personal = v
		case models.KReceiptSlug:
			kReceipt = v
		}
	}

	// no personal data in db
	if personal.Amount == 0 && personal.Slug == "" {
		personal.Amount = DefaultPersonalDeduction
		personal.Slug = models.PersonalSlug
	}
	// no donation data in db
	if donation.Amount == 0 && donation.Slug == "" {
		donation.Amount = DefaultDonationDeduction
		donation.Slug = models.DonationSlug
	}
	// no k-receipt data in db
	if kReceipt.Amount == 0 && kReceipt.Slug == "" {
		kReceipt.Amount = DefaultKReceiptDeduction
		kReceipt.Slug = models.KReceiptSlug
	}

	return personal, donation, kReceipt, nil
}

func CalculateDeductionByType(typeSlug string, allowances []models.Allowance, deduction models.Deduction) (amount float64) {
	for _, allowance := range allowances {
		if allowance.Type == typeSlug {
			amount += allowance.Amount
		}
	}

	if deduction.Amount != 0 && amount > deduction.Amount {
		amount = deduction.Amount
	}
	return
}

func CalculateTaxOutput(input TaxInput) models.TaxResponse {
	tax := input.tax
	personal := input.personal
	donation := input.donation
	kReceipt := input.kReceipt

	netIncome := tax.TotalIncome -
		personal.Amount -
		CalculateDeductionByType(models.DonationSlug, tax.Allowances, donation) -
		CalculateDeductionByType(models.KReceiptSlug, tax.Allowances, kReceipt)
	var result models.TaxResponse
	result.TaxLevel = []models.TaxLevel{}
	for _, v := range TaxStep {
		var taxStep float64
		p := message.NewPrinter(language.English)
		level := p.Sprintf("%.0f-%.0f", v.MinIncome+1, v.MaxIncome)
		overflowStep := netIncome - v.MaxIncome
		if v.MaxIncome <= 0 {
			// that mean unlimit ceiling income
			overflowStep = 0
			level = p.Sprintf("%.0f ขึ้นไป", v.MinIncome+1)
		}
		if overflowStep > 0 {
			// calculate full tax rate on this step
			taxStep = (v.MaxIncome - v.MinIncome) * v.Rate
		} else {
			// calculate remain tax
			remain := netIncome - v.MinIncome
			if remain < 0 {
				remain = 0
			}
			taxStep = remain * v.Rate
		}

		result.Tax += taxStep
		result.TaxLevel = append(result.TaxLevel, models.TaxLevel{Level: level, Tax: taxStep})
	}

	if tax.Wht > result.Tax {
		// over payment tax should refund
		result.TaxRefund = math.Round((tax.Wht-result.Tax)*100) / 100
		result.Tax = 0
	} else {
		result.Tax = math.Round((result.Tax-tax.Wht)*100) / 100
	}

	return result
}

func (ts *TaxService) TaxCalculate(tax models.TaxRequest) (models.TaxResponse, error) {
	personal, donation, kReceipt, err := ts.GetDeductionConfig()
	if err != nil {
		return models.TaxResponse{}, err
	}

	result := CalculateTaxOutput(TaxInput{
		tax:      tax,
		personal: personal,
		donation: donation,
		kReceipt: kReceipt,
	})
	return result, nil
}

func (ts *TaxService) ExtractCsv(reader io.Reader) ([]models.TaxCsv, error) {
	taxes := []models.TaxCsv{}
	csvReader := csv.NewReader(reader)
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	header := rows[0]
	if !validators.IsAllStringInArray(header, []string{"totalIncome", "wht", "donation"}) {
		return nil, errors.New("missing required header field")
	}
	for _, row := range rows[1:] {
		var tax models.TaxCsv
		for idx, col := range row {
			if !slices.Contains([]string{"totalIncome", "wht", "donation", "k-receipt"}, header[idx]) {
				continue
			}
			if col == "" {
				return nil, errors.New("value should not be empty")
			}
			val, err := strconv.ParseFloat(col, 64)
			if err != nil {
				return nil, err
			}
			switch header[idx] {
			case "totalIncome":
				tax.TotalIncome = val
			case "wht":
				tax.Wht = val
			case "donation":
				tax.Donation = val
			case "k-receipt":
				tax.KReceipt = val
			}
		}
		// validate each row data when it is all number value
		if err := validators.ValidateTaxCsv(tax); err != nil {
			return nil, err
		}
		taxes = append(taxes, tax)
	}
	return taxes, nil
}

func TransformTaxCsvToTaxRequest(csv models.TaxCsv) (request models.TaxRequest) {
	request.Allowances = []models.Allowance{}
	request.TotalIncome = csv.TotalIncome
	request.Wht = csv.Wht
	request.Allowances = append(request.Allowances, models.Allowance{
		Type:   models.DonationSlug,
		Amount: csv.Donation,
	})
	request.Allowances = append(request.Allowances, models.Allowance{
		Type:   models.KReceiptSlug,
		Amount: csv.KReceipt,
	})
	return
}

func (ts *TaxService) CalculateTaxCsv(taxes []models.TaxCsv) (models.TaxCsvResponse, error) {
	var result models.TaxCsvResponse = models.TaxCsvResponse{
		Taxes: []models.CsvCalculateResult{},
	}

	personal, donation, kReceipt, err := ts.GetDeductionConfig()
	if err != nil {
		return result, err
	}

	for _, tax := range taxes {
		taxOutput := CalculateTaxOutput(TaxInput{
			personal: personal,
			donation: donation,
			kReceipt: kReceipt,
			tax:      TransformTaxCsvToTaxRequest(tax),
		})
		result.Taxes = append(result.Taxes, models.CsvCalculateResult{
			TotalIncome: tax.TotalIncome,
			Tax:         taxOutput.Tax,
			TaxRefund:   taxOutput.TaxRefund,
		})
	}
	return result, nil
}
