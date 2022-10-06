package model

import "time"

type AddTransactionRequest struct {
	Payer     string    `json:"payer"`
	Points    int       `json:"points"`
	TimeStamp time.Time `json:"timestamp"`
}

type AddTransactionResponse struct {
	PointBalance int `json:"point_balance"`
}

type SpendPointsRequest struct {
	Points int `json:"points"`
}

type SpendPointsResponse struct {
	PointsSpent []PayerPointRecord `json:"points_spent"`
}

type ListPayerBalancesResponse struct {
	PayerBalances []PayerPointRecord
}

type PayerPointRecord struct {
	Payer  string `json:"payer"`
	Points int    `json:"points"`
}
