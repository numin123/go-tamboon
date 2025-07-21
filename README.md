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
   git clone https://github.com/numin123/go-tamboon.git
   cd omise/go-tamboon
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

```
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

## Project Requirements Coverage

Project Requirements:

- Decrypt donation files (ROT-128): [`cipher/rot128.go`](cipher/rot128.go)
- Process CSV donations: [`client/charge_service.go`](omise/go-tamboon/client/charge_service.go)
- Show summary at the end
- Keep running if some donations fail (errors are logged)
- Clean code, organized in: `cipher`, `client`, `processor`, `data`

Bonus:
- Handles API rate limits: [`client/rate_limiter.go`](omise/go-tamboon/client/rate_limiter.go)
- Fast (multi-core)
- Low memory use
- Credit card data never saved
- Reproducible builds (Go modules)
