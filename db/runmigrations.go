package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"path"
	"regexp"

	"github.com/jmoiron/sqlx"
)

var (
	//go:embed schemamigrations.sql
	migrationsTableStatement string
)

func RunMigrations(fs embed.FS, conn *sqlx.DB) error {
	dir, err := fs.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %w", err)
	}

	_, err = conn.Exec(migrationsTableStatement)
	if err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	for _, v := range dir {
		if v.IsDir() {
			continue
		}

		exp := regexp.MustCompile("(\\d{14})_(.*).sql")
		matches := exp.FindStringSubmatch(v.Name())
		if len(matches) != 3 {
			continue
		}

		timestamp := matches[1]
		err := conn.Get(&timestamp, "SELECT * FROM schema_migrations WHERE timestamp = $1", timestamp)
		if err == nil {
			continue
		}

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		content, err := fs.ReadFile(path.Join("migrations", v.Name()))
		if err != nil {
			return fmt.Errorf("error opening migration file: %w", err)
		}

		_, err = conn.Exec(string(content))
		if err != nil {
			return fmt.Errorf("error executing migration: %w", err)
		}

		_, err = conn.Exec("INSERT INTO schema_migrations (timestamp) VALUES ($1)", timestamp)
		if err != nil {
			return fmt.Errorf("error inserting migration into schema_migrations: %w", err)
		}

		fmt.Println("✅ Migration complete:", v.Name())
	}

	return nil
}
