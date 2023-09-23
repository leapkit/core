package db

import (
	"sync"

	"github.com/jmoiron/sqlx"
)

var (
	conn *sqlx.DB
	cmux sync.Mutex

	//DriverName defaults to postgres
	driverName = "postgres"
)

// ConnFn is the database connection builder function that
// will be used by the application based on the driver and
// connection string.
type ConnFn func() (*sqlx.DB, error)

// ConnectionFn is the database connection builder function that
// will be used by the application based on the driver and
// connection string. It opens the connection only once
// and return the same connection on subsequent calls.
func ConnectionFn(url string, opts ...connectionOption) ConnFn {
	return func() (cx *sqlx.DB, err error) {
		cmux.Lock()
		defer cmux.Unlock()

		if conn != nil {
			return conn, nil
		}

		// Apply options before connecting to the database.
		for _, v := range opts {
			v()
		}

		conn, err = sqlx.Connect(driverName, url)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}
}

type connectionOption func()

func WithDriver(name string) connectionOption {
	return func() {
		driverName = name
	}
}
