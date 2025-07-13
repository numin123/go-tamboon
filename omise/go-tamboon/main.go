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

	processor.InitConfig()
	client.InitConfig()

	if len(os.Args) < 2 {
		fmt.Println("Usage: go-tamboon <inputfile.rot128>")
		return
	}

	inputPath := os.Args[1]
	fmt.Println("performing donations...")

	recordCh, err := processor.StreamAndDecryptFile(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	omiseClient := client.NewOmiseClient()
	omiseClient.ProcessDonationsStream(recordCh)

	fmt.Println("All donations completed!")
}
