package client

import (
	"fmt"
	"log"
	"strconv"
	"sync"
)

func (c *OmiseClient) ProcessDonationsStream(recordCh <-chan DonationRecord) {
	s := &donationStats{
		donorAmounts: make(map[string]int64),
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, MaxDonationGoroutines)

	for record := range recordCh {
		amount, _ := strconv.ParseInt(record.AmountSubunits, 10, 64)

		s.mu.Lock()
		s.totalCount++
		s.totalAmount += amount
		s.mu.Unlock()

		wg.Add(1)
		sem <- struct{}{}
		go func(r DonationRecord, amt int64) {
			defer wg.Done()
			defer func() { <-sem }()

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
		rateLimiter:   NewRateLimiter(),
	}
}

func (c *OmiseClient) processSingleDonation(record DonationRecord) error {
	c.rateLimiter.WaitIfPaused()
	tokenID, err := c.tokenService.CreateTokenWithRateLimit(
		record.Name, record.CCNumber, record.CVV, record.ExpMonth, record.ExpYear, c.rateLimiter)
	if err != nil {
		return fmt.Errorf("creating token: %v", err)
	}

	c.rateLimiter.WaitIfPaused()
	description := fmt.Sprintf("charge for %s", record.Name)
	err = c.chargeService.CreateChargeWithRateLimit(
		record.AmountSubunits, tokenID, description, c.rateLimiter)
	if err != nil {
		return fmt.Errorf("creating charge: %v", err)
	}

	return nil
}
