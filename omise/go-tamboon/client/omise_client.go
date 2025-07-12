package client

import (
	"fmt"
	"log"
	"sync"
)

func (c *OmiseClient) ProcessDonationsStream(recordCh <-chan DonationRecord) {
	s := &donationStats{
		donorAmounts: make(map[string]int),
	}

	var wg sync.WaitGroup

	for record := range recordCh {
		amount := record.AmountSubunits

		s.mu.Lock()
		s.totalCount++
		s.totalAmount += amount
		s.mu.Unlock()

		wg.Add(1)
		go func(r DonationRecord, amt int) {
			defer wg.Done()

			err := c.processSingleDonation(r)

			s.mu.Lock()
			if err != nil {
				log.Printf("Error processing donation for %s: %v", r.Name, err)
			} else {
				s.successCount++
				s.successAmount += amt
				s.donorAmounts[r.Name] += amt
			}
			s.mu.Unlock()
		}(record, amount)
	}

	wg.Wait()
	printSummary(s)
}

func NewOmiseClient() *OmiseClient {
	return &OmiseClient{
		tokenService:  NewTokenService(),
		chargeService: NewChargeService(),
	}
}

func (c *OmiseClient) processSingleDonation(record DonationRecord) error {
	tokenID, err := c.tokenService.CreateToken(record.Name, record.CCNumber, record.CVV, record.ExpMonth, record.ExpYear)
	if err != nil {
		return fmt.Errorf("creating token: %v", err)
	}

	description := fmt.Sprintf("charge for %s", record.Name)
	err = c.chargeService.CreateCharge(record.AmountSubunits, tokenID, description)
	if err != nil {
		return fmt.Errorf("creating charge: %v", err)
	}

	return nil
}
