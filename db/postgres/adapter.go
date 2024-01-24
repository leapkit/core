package postgres

import (
	"regexp"

	"github.com/jmoiron/sqlx"
)

var (
	// urlExp is the regular expression to extract the database name
	// and the user credentials from the database URL.
	urlExp = regexp.MustCompile(`postgres:\/\/([^:]+):([^@]+)@([^:]+):(\d+)\/([^?]+).*`)
)

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
