package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type ChargeService struct {
	chargeURL string
}

func NewChargeService() *ChargeService {
	chargeURL := os.Getenv("OMISE_CHARGE_URL")
	if chargeURL == "" {
		chargeURL = DefaultChargeURL
	}
	return &ChargeService{
		chargeURL: chargeURL,
	}
}

func (cs *ChargeService) CreateCharge(amount int, tokenID, description string) error {
	data := url.Values{}
	data.Set("description", description)
	data.Set("amount", fmt.Sprintf("%d", amount))
	data.Set("currency", Currency)
	data.Set("return_uri", ReturnURI)
	data.Set("card", tokenID)

	req, err := http.NewRequest("POST", cs.chargeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("error creating charge request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(os.Getenv("OMISE_SKEY"), "")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making charge request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading charge response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		errorMsg := parseOmiseError(body)
		return fmt.Errorf("API error: %s", errorMsg)
	}

	fmt.Printf("Donation processed (Amount: %s %s)\n", formatTHB(amount), Currency)
	return nil
}
