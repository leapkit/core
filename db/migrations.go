package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Migrations are expected to follow the naming convention:
// YYYYMMDDHHMMSS_description.sql (e.g. 20220101120000_create_users_table.sql)
var migrationExp = regexp.MustCompile(`(\d{14})_(.*).sql`)

// RunMigrationsDir receives a folder and a database URL
// to apply the migrations to the database.
func RunMigrationsDir(dir string, conn *sql.DB) error {
	migrator := NewMigrator(conn)
	err := migrator.Setup()
	if err != nil {
		return fmt.Errorf("error setting up migrations: %w", err)
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking migrations directory: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		return process(migrator, path, os.ReadFile)
	})
}

// RunMigrations by checking in the migrations database
// table, each of the adapters take care of this.
func RunMigrations(fs embed.FS, conn *sql.DB) error {
	dir, err := fs.ReadDir(".")
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %w", err)
	}

	migrator := NewMigrator(conn)
	err = migrator.Setup()
	if err != nil {
		return fmt.Errorf("error setting up migrations: %w", err)
	}

	for _, v := range dir {
		if v.IsDir() {
			continue
		}

		err := process(migrator, v.Name(), fs.ReadFile)
		if err != nil {
			return fmt.Errorf("error processing migration %s: %w", v.Name(), err)
		}
	}

	return nil
}

// process applies a single migration file to the database.
// It extracts the timestamp and name from the filename using regex,
// reads the file content using the provided file reader function,
// and runs the migration using the migrator.
//
// Parameters:
//   - migrator: The migrator that handles applying the migration
//   - filename: The migration file name (should follow YYYYMMDDHHMMSS_description.sql format)
//   - fileReadFn: A function that reads file content (allows supporting both fs.FS and os.File)
//
// Returns:
//   - An error if the migration fails, nil on success
//   - Returns nil silently for files that don't match the migration naming pattern
func process(migrator *Migrator, filename string, fileReadFn func(string) ([]byte, error)) error {
	matches := migrationExp.FindStringSubmatch(filepath.Base(filename))
	if len(matches) != 3 {
		return nil
	}

	content, err := fileReadFn(filename)
	if err != nil {
		return fmt.Errorf("error opening migration file: %w", err)
	}

	err = migrator.Run(matches[1], matches[2], string(content))
	if err != nil {
		return fmt.Errorf("error running migration: %w", err)
	}

	return nil
}
