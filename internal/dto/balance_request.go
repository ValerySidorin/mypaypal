package dto

type BalanceRequest struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Amount int    `json:"amount"`
}
