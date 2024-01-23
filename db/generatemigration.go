package db

import (
	_ "embed"

	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"
)

var (
	//go:embed migration.tmpl
	migrationTemplate string

	// migrationsFolder is the base folder for migrations
	migrationsFolder = filepath.Join("internal", "app", "database", "migrations")
)

// GenerateMigration in the migrations folder using the migrations template
func GenerateMigration(name string, options ...migrationOption) error {
	m := migration{
		Name:      name,
		Timestamp: time.Now().Format("20060102150405"),
	}

	// applying specified options
	for _, option := range options {
		if err := option(); err != nil {
			return fmt.Errorf("error applying migration option: %w", err)
		}
	}

	t, err := template.New("migration").Parse(migrationTemplate)
	if err != nil {
		return fmt.Errorf("error parsing migrations template: %w", err)
	}

	fname := filepath.Join(migrationsFolder, m.Filename())
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("error creating migration file: %w", err)
	}

	err = t.ExecuteTemplate(f, "migration", m)
	if err != nil {
		return fmt.Errorf("error executing migrations template: %w", err)
	}

	fmt.Printf("✅ Migration file `%v` generated\n", fname)

	return nil
}
