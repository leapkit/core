package db

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/leapkit/core/db/migrations"
)

// GenerateMigration in the migrations folder using the migrations template
func GenerateMigration(name string, options ...migrations.Option) error {
	// Applying options before generating the migration
	migrations.Apply(options...)

	m := migrations.New(name)
	t, err := template.New("migration").Parse(migrations.Folder())
	if err != nil {
		return fmt.Errorf("error parsing migrations template: %w", err)
	}

	// Destination file name
	name = filepath.Join(migrations.Folder(), m.Filename())
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("error creating migration file: %w", err)
	}

	err = t.ExecuteTemplate(f, "migration", m)
	if err != nil {
		return fmt.Errorf("error executing migrations template: %w", err)
	}

	fmt.Printf("✅ Migration file `%v` generated\n", name)
	return nil
}
