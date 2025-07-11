package main

import (
	"go-tamboon/client"
	"go-tamboon/processor"
	"os"
	"testing"
)

func TestMainWorkflow(t *testing.T) {
	csvData := `Name,City,PostalCode,Amount
John Doe,Bangkok,10320,100000`

	records := processor.ParseCSVData(csvData)
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	omiseClient := client.NewOmiseClient()
	if omiseClient == nil {
		t.Error("Expected omise client to be created")
	}
}

func init() {
	os.Setenv("OMISE_PKEY", "test_public_key")
	os.Setenv("OMISE_SKEY", "test_secret_key")
}
