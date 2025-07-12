package main

import (
	"go-tamboon/cipher"
	"go-tamboon/client"
	"go-tamboon/processor"
	"os"
	"path/filepath"
	"testing"
)

func TestMainWorkflow(t *testing.T) {
	t.Setenv("OMISE_PKEY", "test_public_key")
	t.Setenv("OMISE_SKEY", "test_secret_key")

	csv := "Name,AmountSubunits,CCNumber,CVV,ExpMonth,ExpYear\nJohn Doe,5000,4242424242424242,123,12,2026\nJane Smith,10000,4000000000000002,456,06,2026\n"
	rot128Path := createTestROT128File(t, csv)

	recordCh, err := processor.StreamAndDecryptFile(rot128Path)
	if err != nil {
		t.Fatalf("StreamAndDecryptFile failed: %v", err)
	}

	omiseClient := client.NewOmiseClient()
	if omiseClient == nil {
		t.Error("Expected omise client to be created")
	}

	var recordCount int
	for record := range recordCh {
		recordCount++
		if record.Name == "" {
			t.Error("Expected non-empty name")
		}
		if record.AmountSubunits == "" {
			t.Error("Expected non-empty amount")
		}
	}

	if recordCount == 0 {
		t.Error("Expected to receive donation records")
	}
	t.Logf("Processed %d donation records", recordCount)
}

func TestMainWorkflowMissingArgs(t *testing.T) {
	recordCh, err := processor.StreamAndDecryptFile("nonexistent.csv")
	if err == nil {
		t.Error("Expected error for file without .rot128 extension")
	}

	select {
	case _, ok := <-recordCh:
		if ok {
			t.Error("Expected channel to be closed when error occurs")
		}
	default:
	}
}

func TestMainWorkflowInvalidFile(t *testing.T) {
	recordCh, err := processor.StreamAndDecryptFile("nonexistent.rot128")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	select {
	case _, ok := <-recordCh:
		if ok {
			t.Error("Expected channel to be closed when error occurs")
		}
	default:
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
