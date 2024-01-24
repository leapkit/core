package db

import (
	"github.com/jmoiron/sqlx"
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
	case "sqlite":
		return sqlite.New(conn)
	default:
		return nil
	}
}
