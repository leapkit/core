package sqlite

import "github.com/jmoiron/sqlx"

// adapter for the sqlite database it includes the connection
// to perform the framework operations.
type adapter struct {
	conn *sqlx.DB
}

// New sqlite adapter with the passed connection.
func New(conn *sqlx.DB) *adapter {
	return &adapter{
		conn: conn,
	}
}
