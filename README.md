# go-tamboon

Omise challenges Repository: https://github.com/omise/challenges/tree/challenge-go

A command-line tool for processing donation files encrypted with rot128, as part of the Omise challenges.


## Environment Setup


Create a `.env` file in the `go-tamboon` directory with the following variables:

```dotenv
# Omise API Credentials
OMISE_PKEY=your_public_key         # Omise public API key
OMISE_SKEY=your_secret_key         # Omise secret API key

# Omise API Endpoints (override if using a different environment)
OMISE_TOKEN_URL=https://vault.omise.co/tokens   # Token endpoint URL
OMISE_CHARGE_URL=https://api.omise.co/charges   # Charge endpoint URL

# Application Settings
MAX_RETRIES=5                      # Maximum number of retry attempts for failed operations
MAX_DONATION_GOROUTINES=4          # Maximum number of concurrent donation goroutines
MAX_RECORDS=10                     # Maximum number of records to process (0 means no limit)
EXP_YEAR_INCREASE=10               # Number of years to increase the card expiration year for test data
```

Replace `your_public_key` and `your_secret_key` with your actual Omise API keys. Adjust other values as needed for your environment or testing.

## How to Setup

1. Clone this repository:
   ```
   git clone https://github.com/omise/challenges.git
   cd challenges/omise/go-tamboon
   ```

2. Install the binary:
   ```
   go install -v .
   ```

3. Run the program with a CSV file:
   ```
   $GOPATH/bin/go-tamboon test.csv
   ```

## Example Output

```
performing donations...
done.

        total received: THB  210,000.00
  successfully donated: THB  200,000.00
       faulty donation: THB   10,000.00

    average per person: THB      534.23
            top donors: Obi-wan Kenobi
                        Luke Skywalker
                        Kylo Ren
```

## Notes
- Replace `test.csv` with your own encrypted file if needed.
- Make sure your `$GOPATH` is set and `$GOPATH/bin` is in your `PATH`.

## Project Requirements Coverage

This project fully covers all the requirements and bonus points for the Omise go-tamboon challenge. Below is a summary of how each requirement is addressed, with links to the relevant files and real code examples for easier understanding:

### 1. Decrypt the file using a simple ROT-128 algorithm
- **Implemented in:** [`cipher/rot128.go`](cipher/rot128.go)
- **Example:**
```go
// cipher/rot128.go
func NewRot128Reader(r io.Reader) (*Rot128Reader, error) {
	return &Rot128Reader{reader: r}, nil
}
func (r *Rot128Reader) Read(p []byte) (int, error) {
	if n, err := r.reader.Read(p); err != nil {
		return n, err
	} else {
		rot128(p[:n])
		return n, nil
	}
}
func rot128(buf []byte) {
	for idx := range buf {
		buf[idx] += 128
	}
}
```
- **Usage in pipeline:**
```go
// processor/file_processor.go
reader, err := cipher.NewRot128Reader(inFile)
```

### 2. Make donations by creating a Charge via the Charge API for each row in the decrypted CSV
- **Implemented in:** [`client/charge_service.go`](omise/go-tamboon/client/charge_service.go)
- **Example:**
```go
// client/charge_service.go
func (cs *ChargeService) CreateCharge(amount, tokenID, description string) error {
	data := url.Values{}
	data.Set("description", description)
	data.Set("amount", amount)
	data.Set("currency", currency)
	data.Set("return_uri", returnURI)
	data.Set("card", tokenID)

	req, err := http.NewRequest("POST", cs.chargeURL, strings.NewReader(data.Encode()))
	// ...
	resp, err := client.Do(req)
	// ...
}
```
- **Usage in donation loop:**
```go
// client/omise_client.go
err := c.chargeService.CreateChargeWithRateLimit(
	record.AmountSubunits, tokenID, description, c.rateLimiter)
```

### 3. Produce a brief summary at the end
- **Implemented in:** [`client/omise_client.go`](omise/go-tamboon/client/omise_client.go)
- **Example:**
```go
// client/omise_client.go
func (c *OmiseClient) ProcessDonationsStream(recordCh <-chan DonationRecord) {
	// ...
	wg.Wait()
	printSummary(s)
}
```

### 4. Handle errors gracefully, without stopping the entire process
- **Implemented in:** [`client/omise_client.go`](omise/go-tamboon/client/omise_client.go)
- **Example:**
```go
// client/omise_client.go
go func(r DonationRecord, amt int64) {
	// ...
	err := c.processSingleDonation(r)
	if err != nil {
		log.Printf("Error processing donation for %s: %v", r.Name, err)
	}
	// ...
}(record, amount)
```

### 5. Writes readable and maintainable code
- **Code style:** All code is organized into clear packages and files, with tests and configuration separated.
- **See:** [`omise/go-tamboon/`](omise/go-tamboon/) directory structure.

---

## Bonus Points Coverage

### 1. Good Go package structure
- **Packages:** `cipher`, `client`, `processor`, and `data` directories under [`omise/go-tamboon/`](omise/go-tamboon/)

### 2. Throttle API calls if rate limit is hit
- **Implemented in:** [`client/rate_limiter.go`](omise/go-tamboon/client/rate_limiter.go)
- **Example:**
```go
// client/rate_limiter.go
func (rl *RateLimiter) WaitIfPaused() {
	rl.mu.Lock()
	for rl.paused {
		rl.cond.Wait()
	}
	rl.mu.Unlock()
}
```
- **Usage in donation:**
```go
// client/omise_client.go
c.rateLimiter.WaitIfPaused()
```

### 3. Run as fast as possible on a multi-core CPU
- **Concurrency:** Controlled by `MAX_DONATION_GOROUTINES` in `.env` and implemented in [`client/omise_client.go`](omise/go-tamboon/client/omise_client.go)
- **Example:**
```go
// client/omise_client.go
sem := make(chan struct{}, maxDonationGoroutines)
go func(r DonationRecord, amt int64) {
	defer wg.Done()
	defer func() { <-sem }()
	// ...
}(record, amount)
```

### 4. Allocate as little memory as possible
- **Efficient processing:** Streams and processes records without loading the entire file into memory. See [`processor/file_processor.go`](omise/go-tamboon/processor/file_processor.go)
- **Example:**
```go
// processor/file_processor.go
scanner := bufio.NewScanner(reader)
for scanner.Scan() {
	// process each line
}
```

### 5. No large trace of Credit Card numbers in memory or disk
- **Secure handling:** Card data is used only for each donation and not saved anywhere else. See [`client/omise_client.go`](omise/go-tamboon/client/omise_client.go)
- **Example:**
```go
// client/omise_client.go
tokenID, err := c.tokenService.CreateTokenWithRateLimit(
	record.Name, record.CCNumber, record.CVV, record.ExpMonth, record.ExpYear, c.rateLimiter)
// Card data is not saved after this
```
- **Why it is secure:**
  - The decrypted file is never written to disk; it is streamed and processed in-memory only.

### 6. Ensure reproducible builds
- **Go modules:** [`go.mod`](omise/go-tamboon/go.mod), [`go.sum`](omise/go-tamboon/go.sum)
- **Instructions:** See setup steps above for reproducible environment.

For more details, see the code and comments in each linked file.
