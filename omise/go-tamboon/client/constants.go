package client

const (
	defaultMaxRetries            = 5
	defaultMaxDonationGoroutines = 4

	defaultTokenURL  = "https://vault.omise.co/tokens"
	defaultChargeURL = "https://api.omise.co/charges"
	currency         = "THB"
	returnURI        = "http://www.example.com/orders/complete"
)
