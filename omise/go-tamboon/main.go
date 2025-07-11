package main

import (
	"fmt"
	"go-tamboon/client"
	"go-tamboon/processor"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: go-tamboon <inputfile.rot128>")
		return
	}

	inputPath := os.Args[1]
	fmt.Println("performing donations...")

	records, err := processor.ReadAndDecryptFile(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	if len(records) == 0 {
		fmt.Println("No donation records found in file")
		return
	}

	omiseClient := client.NewOmiseClient()
	omiseClient.ProcessDonations(records)

	fmt.Println("All donations completed!")
}
