package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	ethcashier "github.com/gotsteez/eth_cashier"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("./configs/.env"); err != nil {
		log.Fatalf("env could not be loaded correctly", err)
	}
	dbPath := "database.db"
	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			fmt.Printf("Failed to create database file: %v\n", err)
			return
		}
		file.Close()
		fmt.Println("Database file created successfully")
	}
	db, err := ethcashier.InitDB(dbPath)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		fmt.Println("RPC URL is missing from env variables")
		return
	}
	rpc, err := ethcashier.NewRPCClient(rpcURL)
	if err != nil {
		fmt.Printf("Failed to initialize rpc client: %v\n", err)
		return
	}

	cmcAPIKey := os.Getenv("CMC_API_KEY")
	if cmcAPIKey == "" {
		fmt.Println("CMC API Key is missing")
		return
	}
	cmc := ethcashier.NewCMCClient(cmcAPIKey)

	adminPrivateKey := os.Getenv("ADMIN_WALLET_PRIV_KEY")
	if adminPrivateKey == "" {
		log.Fatal("No admin private key foudn in env")
	}

	adminWallet, err := ethcashier.ParseECDSAPrivateKeyFromHex(adminPrivateKey)
	if err != nil {
		log.Fatalf("Admin wallet parse error: %v", err)
	}
	api := ethcashier.NewAPI(db, cmc, rpc, adminWallet)
	api.SetupRoutes()

	log.Println("server up and running")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
