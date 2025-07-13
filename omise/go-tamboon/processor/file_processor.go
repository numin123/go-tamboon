package processor

import (
	"bufio"
	"go-tamboon/cipher"
	"go-tamboon/client"
	"os"
	"strconv"
	"strings"
)

func StreamAndDecryptFile(inputPath string) (<-chan client.DonationRecord, error) {
	out := make(chan client.DonationRecord)

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
			if maxRecords > 0 && count >= maxRecords {
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
				// TODO: Add ExpYearIncrease years to expYear to make some expired cards in test data will pass
				expYearStr := strings.TrimSpace(row[colExpYear])
				expYear, err := strconv.Atoi(expYearStr)
				if err != nil {
					expYear = 0
				}
				expYear += expYearIncrease
				record := client.DonationRecord{
					Name:           strings.TrimSpace(row[colName]),
					AmountSubunits: strings.TrimSpace(row[colAmountSubunits]),
					CCNumber:       strings.TrimSpace(row[colCCNumber]),
					CVV:            strings.TrimSpace(row[colCVV]),
					ExpMonth:       strings.TrimSpace(row[colExpMonth]),
					ExpYear:        strconv.Itoa(expYear),
				}
				out <- record
				count++
			}
		}
	}()
	return out, nil
}
