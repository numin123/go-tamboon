package client

import "sync"

type DonationRecord struct {
	Name           string
	AmountSubunits string
	CCNumber       string
	CVV            string
	ExpMonth       string
	ExpYear        string
}

type OmiseClient struct {
	tokenService  *TokenService
	chargeService *ChargeService
	rateLimiter   *RateLimiter
}

type donationStats struct {
	mu            sync.Mutex
	totalCount    int
	totalAmount   int64
	successCount  int
	successAmount int64
	donorAmounts  map[string]int64
}
