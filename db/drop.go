package db

import (
	"fmt"
	"regexp"

	"github.com/jmoiron/sqlx"
)

var (
	urlExp = regexp.MustCompile(`postgres:\/\/([^:]+):([^@]+)@([^:]+):(\d+)\/([^?]+).*`)
)

// Creates the database.
func Drop(url string) error {
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

	if dbexists == 0 {
		return nil
	}

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", matches[5]))
	if err != nil {
		return fmt.Errorf("error dropping database: %w", err)
	}

	return nil
}
