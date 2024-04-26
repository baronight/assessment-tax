package models

type DeductionRequest struct {
	Amount float64 `json:"amount"`
} //@Name DeductionRequest

type PersonalResponse struct {
	Amount float64 `json:"personalDeduction"`
} //@Name PersonalResponse

type KReceiptResponse struct {
	Amount float64 `json:"kReceipt"`
} //@Name kReceiptResponse
