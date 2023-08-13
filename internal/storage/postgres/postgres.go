package postgres

import (
	"api/config"
	"fmt"
	"log"

	"database/sql"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

type PGManager interface {
	PGAccount
	PGPosition
	PGUserData
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
	dsn := fmt.Sprintf("host=postgres user=%s password=%s dbname=%s port=5432 sslmode=disable", user, password, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	log.Println("connected to database")
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init(cleanup func()) (func(), error) {
	cleanupFunc := func() {
		cleanup()
		err := s.db.Close()
		if err != nil {
			log.Println("Error closing postgres connection:", err)
		} else {
			log.Println("Postgres connection closed successfully")
		}
	}

	err := s.createAccountTable()
	if err != nil {
		return cleanupFunc, err
	}

	err = s.createUserPositionTable()
	if err != nil {
		return cleanupFunc, err
	}

	err = s.createUserDataTable()
	if err != nil {
		return cleanupFunc, err
	}

	if err != nil {
		return func() {
			cleanup()
			err := s.db.Close()
			if err != nil {
				log.Println("Error closing postgres connection:", err)
			} else {
				log.Println("Postgres connection closed successfully")
			}
		}, err
	}

	return func() {
		cleanup()
		err := s.db.Close()
		if err != nil {
			log.Println("Error closing postgres connection:", err)
		} else {
			log.Println("Postgres connection closed successfully")
		}
	}, nil
}
