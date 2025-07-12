package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

	err := client.CreateCharge("100000", "tokn_test_123456789", "John Doe")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestProcessDonationsStream(t *testing.T) {
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

	recordCh := make(chan DonationRecord)
	go func() {
		for _, r := range records {
			recordCh <- r
		}
		close(recordCh)
	}()

	client.ProcessDonationsStream(recordCh)
}

func (c *OmiseClient) CreateToken(name, ccNumber, cvv, expMonth, expYear string) (string, error) {
	return c.tokenService.CreateToken(name, ccNumber, cvv, expMonth, expYear)
}

func (c *OmiseClient) CreateCharge(amount string, tokenID, description string) error {
	return c.chargeService.CreateCharge(amount, tokenID, description)
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

func TestPrintSummary(t *testing.T) {
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

	failingTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"object": "error", "message": "invalid card"}`))
	}))
	defer failingTokenServer.Close()

	cases := []struct {
		name             string
		records          []DonationRecord
		check            []string
		useFailingServer bool
	}{
		{
			name: "single donor",
			records: []DonationRecord{
				{Name: "John Doe", AmountSubunits: "100000", CCNumber: "4242424242424242", CVV: "123", ExpMonth: "12", ExpYear: "2025"},
			},
			check: []string{"done.", "total received: THB", "successfully donated: THB", "faulty donation: THB", "average per person: THB", "top donors:", "John Doe"},
		},
		{
			name:    "no donations",
			records: []DonationRecord{},
			check:   []string{"done.", "total received: THB", "successfully donated: THB", "faulty donation: THB", "average per person: THB", "top donors:"},
		},
		{
			name: "multiple top donors",
			records: []DonationRecord{
				{Name: "Alice", AmountSubunits: "120000", CCNumber: "4242424242424242", CVV: "123", ExpMonth: "12", ExpYear: "2025"},
				{Name: "Bob", AmountSubunits: "100000", CCNumber: "5555555555554444", CVV: "456", ExpMonth: "11", ExpYear: "2026"},
				{Name: "Carol", AmountSubunits: "80000", CCNumber: "4111111111111111", CVV: "789", ExpMonth: "10", ExpYear: "2027"},
				{Name: "Dave", AmountSubunits: "50000", CCNumber: "4000000000000002", CVV: "321", ExpMonth: "09", ExpYear: "2028"},
			},
			check: []string{"done.", "total received: THB", "successfully donated: THB", "faulty donation: THB", "average per person: THB", "top donors:", "Alice", "Bob", "Carol"},
		},
		{
			name: "all successful",
			records: []DonationRecord{
				{Name: "Donor1", AmountSubunits: "100000", CCNumber: "4242424242424242", CVV: "123", ExpMonth: "12", ExpYear: "2025"},
				{Name: "Donor2", AmountSubunits: "100000", CCNumber: "5555555555554444", CVV: "456", ExpMonth: "11", ExpYear: "2026"},
			},
			check: []string{"successfully donated: THB", "faulty donation: THB       0.00", "Donor1", "Donor2"},
		},
		{
			name: "all failed",
			records: []DonationRecord{
				{Name: "Fail1", AmountSubunits: "100000", CCNumber: "0000000000000000", CVV: "000", ExpMonth: "01", ExpYear: "2000"},
				{Name: "Fail2", AmountSubunits: "100000", CCNumber: "0000000000000000", CVV: "000", ExpMonth: "01", ExpYear: "2000"},
			},
			check:            []string{"successfully donated: THB       0.00", "faulty donation: THB   2,000.00", "top donors:"},
			useFailingServer: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var client *OmiseClient
			if c.useFailingServer {
				client = NewOmiseClientWithURLs(failingTokenServer.URL, mockChargeServer.URL)
			} else {
				client = NewOmiseClientWithURLs(mockTokenServer.URL, mockChargeServer.URL)
			}

			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			recordCh := make(chan DonationRecord)
			go func() {
				for _, r := range c.records {
					recordCh <- r
				}
				close(recordCh)
			}()

			client.ProcessDonationsStream(recordCh)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			for _, expect := range c.check {
				if !strings.Contains(output, expect) {
					t.Errorf("[%s] Expected output to contain '%s', but got:\n%s", c.name, expect, output)
				}
			}
		})
	}
}
