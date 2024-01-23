package db

// migrationOptions are the options for the migration
type migrationOption func() error

// UseMigrationFolder sets the folder for migrations
func UseMigrationFolder(folder string) migrationOption {
	return func() error {
		migrationsFolder = folder

		return nil
	}
}
