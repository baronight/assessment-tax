package handlers

import (
	"database/sql"
	"net/http"

	"github.com/baronight/assessment-tax/models"
	"github.com/baronight/assessment-tax/utils"
	"github.com/labstack/echo/v4"
)

type AdminHandlers struct {
	Service AdminServicer
}

type AdminServicer interface {
	ValidateDeductionRequest(slug string, amount float32) error
	UpdateDeductionConfig(deduction models.DeductionRequest) (models.Deduction, error)
}

func NewAdminHandlers(service AdminServicer) *AdminHandlers {
	return &AdminHandlers{Service: service}
}

// PersonalDeductionConfigHandler
//
// @Summary Personal Deduction Config API
// @Description To setting personal deduction amount for use in tax calculate
// @Tags admin, deduction
// @Accept json
// @Produce json
// @securityDefinitions.basic BasicAuth
// @Param tax body DeductionRequest true "new amount that you want to set"
// @Success 200 {object} PersonalResponse
// @Router /admin/deductions/personal [post]
// @Failure 404 {object} ErrorResponse "data not found"
// @Failure 400 {object} ErrorResponse "validate error or cannot get body"
// @Failure 500 {object} ErrorResponse "internal server error"
func (h *AdminHandlers) PersonalDeductionConfigHandler(c echo.Context) error {
	body := new(models.DeductionRequest)
	if err := c.Bind(body); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
	}

	if err := h.Service.ValidateDeductionRequest(models.PersonalSlug, body.Amount); err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
	}

	deduction, err := h.Service.UpdateDeductionConfig(*body)

	if err != nil {
		c.Logger().Error(err)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "data not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: utils.ErrInternalServer.Error()})
	}
	result := models.PersonalResponse{
		Amount: deduction.Amount,
	}
	return c.JSON(http.StatusOK, result)
}
