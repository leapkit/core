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
)

// GenerateMigration in internal/app/database/migrations/<timestamp>_<name>.go
func GenerateMigration(name string) error {
	m := migration{
		Name:      name,
		Timestamp: time.Now().Format("20060102150405"),
	}

	t, err := template.New("migration").Parse(migrationTemplate)
	if err != nil {
		return fmt.Errorf("error parsing migrations template: %w", err)
	}

	fname := filepath.Join("internal", "app", "database", "migrations", m.Filename())
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
