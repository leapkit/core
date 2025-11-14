package db

import (
	"database/sql"
	"net/url"
	"strings"
	"sync"
)

var (
	dbPool = map[string]*sql.DB{}
	cmux   sync.Mutex

	// DriverName defaults to postgres
	driverName = "postgres"

	// Connection params to be appended to the connection string
	connParams string
)

// ConnFn is the database connection builder function that
// will be used by the application based on the driver and
// connection string.
type ConnFn func() (*sql.DB, error)

// connectionOptions for the database
type connectionOption func()

// ConnectionFn is the database connection builder function that
// will be used by the application based on the driver and
// connection string. It opens the connection only once
// and return the same connection on subsequent calls.
func ConnectionFn(url string, opts ...connectionOption) ConnFn {
	return func() (cx *sql.DB, err error) {
		cmux.Lock()
		defer cmux.Unlock()

		if conn := dbPool[url]; conn != nil && conn.Ping() == nil {
			return conn, nil
		}

		// Apply options before connecting to the database.
		for _, v := range opts {
			v()
		}

		// Modify the URL to include connection params
		// if any.
		modURL := url
		if strings.Contains(modURL, "?") {
			modURL = modURL + "&" + connParams
		} else if connParams != "" {
			modURL = modURL + "?" + connParams
		}

		conn, err := sql.Open(driverName, modURL)
		if err != nil {
			return nil, err
		}

		// This uses url instead of modURL while the params apply
		// to all connections.
		dbPool[url] = conn

		return conn, nil
	}
}

// WithDriver allows to specify the driver to use driver defaults to
// postgres.
func WithDriver(name string) connectionOption {
	return func() {
		driverName = name
	}
}

// Params allows to specify additional connection parameters
// that will be encoded as URL params next to the connection string.
// params should be in key,value,key,value,... format.
// e.g Params("sslmode", "disable", "timezone", "UTC")
func Params(params ...string) connectionOption {
	vals := url.Values{}
	for i := 0; i < len(params); i += 2 {
		key := params[i]
		if i+1 >= len(params) {
			vals.Add(key, "")
			break
		}

		vals.Add(key, params[i+1])
	}

	return func() {
		connParams = vals.Encode()
	}
}
