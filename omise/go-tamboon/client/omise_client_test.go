package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateToken(t *testing.T) {
	mockResponse := map[string]interface{}{
		"object": "token",
		"id":     "tokn_test_123456789",
		"card": map[string]interface{}{
			"name": "John Doe",
			"city": "Bangkok",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/tokens" {
			t.Errorf("Expected /tokens path, got %s", r.URL.Path)
		}

		err := r.ParseForm()
		if err != nil {
			t.Fatal(err)
		}

		if r.FormValue("card[name]") != "John Doe" {
			t.Errorf("Expected card[name] to be 'John Doe', got %s", r.FormValue("card[name]"))
		}

		if r.FormValue("card[number]") != "4242424242424242" {
			t.Errorf("Expected card[number] to be '4242424242424242', got %s", r.FormValue("card[number]"))
		}

		if r.FormValue("card[security_code]") != "123" {
			t.Errorf("Expected card[security_code] to be '123', got %s", r.FormValue("card[security_code]"))
		}

		if r.FormValue("card[expiration_month]") != "12" {
			t.Errorf("Expected card[expiration_month] to be '12', got %s", r.FormValue("card[expiration_month]"))
		}

		if r.FormValue("card[expiration_year]") != "2025" {
			t.Errorf("Expected card[expiration_year] to be '2025', got %s", r.FormValue("card[expiration_year]"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewOmiseClientWithURLs(server.URL+"/tokens", "https://api.omise.co/charges")

	tokenID, err := client.CreateToken("John Doe", "4242424242424242", "123", "12", "2025")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if tokenID != "tokn_test_123456789" {
		t.Errorf("Expected token ID 'tokn_test_123456789', got %s", tokenID)
	}
}

func TestCreateCharge(t *testing.T) {
	mockResponse := map[string]interface{}{
		"object": "charge",
		"id":     "chrg_test_123456789",
		"amount": 100000,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/charges" {
			t.Errorf("Expected /charges path, got %s", r.URL.Path)
		}

		err := r.ParseForm()
		if err != nil {
			t.Fatal(err)
		}

		if r.FormValue("amount") != "100000" {
			t.Errorf("Expected amount to be '100000', got %s", r.FormValue("amount"))
		}

		if r.FormValue("card") != "tokn_test_123456789" {
			t.Errorf("Expected card to be 'tokn_test_123456789', got %s", r.FormValue("card"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewOmiseClientWithURLs("https://vault.omise.co/tokens", server.URL+"/charges")

	err := client.CreateCharge("John Doe", "100000", "tokn_test_123456789")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestProcessDonations(t *testing.T) {
	records := []DonationRecord{
		{Name: "John Doe", AmountSubunits: "100000", CCNumber: "4242424242424242", CVV: "123", ExpMonth: "12", ExpYear: "2025"},
		{Name: "Jane Smith", AmountSubunits: "200000", CCNumber: "5555555555554444", CVV: "456", ExpMonth: "11", ExpYear: "2026"},
	}

	mockTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := map[string]interface{}{
			"object": "token",
			"id":     "tokn_test_123456789",
		}
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockTokenServer.Close()

	mockChargeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := map[string]interface{}{
			"object": "charge",
			"id":     "chrg_test_123456789",
		}
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockChargeServer.Close()

	client := NewOmiseClientWithURLs(mockTokenServer.URL, mockChargeServer.URL)

	client.ProcessDonations(records)
}

func (c *OmiseClient) CreateToken(name, ccNumber, cvv, expMonth, expYear string) (string, error) {
	return c.tokenService.CreateToken(name, ccNumber, cvv, expMonth, expYear)
}

func (c *OmiseClient) CreateCharge(description, amount, tokenID string) error {
	return c.chargeService.CreateCharge(description, amount, tokenID)
}

func NewOmiseClientWithURLs(tokenURL, chargeURL string) *OmiseClient {
	oldTokenURL := os.Getenv("OMISE_TOKEN_URL")
	oldChargeURL := os.Getenv("OMISE_CHARGE_URL")

	os.Setenv("OMISE_TOKEN_URL", tokenURL)
	os.Setenv("OMISE_CHARGE_URL", chargeURL)

	client := &OmiseClient{
		tokenService:  NewTokenService(),
		chargeService: NewChargeService(),
	}

	os.Setenv("OMISE_TOKEN_URL", oldTokenURL)
	os.Setenv("OMISE_CHARGE_URL", oldChargeURL)

	return client
}

func init() {
	os.Setenv("OMISE_PKEY", "test_public_key")
	os.Setenv("OMISE_SKEY", "test_secret_key")
}
