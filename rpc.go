package ethcashier

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func ParseECDSAPrivateKeyFromHex(hexString string) (*ecdsa.PrivateKey, error) {
	// Decode hex string
	b, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, err
	}

	// Use crypto.ToECDSA from go-ethereum
	privateKey, err := crypto.ToECDSA(b)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// Client struct to hold the RPC client
type RPCClient struct {
	rpcURL string
	client *ethclient.Client
}

// NewRPCClient creates a new instance of Client
func NewRPCClient(rpcURL string) (*RPCClient, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %v", err)
	}

	return &RPCClient{
		rpcURL: rpcURL,
		client: client,
	}, nil
}

// GetBalance returns the balance of the given address
func (c *RPCClient) GetBalance(address string) (*big.Int, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid address format")
	}

	account := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	return balance, nil
}

func (c *RPCClient) Send(from *ecdsa.PrivateKey, to string, amount *big.Int) error {
	ctx := context.Background()

	// Get the public address from the private key
	publicKey := from.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Validate recipient address
	if !common.IsHexAddress(to) {
		return fmt.Errorf("invalid recipient address format")
	}
	toAddress := common.HexToAddress(to)

	// Get the sender's nonce
	nonce, err := c.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}

	// Get current gas price
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gas price: %v", err)
	}

	// Create transaction data
	tx := types.NewTransaction(
		nonce,
		toAddress,
		amount,
		21000, // Standard gas limit for ETH transfers
		gasPrice,
		nil, // No data for simple transfers
	)

	// Get the chain ID
	chainID, err := c.client.NetworkID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain id: %v", err)
	}

	// Calculate total cost (amount + gas)
	gasCost := new(big.Int).Mul(gasPrice, big.NewInt(21000))
	totalCost := new(big.Int).Add(amount, gasCost)

	// Check if sender has sufficient balance
	balance, err := c.GetBalance(fromAddress.Hex())
	if err != nil {
		return fmt.Errorf("failed to get sender balance: %v", err)
	}

	if balance.Cmp(totalCost) < 0 {
		return fmt.Errorf("insufficient funds for transfer: need %v but got %v", totalCost, balance)
	}

	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), from)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Send the transaction
	err = c.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	return nil
}
