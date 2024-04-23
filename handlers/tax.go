package handlers

import (
	"errors"
	"net/http"

	"github.com/baronight/assessment-tax/models"
	"github.com/baronight/assessment-tax/validators"
	"github.com/labstack/echo/v4"
)

var (
	ErrInternalServer = errors.New("Internal Server Error")
)

type TaxHandlers struct {
	Service TaxServicer
}

type TaxServicer interface {
	TaxCalculate(tax models.TaxRequest) (models.TaxResponse, error)
}

func NewTaxHandlers(service TaxServicer) *TaxHandlers {
	return &TaxHandlers{Service: service}
}

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
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: ErrInternalServer.Error()})
	}

	return c.JSON(http.StatusOK, result)
}
