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
	PositionID     string `json:"positionID"`
	AccountID      int    `json:"accountID"`
	CopyPositionID string `json:"copyPositionID"`
}

func (s *PostgresStore) createUserPositionTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `CREATE TABLE IF NOT EXISTS user_position (
		position_id VARCHAR(50) PRIMARY KEY,
		account_id INTEGER REFERENCES account(id), 
		copy_position_id VARCHAR(50) UNIQUE
	);`
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *PostgresStore) CreateUserPosition(up *UserPosition) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	INSERT INTO user_position (position_id, account_id, copy_position_id)
	VALUES ($1, $2, $3);
	`

	_, err := s.db.ExecContext(ctx, query, up.PositionID, up.AccountID, up.CopyPositionID)
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

	_, err := s.db.ExecContext(ctx, query, positionID, accountID)
	if err != nil {
		return fmt.Errorf("error deleting user position: %+v", err)
	}
	return nil
}

func (s *PostgresStore) GetAllPositionsByAccountID(accountID int) ([]*UserPosition, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT position_id, account_id, copy_position_id
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
		err = rows.Scan(&userPosition.PositionID, &userPosition.AccountID, &userPosition.CopyPositionID)
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
