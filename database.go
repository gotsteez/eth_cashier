package ethcashier

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

// Custom errors
var (
	ErrInsufficientFunds = errors.New("insufficient funds for withdrawal")
	ErrNegativeAmount    = errors.New("amount cannot be negative")
)

func InitDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	err = createTables(db)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func createTables(db *sql.DB) error {
	userTable := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        encrypted_private_key TEXT,
        public_key TEXT,
        balance REAL
    );`

	_, err := db.Exec(userTable)
	return err
}

func (db *DB) CreateUser(user *User) error {
	query := `
    INSERT INTO users (id, encrypted_private_key, public_key, balance)
    VALUES (?, ?, ?, ?)`

	_, err := db.Exec(query,
		user.ID,
		user.Wallet.EncryptedPrivateKey,
		user.Wallet.PublicKey,
		user.Balance)
	return err
}

func (db *DB) GetUser(id string) (*User, error) {
	user := &User{}
	query := `
    SELECT id, encrypted_private_key, public_key, balance
    FROM users WHERE id = ?`

	row := db.QueryRow(query, id)
	err := row.Scan(
		&user.ID,
		&user.Wallet.EncryptedPrivateKey,
		&user.Wallet.PublicKey,
		&user.Balance)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// AddToBalance adds the specified amount to user's balance
func (db *DB) AddToBalance(id string, amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current balance
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = ?", id).Scan(&currentBalance)
	if err != nil {
		return err
	}

	// Update balance
	newBalance := currentBalance + amount
	_, err = tx.Exec("UPDATE users SET balance = ? WHERE id = ?", newBalance, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// SubtractFromBalance subtracts the specified amount from user's balance
func (db *DB) SubtractFromBalance(id string, amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current balance
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = ?", id).Scan(&currentBalance)
	if err != nil {
		return err
	}

	// Check if there are sufficient funds
	if currentBalance < amount {
		return ErrInsufficientFunds
	}

	// Update balance
	newBalance := currentBalance - amount
	_, err = tx.Exec("UPDATE users SET balance = ? WHERE id = ?", newBalance, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

func (db *DB) ListUsers() ([]User, error) {
	query := `
    SELECT id, encrypted_private_key, public_key, balance
    FROM users`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Wallet.EncryptedPrivateKey,
			&user.Wallet.PublicKey,
			&user.Balance)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
