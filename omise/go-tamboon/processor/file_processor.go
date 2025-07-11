package processor

import (
	"fmt"
	"go-tamboon/cipher"
	"go-tamboon/client"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ReadAndDecryptFile(inputPath string) ([]client.DonationRecord, error) {
	if filepath.Ext(inputPath) != ".rot128" {
		return nil, fmt.Errorf("input file must have .rot128 extension")
	}

	inFile, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	reader, err := cipher.NewRot128Reader(inFile)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return ParseCSVData(string(data)), nil
}

func ParseCSVData(csvData string) []client.DonationRecord {
	lines := strings.Split(csvData, "\n")
	var records []client.DonationRecord

	for i, line := range lines {
		if line == "" || i == 0 {
			continue
		}

		if len(records) >= MaxRecords {
			break
		}

		row := strings.Split(line, ",")
		if len(row) >= 6 {
			record := client.DonationRecord{
				Name:           strings.TrimSpace(row[0]),
				AmountSubunits: strings.TrimSpace(row[1]),
				CCNumber:       strings.TrimSpace(row[2]),
				CVV:            strings.TrimSpace(row[3]),
				ExpMonth:       strings.TrimSpace(row[4]),
				ExpYear:        "2026",
			}
			records = append(records, record)
		}
	}

	return records
}
