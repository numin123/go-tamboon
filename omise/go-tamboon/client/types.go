package client

import "sync"

type DonationRecord struct {
	Name           string
	AmountSubunits int
	CCNumber       string
	CVV            string
	ExpMonth       string
	ExpYear        string
}

type TokenResponse struct {
	ID string `json:"id"`
}

type ChargeResponse struct {
	ID     string `json:"id"`
	Amount int    `json:"amount"`
}

type OmiseClient struct {
	tokenService  *TokenService
	chargeService *ChargeService
}

type donationStats struct {
	mu            sync.Mutex
	totalCount    int
	totalAmount   int
	successCount  int
	successAmount int
	donorAmounts  map[string]int
}
