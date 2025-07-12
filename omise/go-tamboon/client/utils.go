package client

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	msgUnknownError        = "unknown error"
	msgDone                = "done."
	msgTotalReceived       = "        total received: THB %10s\n"
	msgSuccessfullyDonated = "  successfully donated: THB %10s\n"
	msgFaultyDonation      = "       faulty donation: THB %10s\n"
	msgAveragePerPerson    = "    average per person: THB %10s\n"
	msgTopDonors           = "            top donors:"
)

func parseOmiseError(body []byte) string {
	var errorResponse map[string]interface{}
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		return msgUnknownError
	}

	if object, ok := errorResponse["object"].(string); ok && object == "error" {
		if message, ok := errorResponse["message"].(string); ok {
			return message
		}
	}

	return msgUnknownError
}

func formatTHB(subunits int64) string {
	val := float64(subunits) / 100.0
	s := fmt.Sprintf("%.2f", val)
	n := len(s)
	dot := n - 3
	var out []byte
	for i := 0; i < dot; i++ {
		if (dot-i)%3 == 0 && i != 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	out = append(out, s[dot:]...)
	return string(out)
}

func printSummary(s *donationStats) {
	faultyAmount := s.totalAmount - s.successAmount
	avgPerPerson := int64(0)
	if s.totalCount > 0 {
		avgPerPerson = s.totalAmount / int64(s.totalCount)
	}

	type donorPair struct {
		name   string
		amount int64
	}
	topDonors := make([]donorPair, 0, len(s.donorAmounts))
	for name, amount := range s.donorAmounts {
		topDonors = append(topDonors, donorPair{name, amount})
	}
	sort.Slice(topDonors, func(i, j int) bool {
		return topDonors[i].amount > topDonors[j].amount
	})

	if len(topDonors) > 3 {
		topDonors = topDonors[:3]
	}

	fmt.Println(msgDone)
	fmt.Println()
	fmt.Printf(msgTotalReceived, formatTHB(s.totalAmount))
	fmt.Printf(msgSuccessfullyDonated, formatTHB(s.successAmount))
	fmt.Printf(msgFaultyDonation, formatTHB(faultyAmount))
	fmt.Println("")
	fmt.Printf(msgAveragePerPerson, formatTHB(int64(avgPerPerson)))
	fmt.Print(msgTopDonors)

	if len(topDonors) == 0 {
		fmt.Println()
		return
	}

	for i, donor := range topDonors {
		if i == 0 {
			fmt.Printf(" %s\n", donor.name)
		} else {
			fmt.Printf("                        %s\n", donor.name)
		}
	}
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "rate limit")
}
