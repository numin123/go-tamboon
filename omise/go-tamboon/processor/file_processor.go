package processor

import (
	"bufio"
	"fmt"
	"go-tamboon/cipher"
	"go-tamboon/client"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func StreamAndDecryptFile(inputPath string) (<-chan client.DonationRecord, error) {
	out := make(chan client.DonationRecord)
	if filepath.Ext(inputPath) != ".rot128" {
		close(out)
		return out, fmt.Errorf("input file must have .rot128 extension")
	}

	inFile, err := os.Open(inputPath)
	if err != nil {
		close(out)
		return out, err
	}

	reader, err := cipher.NewRot128Reader(inFile)
	if err != nil {
		inFile.Close()
		close(out)
		return out, err
	}

	go func() {
		defer inFile.Close()
		defer close(out)
		scanner := bufio.NewScanner(reader)
		first := true
		count := 0
		for scanner.Scan() {
			if MaxRecords > 0 && count >= MaxRecords {
				break
			}
			line := scanner.Text()
			if first {
				first = false
				continue
			}
			if line == "" {
				continue
			}
			row := strings.Split(line, ",")
			if len(row) >= 6 {
				amountStr := strings.TrimSpace(row[ColAmountSubunits])
				amount, err := strconv.Atoi(amountStr)
				if err != nil {
					amount = 0
				}
				// TODO: Add ExpYearIncrease years to expYear to make some expired cards in test data will pass
				expYearStr := strings.TrimSpace(row[ColExpYear])
				expYear, err := strconv.Atoi(expYearStr)
				if err != nil {
					expYear = 0
				}
				expYear += ExpYearIncrease
				record := client.DonationRecord{
					Name:           strings.TrimSpace(row[ColName]),
					AmountSubunits: amount,
					CCNumber:       strings.TrimSpace(row[ColCCNumber]),
					CVV:            strings.TrimSpace(row[ColCVV]),
					ExpMonth:       strings.TrimSpace(row[ColExpMonth]),
					ExpYear:        strconv.Itoa(expYear),
				}
				out <- record
				count++
			}
		}
	}()
	return out, nil
}
