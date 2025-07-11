package client

type DonationRecord struct {
	Name           string
	AmountSubunits string
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
