package client

import (
	"fmt"
	"log"
	"sync"
)

type OmiseClient struct {
	tokenService  *TokenService
	chargeService *ChargeService
}

func NewOmiseClient() *OmiseClient {
	return &OmiseClient{
		tokenService:  NewTokenService(),
		chargeService: NewChargeService(),
	}
}

func (c *OmiseClient) ProcessDonations(records []DonationRecord) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	totalCount := len(records)

	for _, record := range records {
		wg.Add(1)
		go func(r DonationRecord) {
			defer wg.Done()
			err := c.processSingleDonation(r)
			if err != nil {
				log.Printf("Error processing donation for %s: %v", r.Name, err)
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(record)
	}

	wg.Wait()
	fmt.Printf("Successfully processed %d out of %d donations\n", successCount, totalCount)
}

func (c *OmiseClient) processSingleDonation(record DonationRecord) error {
	tokenID, err := c.tokenService.CreateToken(record.Name, record.CCNumber, record.CVV, record.ExpMonth, record.ExpYear)
	if err != nil {
		return fmt.Errorf("creating token: %v", err)
	}

	err = c.chargeService.CreateCharge(record.Name, record.AmountSubunits, tokenID)
	if err != nil {
		return fmt.Errorf("creating charge: %v", err)
	}

	return nil
}
