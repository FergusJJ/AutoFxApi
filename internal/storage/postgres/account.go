package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type PGAccount interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountByID(int) (*Account, error)
	CheckRelation(int, string) (bool, error)
	GetAccountID(string) (int, error)
}

type Account struct {
	ID        int       `json:"ID"`        //handled by postgres SERIAL tag
	License   string    `json:"license"`   //license key
	Email     string    `json:"email"`     //email of the user provided by whop
	CreatedAt time.Time `json:"createdAt"` //time that user first uses app, not when buys keys on whop
}

func (s *PostgresStore) createAccountTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `CREATE TABLE IF NOT EXISTS account (
		id SERIAL PRIMARY KEY, 
		license VARCHAR(50) NOT NULL UNIQUE,
		email VARCHAR(255) NOT NULL UNIQUE,
		created_at TIMESTAMP NOT NULL
	);`
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *PostgresStore) CreateAccount(account *Account) error {
	query := `
	INSERT INTO account (license, email, created_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (license)
	DO NOTHING
	RETURNING id;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.db.QueryRowContext(ctx, query, account.License, account.Email, account.CreatedAt).Scan(&account.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return fmt.Errorf("error creating account: %+v", err)
	}
	return nil
}

func (s *PostgresStore) DeleteAccount(accountID int) error {
	query := `DELETE FROM account WHERE id = $1;`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query, accountID)
	if err != nil {
		return fmt.Errorf("error deleting account: %+v", err)
	}
	return nil

}

func (s *PostgresStore) UpdateAccount(account *Account) error {
	query := `
	UPDATE account
	SET license = $1, email = $2
	WHERE id = $3;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query, account.License, account.Email, account.ID)
	if err != nil {
		return fmt.Errorf("error updating account: %+v", err)
	}
	return nil
}

func (s *PostgresStore) GetAccountByID(accountID int) (*Account, error) {
	var account = &Account{}
	log.Println(accountID)
	query := `
	SELECT id, license, email, created_at
	FROM account
	WHERE id = $1;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.db.QueryRowContext(ctx, query, accountID).Scan(&account.ID, &account.License, &account.Email, &account.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error fetching account by id: %+v", err)
	}
	return account, nil
}

func (s *PostgresStore) GetAccountID(license string) (int, error) {
	var id int
	query := `
	SELECT id
	FROM account
	WHERE license = $1;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.db.QueryRowContext(ctx, query, license).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("error fetching account by id: %+v", err)
	}
	return id, nil
}

func (s *PostgresStore) CheckRelation(accountID int, license string) (bool, error) {
	query := `
	SELECT COUNT(*)
	FROM account
	WHERE id = $1 AND license = $2;
	`
	var count int
	err := s.db.QueryRow(query, accountID, license).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking license connection: %+v", err)
	}
	return count > 0, nil
}
