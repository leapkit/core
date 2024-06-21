package db

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jmoiron/sqlx"
	"github.com/leapkit/core/db/migrations"
	"github.com/leapkit/core/db/postgres"
	"github.com/leapkit/core/db/sqlite"
)

// migratorFor the adapter for the passed SQL connection
// based on the driver name.
func migratorFor(conn *sqlx.DB) any {
	// Migrator for the passed SQL connection.
	switch conn.DriverName() {
	case "postgres":
		return postgres.New(conn)
	case "sqlite", "sqlite3":
		return sqlite.New(conn)
	default:
		return nil
	}
}

// GenerateMigration in the migrations folder using the migrations template
func GenerateMigration(name string, options ...migrations.Option) error {
	// Applying options before generating the migration
	migrations.Apply(options...)

	m := migrations.New(name)
	t, err := template.New("migration").Parse(migrations.Template())
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

// RunMigrationsDir receives a folder and a database URL
// to apply the migrations to the database.
func RunMigrationsDir(dir, url string) error {
	conn, err := sqlx.Open("sqlite3", url)
	if err != nil {
		return fmt.Errorf("error opening database connection: %w", err)
	}

	migrator := migratorFor(conn).(migrations.Migrator)
	err = migrator.Setup()
	if err != nil {
		return fmt.Errorf("error setting up migrations: %w", err)
	}

	exp := regexp.MustCompile("(\\d{14})_(.*).sql")
	return os.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		matches := exp.FindStringSubmatch(path)
		if len(matches) != 3 {
			return nil
		}

		timestamp := matches[1]
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error opening migration file: %w", err)
		}

		err = migrator.Run(timestamp, string(content))
		if err != nil {
			return fmt.Errorf("error running migration: %w", err)
		}

		return nil
	})

}

// RunMigrations by checking in the migrations database
// table, each of the adapters take care of this.
func RunMigrations(folder embed.FS, conn *sqlx.DB) error {
	dir, err := folder.ReadDir(".")
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
	}

	return nil
}
