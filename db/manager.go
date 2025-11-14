// Package db provides database connection management, migration system,
// and database creation utilities with support for SQLite and PostgreSQL.
package db

import (
	"cmp"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// postgresURLExp is the regular expression to extract the database name
// and the user credentials from the database URL.
var postgresURLExp = regexp.MustCompile(`postgres://(.*):(.*)@(.*):(.*)/([^?]*)`)

// Create a new database based on the passed URL.
func Create(url string) error {
	createFn := createSQLite
	if strings.Contains(url, "postgres") {
		createFn = createPostgres
	}

	return createFn(url)
}

func createSQLite(conURL string) error {
	u, err := url.Parse(conURL)
	if err != nil {
		return fmt.Errorf("parsing database URL: %w", err)
	}

	_, err = os.Create(u.Path)
	if err != nil {
		return fmt.Errorf("creating database file %s: %w", u.Path, err)
	}

	return nil
}

func createPostgres(conURL string) error {
	matches := postgresURLExp.FindStringSubmatch(conURL)
	if len(matches) != 6 {
		return fmt.Errorf("invalid database URL: %s", conURL)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		matches[1], matches[2], matches[3], matches[4],
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	var exists int
	row := db.QueryRow("SELECT COUNT(datname) FROM pg_database WHERE datname ilike $1", matches[5])
	err = row.Scan(&exists)
	if err != nil {
		return err
	}

	if exists == 1 {
		return nil
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", matches[5]))
	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}

	return nil
}

// Drop a database based on the passed URL.
func Drop(url string) error {
	dropFn := dropSQLite
	if strings.Contains(url, "postgres") {
		dropFn = dropPostgres
	}

	return dropFn(url)
}

func dropSQLite(conURL string) error {
	u, err := url.Parse(conURL)
	if err != nil {
		return fmt.Errorf("parsing database URL: %w", err)
	}

	err = os.Remove(u.Path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("dropping database: %w", err)
	}

	return nil
}

func dropPostgres(conURL string) error {
	matches := postgresURLExp.FindStringSubmatch(conURL)
	if len(matches) != 3 || matches[1] == "" {
		return fmt.Errorf("invalid database URL: %s", conURL)
	}

	db, err := sql.Open("postgres", matches[1])
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	dbName := cmp.Or(matches[2], "postgres")

	var dbexists int
	row := db.QueryRow("SELECT COUNT(datname) FROM pg_database WHERE datname ilike $1", dbName)
	err = row.Scan(&dbexists)
	if err != nil {
		return err
	}

	if dbexists == 0 {
		return nil
	}

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", dbName))
	if err != nil {
		return fmt.Errorf("dropping database: %w", err)
	}

	return nil
}
