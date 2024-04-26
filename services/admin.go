package services

import (
	"errors"

	"github.com/baronight/assessment-tax/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	ErrDeductionInvalid = errors.New("invalid deduction")
)

type AdminService struct {
	Db AdminStorer
}

type AdminStorer interface {
	GetDeduction(slug string) (models.Deduction, error)
	UpdateDeduction(slug string, amount float64) (models.Deduction, error)
}

func NewAdminService(db AdminStorer) *AdminService {
	return &AdminService{
		Db: db,
	}
}

func (as *AdminService) UpdateDeductionConfig(slug string, amount models.DeductionRequest) (response models.Deduction, err error) {
	err = as.ValidateDeductionRequest(slug, amount.Amount)
	if err != nil {
		return
	}

	response, err = as.Db.UpdateDeduction(slug, amount.Amount)
	return
}

func (as *AdminService) ValidateDeductionRequest(slug string, amount float64) error {
	printer := message.NewPrinter(language.English)
	deduction, err := as.Db.GetDeduction(slug)
	if err != nil {
		return ErrDeductionInvalid
	}

	if amount < deduction.MinAmount {
		err = errors.New(printer.Sprintf("amount should not be less than %.2f", deduction.MinAmount))
		return err
	}
	if deduction.MaxAmount > 0 && amount > deduction.MaxAmount {
		err = errors.New(printer.Sprintf("amount should not be more than %.2f", deduction.MaxAmount))
		return err
	}
	return nil
}
