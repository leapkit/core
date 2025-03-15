package db_test

import (
	"database/sql"
	"testing"

	"go.leapkit.dev/core/db"
)

func TestConnection(t *testing.T) {
	urls := []string{"test_1.db", "test_2.db", "test_3.db"}

	createDatabases := func() {
		for _, dbUrl := range urls {
			if err := db.Create(dbUrl); err != nil {
				t.Errorf("error creating db: %v", err)
			}
		}

	}
	t.Run("Correct - creating multiple connections", func(t *testing.T) {
		t.Cleanup(func() {
			for _, dbUrl := range urls {
				if err := db.Drop(dbUrl); err != nil {
					t.Errorf("error dropping db: %v", err)
				}
			}
		})

		createDatabases()

		for _, dbUrl := range urls {
			conFn := db.ConnectionFn(dbUrl, db.WithDriver("sqlite3"))

			conn, err := conFn()
			if err != nil {
				t.Errorf("Expected nil, got err %v", err)
			}

			if err := conn.Ping(); err != nil {
				t.Errorf("Expected nil, got err %v", err)
			}
		}
	})
	t.Run("Correct - using current exiting connection if already created", func(t *testing.T) {
		t.Cleanup(func() {
			for _, dbUrl := range urls {
				if err := db.Drop(dbUrl); err != nil {
					t.Errorf("error dropping db: %v", err)
				}
			}
		})

		createDatabases()

		var currentConn *sql.DB

		for _, dbUrl := range urls {
			conFn := db.ConnectionFn(dbUrl, db.WithDriver("sqlite3"))

			conn, err := conFn()
			if err != nil {
				t.Errorf("Expected nil, got err %v", err)
			}

			if dbUrl == urls[0] {
				currentConn = conn
			}

			if err := conn.Ping(); err != nil {
				t.Errorf("Expected nil, got err %v", err)
			}
		}

		connFn := db.ConnectionFn(urls[0], db.WithDriver("sqlite3"))

		conn, err := connFn()
		if err != nil {
			t.Errorf("Expected nil, got err %v", err)
		}

		if currentConn != conn {
			t.Errorf("Expected current connection to be %v, got %v", currentConn, conn)
		}
	})
}
