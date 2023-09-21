package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Creates the database.
func Create(url string) error {
	matches := urlExp.FindStringSubmatch(url)
	if len(matches) != 6 {
		return fmt.Errorf("invalid database url: %s", url)
	}

	db, err := sqlx.Connect("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", matches[1], matches[2], matches[3], matches[4]))
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	var dbexists int
	err = db.Get(&dbexists, "SELECT COUNT(datname) FROM pg_database WHERE datname ilike $1", matches[5])
	if err != nil {
		return err
	}

	if dbexists == 1 {
		return nil
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", matches[5]))
	if err != nil {
		return fmt.Errorf("error creating database: %w", err)
	}

	return nil
}
