package main

import (
	"fmt"
	"log"
	"os"
	"pf-launcher/internal/pinata"
	"pf-launcher/internal/services"
	"pf-launcher/internal/types"
	"time"

	"github.com/joho/godotenv"
)

func LoadEnvironment() {
	ENV := os.Getenv("ENV")
	if ENV != "" {
		return
	}
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Error loading .env file", err)
	}
}

func main() {
	LoadEnvironment()

	start := time.Now()

	pinataClient := pinata.NewClient(os.Getenv("PINATA_JWT_SECRET"))

	rpcClient, err := services.NewRPCClient(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		log.Fatalf("Failed to create RPC client: %v", err)
	}

	imageHash, err := pinataClient.UploadFile("tweet_surge_io.jpg")
	if err != nil {
		log.Fatalf("Failed to upload image file: %v", err)
	}

	buyAmount := uint64(0.01 * 1e9)
	metadata := types.Metadata{
		Name:        "Test Token",
		Symbol:      "TEST",
		Description: "Test Description",
		Twitter:     "https://x.com/test",
		Telegram:    "https://t.me/test",
		Website:     "https://test.com",
		Image:       fmt.Sprintf("ipfs://%s", imageHash),
	}

	metadataHash, err := pinataClient.UploadJSON(metadata)
	if err != nil {
		log.Fatalf("Failed to upload metadata: %v", err)
	}

	metadataUri := fmt.Sprintf("ipfs://%s", metadataHash)

	err = rpcClient.LaunchToken(metadata, metadataUri, buyAmount)
	if err != nil {
		log.Fatalf("Failed to launch token: %v", err)
	}

	elapsed := time.Since(start)
	log.Printf("Launch took %s", elapsed)
}
