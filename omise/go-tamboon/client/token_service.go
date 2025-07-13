package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type TokenService struct {
	tokenURL string
}

func NewTokenService() *TokenService {
	tokenURL := os.Getenv("OMISE_TOKEN_URL")
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}
	return &TokenService{
		tokenURL: tokenURL,
	}
}

func (ts *TokenService) CreateToken(name, ccNumber, cvv, expMonth, expYear string) (string, error) {
	data := url.Values{}
	data.Set("card[name]", name)
	data.Set("card[number]", ccNumber)
	data.Set("card[security_code]", cvv)
	data.Set("card[expiration_month]", expMonth)
	data.Set("card[expiration_year]", expYear)

	req, err := http.NewRequest("POST", ts.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(os.Getenv("OMISE_PKEY"), "")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		errorMsg := parseOmiseError(body)
		return "", fmt.Errorf("API error: %s", errorMsg)
	}

	var tokenResponse map[string]interface{}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("error parsing token response: %v", err)
	}

	tokenID, ok := tokenResponse["id"].(string)
	if !ok {
		return "", fmt.Errorf("error extracting token ID from response")
	}

	return tokenID, nil
}

func (ts *TokenService) CreateTokenWithRateLimit(name, ccNumber, cvv, expMonth, expYear string, rl *RateLimiter) (string, error) {
	retries := 0
	for {
		tokenID, err := ts.CreateToken(name, ccNumber, cvv, expMonth, expYear)
		if err != nil && isRateLimitError(err) {
			if retries >= maxRetries {
				return "", fmt.Errorf("rate limit: exceeded max retries")
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
		return tokenID, err
	}
}
