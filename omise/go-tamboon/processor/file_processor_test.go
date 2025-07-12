package processor

import (
	"fmt"
	"go-tamboon/cipher"
	"go-tamboon/client"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStreamAndDecryptFile_Success(t *testing.T) {
	testData := "Name,AmountSubunits,CCNumber,CVV,ExpMonth,ExpYear\nJohn Doe,5000,4242424242424242,123,12,2026\nJane Smith,10000,4000000000000002,456,06,2026"

	tempFile := createTestROT128File(t, testData)
	defer os.Remove(tempFile)

	ch, err := StreamAndDecryptFile(tempFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var records []client.DonationRecord
	for record := range ch {
		records = append(records, record)
	}

	expectedRecords := []client.DonationRecord{
		{
			Name:           "John Doe",
			AmountSubunits: "5000",
			CCNumber:       "4242424242424242",
			CVV:            "123",
			ExpMonth:       "12",
			ExpYear:        fmt.Sprintf("%d", 2026+ExpYearIncrease),
		},
		{
			Name:           "Jane Smith",
			AmountSubunits: "10000",
			CCNumber:       "4000000000000002",
			CVV:            "456",
			ExpMonth:       "06",
			ExpYear:        fmt.Sprintf("%d", 2026+ExpYearIncrease),
		},
	}

	if len(records) != len(expectedRecords) {
		t.Fatalf("Expected %d records, got %d", len(expectedRecords), len(records))
	}

	for i, expected := range expectedRecords {
		if records[i] != expected {
			t.Errorf("Record %d mismatch. Expected %+v, got %+v", i, expected, records[i])
		}
	}
}

func TestStreamAndDecryptFile_MaxRecordsLimit(t *testing.T) {
	testData := "Name,AmountSubunits,CCNumber,CVV,ExpMonth,ExpYear\nJohn Doe,5000,4242424242424242,123,12,2026\nJane Smith,10000,4000000000000002,456,06,2026\nBob Wilson,15000,4111111111111111,789,03,2026\nAlice Brown,20000,4222222222222222,123,01,2026\nCharlie Davis,25000,4333333333333333,456,02,2026\nEve Johnson,30000,4444444444444444,789,03,2026\nFrank Miller,35000,4555555555555555,123,04,2026"

	tempFile := createTestROT128File(t, testData)
	defer os.Remove(tempFile)

	ch, err := StreamAndDecryptFile(tempFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var records []client.DonationRecord
	for record := range ch {
		records = append(records, record)
	}

	if len(records) != MaxRecords {
		t.Errorf("Expected %d records due to MaxRecords limit, got %d", MaxRecords, len(records))
	}
}

func TestStreamAndDecryptFile_EmptyLines(t *testing.T) {
	testData := "Name,AmountSubunits,CCNumber,CVV,ExpMonth,ExpYear\nJohn Doe,5000,4242424242424242,123,12,2026\n\n\nJane Smith,10000,4000000000000002,456,06,2026\n"

	tempFile := createTestROT128File(t, testData)
	defer os.Remove(tempFile)

	ch, err := StreamAndDecryptFile(tempFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var records []client.DonationRecord
	for record := range ch {
		records = append(records, record)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records (empty lines should be skipped), got %d", len(records))
	}
}

func TestStreamAndDecryptFile_InsufficientColumns(t *testing.T) {
	testData := "Name,AmountSubunits,CCNumber,CVV,ExpMonth,ExpYear\nJohn Doe,5000\nJane Smith,10000,4000000000000002,456,06,2026"

	tempFile := createTestROT128File(t, testData)
	defer os.Remove(tempFile)

	ch, err := StreamAndDecryptFile(tempFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var records []client.DonationRecord
	for record := range ch {
		records = append(records, record)
	}

	if len(records) != 1 {
		t.Errorf("Expected 1 record (line with insufficient columns should be skipped), got %d", len(records))
	}

	expected := client.DonationRecord{
		Name:           "Jane Smith",
		AmountSubunits: "10000",
		CCNumber:       "4000000000000002",
		CVV:            "456",
		ExpMonth:       "06",
		ExpYear:        fmt.Sprintf("%d", 2026+ExpYearIncrease),
	}

	if records[0] != expected {
		t.Errorf("Expected %+v, got %+v", expected, records[0])
	}
}

func TestStreamAndDecryptFile_WrongExtension(t *testing.T) {
	tempFile := createTempFile(t, "test.csv", "test data")
	defer os.Remove(tempFile)

	ch, err := StreamAndDecryptFile(tempFile)
	if err == nil {
		t.Fatal("Expected error for wrong file extension")
	}

	if !strings.Contains(err.Error(), ".rot128 extension") {
		t.Errorf("Expected error about .rot128 extension, got %v", err)
	}

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Channel should be closed immediately")
	}
}

func TestStreamAndDecryptFile_FileNotFound(t *testing.T) {
	ch, err := StreamAndDecryptFile("nonexistent.rot128")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Channel should be closed immediately")
	}
}

func TestStreamAndDecryptFile_HeaderOnlyFile(t *testing.T) {
	testData := "Name,AmountSubunits,CCNumber,CVV,ExpMonth,ExpYear"

	tempFile := createTestROT128File(t, testData)
	defer os.Remove(tempFile)

	ch, err := StreamAndDecryptFile(tempFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var records []client.DonationRecord
	for record := range ch {
		records = append(records, record)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records (header only), got %d", len(records))
	}
}

func TestStreamAndDecryptFile_WhitespaceHandling(t *testing.T) {
	testData := "Name,AmountSubunits,CCNumber,CVV,ExpMonth,ExpYear\n  John Doe  ,  5000  ,  4242424242424242  ,  123  ,  12  ,2026"

	tempFile := createTestROT128File(t, testData)
	defer os.Remove(tempFile)

	ch, err := StreamAndDecryptFile(tempFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var records []client.DonationRecord
	for record := range ch {
		records = append(records, record)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	expected := client.DonationRecord{
		Name:           "John Doe",
		AmountSubunits: "5000",
		CCNumber:       "4242424242424242",
		CVV:            "123",
		ExpMonth:       "12",
		ExpYear:        fmt.Sprintf("%d", 2026+ExpYearIncrease),
	}

	if records[0] != expected {
		t.Errorf("Expected %+v, got %+v", expected, records[0])
	}
}

func createTestROT128File(t *testing.T, data string) string {
	tempFile := createTempFile(t, "test.rot128", "")

	file := createFileWriter(t, tempFile)
	defer file.Close()

	writer, err := cipher.NewRot128Writer(file)
	if err != nil {
		t.Fatalf("Failed to create ROT128 writer: %v", err)
	}

	_, err = writer.Write([]byte(data))
	if err != nil {
		t.Fatalf("Failed to write encrypted data: %v", err)
	}

	return tempFile
}

func createTempFile(t *testing.T, filename, content string) string {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, filename)

	if content != "" {
		err := os.WriteFile(tempFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
	}

	return tempFile
}

func createFileWriter(t *testing.T, filename string) *os.File {
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	return file
}
