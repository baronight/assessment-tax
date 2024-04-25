package models

type DeductionRequest struct {
	Amount float32 `json:"amount"`
} //@Name DeductionRequest

type PersonalResponse struct {
	Amount float32 `json:"personalDeduction"`
} //@Name PersonalResponse
