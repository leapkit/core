package db

import (
	"embed"
	"fmt"
	"regexp"

	"github.com/jmoiron/sqlx"
	"github.com/leapkit/core/db/migrations"
)

// RunMigrations by checking in the migrations database
// table, each of the adapters take care of this.
func RunMigrations(fs embed.FS, conn *sqlx.DB) error {
	dir, err := fs.ReadDir(".")
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %w", err)
	}

	migrator := migratorFor(conn).(migrations.Migrator)
	err = migrator.Setup()
	if err != nil {
		return fmt.Errorf("error setting up migrations: %w", err)
	}

	exp := regexp.MustCompile("(\\d{14})_(.*).sql")
	for _, v := range dir {
		if v.IsDir() {
			continue
		}

		matches := exp.FindStringSubmatch(v.Name())
		if len(matches) != 3 {
			continue
		}

		timestamp := matches[1]
		content, err := fs.ReadFile(v.Name())
		if err != nil {
			return fmt.Errorf("error opening migration file: %w", err)
		}

		err = migrator.Run(timestamp, string(content))
		if err != nil {
			return fmt.Errorf("error running migration: %w", err)
		}

		fmt.Println("✅ Migration complete:", v.Name())
	}

	return nil
}
