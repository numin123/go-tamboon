package processor

import (
	"testing"
)

func TestParseCSVData(t *testing.T) {
	csvData := `Name,AmountSubunits,CCNumber,CVV,ExpMonth,ExpYear
John Doe,100000,4242424242424242,123,12,2025
Jane Smith,200000,5555555555554444,456,11,2026`

	records := ParseCSVData(csvData)

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	if records[0].Name != "John Doe" {
		t.Errorf("Expected first record name to be 'John Doe', got %s", records[0].Name)
	}

	if records[1].AmountSubunits != "200000" {
		t.Errorf("Expected second record amount to be '200000', got %s", records[1].AmountSubunits)
	}

	if records[0].CCNumber != "4242424242424242" {
		t.Errorf("Expected first record CCNumber to be '4242424242424242', got %s", records[0].CCNumber)
	}

	if records[1].CVV != "456" {
		t.Errorf("Expected second record CVV to be '456', got %s", records[1].CVV)
	}
}

func TestParseCSVDataWithInvalidData(t *testing.T) {
	csvData := `Name,AmountSubunits,CCNumber
John Doe,100000`

	records := ParseCSVData(csvData)

	if len(records) != 0 {
		t.Errorf("Expected 0 records for invalid data, got %d", len(records))
	}
}

func TestParseCSVDataEmpty(t *testing.T) {
	csvData := ""
	records := ParseCSVData(csvData)

	if len(records) != 0 {
		t.Errorf("Expected 0 records for empty data, got %d", len(records))
	}
}

func TestParseCSVDataHeaderOnly(t *testing.T) {
	csvData := "Name,AmountSubunits,CCNumber,CVV,ExpMonth,ExpYear"
	records := ParseCSVData(csvData)

	if len(records) != 0 {
		t.Errorf("Expected 0 records for header only, got %d", len(records))
	}
}

func TestParseCSVDataMaxRecordsLimit(t *testing.T) {
	csvData := `Name,AmountSubunits,CCNumber,CVV,ExpMonth,ExpYear
John Doe,100000,4242424242424242,123,12,2025
Jane Smith,200000,5555555555554444,456,11,2026
Bob Wilson,300000,4000000000000002,789,10,2027
Alice Brown,400000,5105105105105100,321,09,2028
Charlie Davis,500000,4111111111111111,654,08,2029`

	records := ParseCSVData(csvData)

	if len(records) != 2 {
		t.Errorf("Expected MaxRecords (2) records due to limit, got %d", len(records))
	}

	if records[0].Name != "John Doe" {
		t.Errorf("Expected first record name to be 'John Doe', got %s", records[0].Name)
	}

	if records[1].Name != "Jane Smith" {
		t.Errorf("Expected second record name to be 'Jane Smith', got %s", records[1].Name)
	}
}
