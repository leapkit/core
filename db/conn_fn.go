package db

import (
	"database/sql"
	"sync"
)

var (
	dbPool = map[string]*sql.DB{}
	cmux   sync.Mutex
)

// ConnFn is the database connection builder function that
// will be used by the application based on the driver and
// connection string.
type ConnFn func() (*sql.DB, error)

// ConnectionFn is the database connection builder function that
// will be used by the application based on the driver and
// connection string. It opens the connection only once
// and return the same connection on subsequent calls.
func ConnectionFn(url string, opts ...connectionOption) ConnFn {
	return func() (cx *sql.DB, err error) {
		cmux.Lock()
		defer cmux.Unlock()

		// Return existing connection if available and valid
		// to avoid reopening connections. Otherwise continue
		// to create a new one.
		if conn := dbPool[url]; conn != nil && conn.Ping() == nil {
			return conn, nil
		}

		cs := &connSettings{
			url: url,

			driver: "postgres",
			params: "",
		}

		// Apply options before connecting to the database.
		for _, v := range opts {
			v(cs)
		}

		conn, err := sql.Open(cs.driver, cs.connectionURL())
		if err != nil {
			return nil, err
		}

		// This uses url instead of modURL while the params apply
		// to all connections.
		dbPool[url] = conn

		return conn, nil
	}
}
