package postgres

import (
	"context"
	"fmt"
	"time"
)

type PGPosition interface {
	CreateUserPosition(*UserPosition) error
	DeleteUserPosition(int, string) error
	GetAllPositionsByAccountID(int) ([]*UserPosition, error)
}

type UserPosition struct {
	PositionID      string `json:"positionID"`
	AccountID       int    `json:"accountID"`
	CopyPositionID  string `json:"copyPositionID"`
	OpenedTimestamp string `json:"openedTimestamp"`
	Symbol          string `json:"symbol"`
	SymbolID        int    `json:"symbolID"`
	Volume          int    `json:"volume"`
	Side            string `json:"Side"`
	AveragePrice    string `json:"averagePrice"`
}

func (s *PostgresStore) createUserPositionTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `CREATE TABLE IF NOT EXISTS user_position (
		position_id VARCHAR(50) PRIMARY KEY,
		account_id INTEGER REFERENCES account(id), 
		copy_position_id VARCHAR(50) UNIQUE,
		opened_timestamp VARCHAR(100),
		symbol VARCHAR(25),
		symbol_id INTEGER,
		volume INTEGER,
		side VARCHAR(25),
		average_price VARCHAR(25)
	);`
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *PostgresStore) CreateUserPosition(up *UserPosition) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	INSERT INTO user_position (position_id, account_id, copy_position_id, opened_timestamp, symbol, symbol_id, volume, side, average_price)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`

	_, err := s.db.ExecContext(ctx, query, up.PositionID, up.AccountID, up.CopyPositionID, up.OpenedTimestamp, up.Symbol, up.SymbolID, up.Volume, up.Side, up.AveragePrice)
	if err != nil {
		return fmt.Errorf("error creating user position: %+v", err)
	}
	return nil
}

func (s *PostgresStore) DeleteUserPosition(accountID int, positionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	DELETE FROM user_position 
	WHERE position_id = $1
	AND account_id = $2;

	`

	result, err := s.db.ExecContext(ctx, query, positionID, accountID)
	if err != nil {
		return fmt.Errorf("error deleting user position: %+v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %+v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user position found with position_id: %s for account_id: %d", positionID, accountID)
	}
	return nil
}

func (s *PostgresStore) GetAllPositionsByAccountID(accountID int) ([]*UserPosition, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT position_id, account_id, copy_position_id, opened_timestamp,
	symbol,
	symbol_id,
	volume,
	side,
	average_price
	FROM user_position
	WHERE account_id = $1;
	`

	rows, err := s.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting user positions: %+v", err)
	}
	defer rows.Close()

	var userPositions []*UserPosition
	for rows.Next() {
		userPosition := &UserPosition{}
		err = rows.Scan(&userPosition.PositionID, &userPosition.AccountID, &userPosition.CopyPositionID, &userPosition.OpenedTimestamp, &userPosition.Symbol, &userPosition.SymbolID, &userPosition.Volume, &userPosition.Side, &userPosition.AveragePrice)
		if err != nil {
			return nil, err
		}
		userPositions = append(userPositions, userPosition)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error scanning rows: %+v", err)
	}

	return userPositions, nil
}
