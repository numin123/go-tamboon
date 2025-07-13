package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type ChargeService struct {
	chargeURL string
}

func NewChargeService() *ChargeService {
	chargeURL := os.Getenv("OMISE_CHARGE_URL")
	if chargeURL == "" {
		chargeURL = defaultChargeURL
	}
	return &ChargeService{
		chargeURL: chargeURL,
	}
}

func (cs *ChargeService) CreateCharge(amount, tokenID, description string) error {
	data := url.Values{}
	data.Set("description", description)
	data.Set("amount", amount)
	data.Set("currency", currency)
	data.Set("return_uri", returnURI)
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

	amtInt64, _ := strconv.ParseInt(amount, 10, 64)
	fmt.Printf("Donation processed (Amount: %s %s)\n", formatTHB(amtInt64), currency)
	return nil
}

func (cs *ChargeService) CreateChargeWithRateLimit(amount, tokenID, description string, rl *RateLimiter) error {
	retries := 0
	for {
		err := cs.CreateCharge(amount, tokenID, description)
		if err != nil && isRateLimitError(err) {
			if retries >= maxRetries {
				return fmt.Errorf("rate limit: exceeded max retries")
			}
			rl.Pause()
			waitTime := time.Duration(5*(retries+1)) * time.Second
			go func() {
				time.Sleep(waitTime)
				rl.Resume()
			}()
			rl.WaitIfPaused()
			retries++
			continue
		}
		return err
	}
}
