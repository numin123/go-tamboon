package client

const (
	maxRetries            = 5
	maxDonationGoroutines = 10

	defaultTokenURL  = "https://vault.omise.co/tokens"
	defaultChargeURL = "https://api.omise.co/charges"
	currency         = "THB"
	returnURI        = "http://www.example.com/orders/complete"
)
