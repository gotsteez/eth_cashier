package ethcashier

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
)

// API struct to hold shared resources
type API struct {
	db              *DB
	cmc             *CMCClient
	rpc             *RPCClient
	adminPrivateKey *ecdsa.PrivateKey
}

// NewAPI creates a new instance of the API
func NewAPI(db *DB, cmc *CMCClient, rpc *RPCClient, adminPrivateKey *ecdsa.PrivateKey) *API {
	return &API{
		db:              db,
		cmc:             cmc,
		rpc:             rpc,
		adminPrivateKey: adminPrivateKey,
	}
}

type UserRequest struct {
	User string `json:"user"`
}

type NewUserResponse struct {
	User            string `json:"user"`
	WalletPublicKey string `json:"walletPublicKey"`
}

type BalanceResponse struct {
	Balance float64 `json:"balance"`
}

type UserResponse struct {
	User            string  `json:"user"`
	Balance         float64 `json:"balance"`
	WalletPublicKey string  `json:"walletPublicKey"`
}

// HandleNewUser creates a new user and returns their ID
func (api *API) HandleNewUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create new user
	user := NewUser()
	err := api.db.CreateUser(user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response := NewUserResponse{
		User:            user.ID,
		WalletPublicKey: user.Wallet.PublicKey,
	}

	json.NewEncoder(w).Encode(response)
}

type CheckRequest struct {
	User string `json:"user"`
}

func (api *API) Check(user *User) (float64, error) {
	privateKey, err := ParseECDSAPrivateKeyFromHex(user.Wallet.EncryptedPrivateKey)
	if err != nil {
		return 0, fmt.Errorf("failed to parse private key: %v", err)
	}

	balance, err := api.rpc.GetBalance(user.Wallet.PublicKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get wallet balance: %v", err)
	}

	// If balance is 0, return early
	if balance.Cmp(big.NewInt(0)) == 0 {
		return 0, fmt.Errorf("wallet has no ETH balance")
	}

	// Get admin's public key from private key
	adminPublicKey := api.adminPrivateKey.Public()
	adminPublicKeyECDSA, ok := adminPublicKey.(*ecdsa.PublicKey)
	if !ok {
		return 0, fmt.Errorf("failed to get admin public key")
	}
	adminAddress := crypto.PubkeyToAddress(*adminPublicKeyECDSA).Hex()

	// 3. Send entire balance to admin wallet
	// Subtract a small amount for gas (0.001 ETH)
	gasReserve := big.NewInt(1000000000000000) // 0.001 ETH in Wei
	transferAmount := new(big.Int).Sub(balance, gasReserve)

	err = api.rpc.Send(privateKey, adminAddress, transferAmount)
	if err != nil {
		return 0, fmt.Errorf("failed to send ETH to admin wallet: %v", err)
	}

	// 4. Get current ETH price
	ethPrice, err := api.cmc.GetEthereumPrice()
	if err != nil {
		return 0, fmt.Errorf("failed to get ETH price: %v", err)
	}

	// Convert Wei to ETH and calculate USD value
	ethAmount := new(big.Float).Quo(
		new(big.Float).SetInt(transferAmount),
		new(big.Float).SetInt(big.NewInt(1000000000000000000)), // 10^18 (Wei to ETH)
	)
	var ethFloat64 float64
	ethFloat64, _ = ethAmount.Float64()
	usdValue := ethFloat64 * ethPrice

	// 5. Credit the user's balance
	err = api.db.AddToBalance(user.ID, usdValue)
	if err != nil {
		return 0, fmt.Errorf("failed to credit user balance: %v", err)
	}

	// 6. Get and return updated balance
	updatedUser, err := api.db.GetUser(user.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get updated balance: %v", err)
	}

	return updatedUser.Balance, nil
}

// HandleCheck checks the user's current balance
func (api *API) HandleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := api.db.GetUser(req.User)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	newBalance, err := api.Check(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking balance: %v", err), http.StatusInternalServerError)
		return
	}
	response := BalanceResponse{
		Balance: newBalance,
	}

	json.NewEncoder(w).Encode(response)
}

type WithdrawRequest struct {
	User   string  `json:"user"`
	Wallet string  `json:"wallet"` // wallet to send the money to
	Amount float64 `json:"amount"`
}

// Withdraw sends money back to the user
func (api *API) Withdraw(user *User, amount float64, userAddress string) (float64, error) {
	if err := api.db.SubtractFromBalance(user.ID, amount); err != nil {
		return 0, fmt.Errorf("Unable to subtract from balance: %v", err)
	}

	ethPrice, err := api.cmc.GetEthereumPrice()
	if err != nil {
		// If we fail here, we should add the amount back to user's balance
		api.db.AddToBalance(user.ID, amount)
		return 0, fmt.Errorf("failed to get ETH price: %v", err)
	}

	// 4. Convert USD to Wei (multiply by 10^18)
	ethAmount := amount / ethPrice
	weiAmount := new(big.Int).Mul(
		new(big.Int).SetInt64(int64(ethAmount*1000000)), // Convert to integer (multiplied by 10^6 for precision)
		new(big.Int).SetInt64(1000000000000),            // Multiply by 10^12 to get to Wei (10^6 * 10^12 = 10^18)
	)

	// 5. Send the ETH to the user's address
	if err := api.rpc.Send(api.adminPrivateKey, userAddress, weiAmount); err != nil {
		// If the transfer fails, add the amount back to user's balance
		api.db.AddToBalance(user.ID, amount)
		return 0, fmt.Errorf("failed to send ETH: %v", err)
	}

	// 6. Get and return the updated balance
	updatedUser, err := api.db.GetUser(user.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get updated balance: %v", err)
	}
	return updatedUser.Balance, nil
}

// HandleWithdraw processes a withdrawal request
func (api *API) HandleWithdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get updated user info
	user, err := api.db.GetUser(req.User)
	if err != nil {
		http.Error(w, "Failed to find user", http.StatusInternalServerError)
		return
	}

	newBalance, err := api.Withdraw(user, req.Amount, req.Wallet)
	if err != nil {
		http.Error(w, "Failed to withdraw balance", http.StatusInternalServerError)
		return
	}
	response := BalanceResponse{
		Balance: newBalance,
	}

	json.NewEncoder(w).Encode(response)
}

// HandleGetUser gets information about a specific user
func (api *API) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := api.db.GetUser(req.User)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := UserResponse{
		User:            user.ID,
		Balance:         user.Balance,
		WalletPublicKey: user.Wallet.PublicKey,
	}

	json.NewEncoder(w).Encode(response)
}

// SetupRoutes configures the HTTP routes
func (api *API) SetupRoutes() {
	http.HandleFunc("/newUser", api.HandleNewUser)
	http.HandleFunc("/check", api.HandleCheck)
	http.HandleFunc("/withdraw", api.HandleWithdraw)
	http.HandleFunc("/user", api.HandleGetUser)
}
