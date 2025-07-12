package client

const (
	maxRetries            = 5
	MaxDonationGoroutines = 10

	DefaultTokenURL  = "https://vault.omise.co/tokens"
	DefaultChargeURL = "https://api.omise.co/charges"
	Currency         = "THB"
	ReturnURI        = "http://www.example.com/orders/complete"
)
