package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type manager struct {
	url string
}

func NewManager(url string) *manager {
	return &manager{
		url: url,
	}
}

// Create postgres database based on the URL.
func (m *manager) Create() error {
	matches := urlExp.FindStringSubmatch(m.url)
	if len(matches) != 6 {
		return fmt.Errorf("invalid database url: %s", m.url)
	}

	db, err := sqlx.Connect("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", matches[1], matches[2], matches[3], matches[4]))
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	var exists int
	err = db.Get(&exists, "SELECT COUNT(datname) FROM pg_database WHERE datname ilike $1", matches[5])
	if err != nil {
		return err
	}

	if exists == 1 {
		return nil
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", matches[5]))
	if err != nil {
		return fmt.Errorf("error creating database: %w", err)
	}

	return nil
}

// Drop postgres database based on the URL.
func (m *manager) Drop() error {
	matches := urlExp.FindStringSubmatch(m.url)
	if len(matches) != 6 {
		return fmt.Errorf("invalid database url: %s", m.url)
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
