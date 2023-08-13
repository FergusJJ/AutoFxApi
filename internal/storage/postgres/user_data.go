package postgres

import (
	"context"
	"fmt"
	"time"
)

type PGUserData interface {
	UpsertUserData(*UserData) error
	UpdateUserData(*UserData) error
	GetUserDataByAccountID(int) (*UserData, error)
	DeleteUserData(int) error
}

type UserData struct {
	DataID             int    `json:"dataID"`    //handled by postgres SERIAL tag
	AccountID          int    `json:"accountID"` //ref account(id)
	RefreshToken       string `json:"refreshToken"`
	RefreshTokenExpiry int64  `json:"refreshTokenExpiry"`
}

func (s *PostgresStore) createUserDataTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `CREATE TABLE IF NOT EXISTS user_data (
		data_id SERIAL PRIMARY KEY,
		account_id INTEGER REFERENCES account(id) NOT NULL UNIQUE, 
		refresh_token TEXT NOT NULL,
		refresh_token_expiry INT NOT NULL
	);`
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *PostgresStore) UpsertUserData(data *UserData) error {
	query := `
	INSERT INTO user_data (account_id, refresh_token, refresh_token_expiry)
	VALUES ($1, $2, $3)
	ON CONFLICT (account_id)
	DO UPDATE SET refresh_token = $2, refresh_token_expiry = $3
	RETURNING data_id;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, data.AccountID, data.RefreshToken, data.RefreshTokenExpiry).Scan(&data.DataID)
	if err != nil {
		return fmt.Errorf("error creating user data: %+v", err)
	}
	return nil
}

func (s *PostgresStore) UpdateUserData(data *UserData) error {
	query := `
	UPDATE user_data
	SET account_id = $1, refresh_token = $2, refresh_token_expiry = $3
	WHERE data_id = $4;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query, data.AccountID, data.RefreshToken, data.RefreshTokenExpiry, data.DataID)
	if err != nil {
		return fmt.Errorf("error updating user data: %+v", err)
	}
	return nil
}

func (s *PostgresStore) GetUserDataByAccountID(accountID int) (*UserData, error) {
	query := `
	SELECT data_id, account_id, refresh_token, refresh_token_expiry
	FROM user_data
	WHERE account_id = $1;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, query, accountID)
	data := &UserData{}
	err := row.Scan(&data.DataID, &data.AccountID, &data.RefreshToken, &data.RefreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("error getting user data: %+v", err)
	}
	return data, nil
}

func (s *PostgresStore) DeleteUserData(dataID int) error {
	query := `
	DELETE FROM user_data
	WHERE data_id = $1;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query, dataID)
	if err != nil {
		return fmt.Errorf("error deleting user data: %+v", err)
	}
	return nil
}

// func (s *)
