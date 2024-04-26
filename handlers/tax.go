package handlers

import (
	"io"
	"net/http"

	"github.com/baronight/assessment-tax/models"
	"github.com/baronight/assessment-tax/utils"
	"github.com/baronight/assessment-tax/validators"
	"github.com/labstack/echo/v4"
)

type TaxHandlers struct {
	Service TaxServicer
}

type TaxServicer interface {
	TaxCalculate(tax models.TaxRequest) (models.TaxResponse, error)
	ExtractCsv(reader io.Reader) ([]models.TaxCsv, error)
	CalculateTaxCsv(taxes []models.TaxCsv) (models.TaxCsvResponse, error)
}

func NewTaxHandlers(service TaxServicer) *TaxHandlers {
	return &TaxHandlers{Service: service}
}

// TaxCalculateHandler
//
// @Summary Tax Calculate API
// @Description To calculate personal tax and return how much addition pay tax / refund tax
// @Tags tax
// @Accept json
// @Produce json
// @Param tax body TaxRequest true "tax data that want to calculate"
// @Success 200 {object} TaxResponse
// @Router /tax/calculations [post]
// @Failure 400 {object} ErrorResponse "validate error or cannot get body"
// @Failure 500 {object} ErrorResponse "internal server error"
func (h *TaxHandlers) TaxCalculateHandler(c echo.Context) error {
	body := new(models.TaxRequest)
	if err := c.Bind(body); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
	}

	if err := validators.ValidateTaxRequest(*body); err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
	}

	result, err := h.Service.TaxCalculate(*body)

	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: utils.ErrInternalServer.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// TaxUploadCsvHandler
//
// @Summary Tax Calculate From CSV file API
// @Description To calculate personal tax from csv file and return list of total income, tax and tax refund of each row data
// @Tags tax
// @Accept mpfd
// @Produce json
// @Param taxFile formData file true "csv tax file"
// @Success 200 {object} TaxCsvResponse
// @Router /tax/calculations/upload-csv [post]
// @Failure 400 {object} ErrorResponse "validate error or cannot get file"
// @Failure 500 {object} ErrorResponse "internal server error"
func (h *TaxHandlers) TaxUploadCsvHandler(c echo.Context) error {
	file, err := c.FormFile("taxFile")
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
	}
	if fileType := file.Header.Get("Content-Type"); fileType != "text/csv" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "support only csv file"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
	}
	defer src.Close()

	csv, err := h.Service.ExtractCsv(src)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
	}

	result, err := h.Service.CalculateTaxCsv(csv)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: utils.ErrInternalServer.Error()})
	}
	return c.JSON(http.StatusOK, result)
}
