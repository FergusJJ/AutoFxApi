package storage

import (
	"api/config"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Account struct {
	gorm.Model
	UserID uint `json:"userID" gorm:"primaryKey"`
}

type PGStorage interface {
	CreateAccount() error
	DeleteAccount(string) error
	UpdateAccount() error
	GetAccountByID(string) error
}

type PostgresStore struct {
	db *gorm.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	password, err := config.Config("PG_PASSWORD")
	if err != nil {
		return nil, err
	}
	dbname, err := config.Config("PG_DB")
	if err != nil {
		return nil, err
	}
	user, err := config.Config("PG_USER")
	if err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("host=db user=%s password=%s dbname=%s port=5432 sslmode=disable", user, password, dbname)
	db, err := gorm.Open(postgres.Open(dsn),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		},
	)
	if err != nil {
		return nil, err
	}
	log.Println("connected to database")

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return nil
}

func (s *PostgresStore) createAccountTable() error {
	return nil
}

func (s *PostgresStore) CreateAccount() error {
	return nil
}

func (s *PostgresStore) DeleteAccount(accountId int) error {
	return nil

}

func (s *PostgresStore) UpdateAccount() error {
	return nil

}

func (s *PostgresStore) GetAccountByID(accountId int) error {
	return nil

}
