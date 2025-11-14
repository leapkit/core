package db_test

import (
	"database/sql"
	"testing"

	"go.leapkit.dev/core/db"
)

func TestConnection(t *testing.T) {
	urls := []string{"test_1.db", "test_2.db", "test_3.db"}

	createDatabases := func() {
		for _, url := range urls {
			if err := db.Create(url); err != nil {
				t.Errorf("error creating db: %v", err)
			}
		}
	}

	t.Run("Correct - creating multiple connections", func(t *testing.T) {
		t.Cleanup(func() {
			for _, url := range urls {
				if err := db.Drop(url); err != nil {
					t.Errorf("error dropping db: %v", err)
				}
			}
		})

		createDatabases()

		for _, url := range urls {
			conFn := db.ConnectionFn(url, db.WithDriver("sqlite3"))

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
			for _, url := range urls {
				if err := db.Drop(url); err != nil {
					t.Errorf("error dropping db: %v", err)
				}
			}
		})

		createDatabases()

		var currentConn *sql.DB

		for _, url := range urls {
			conFn := db.ConnectionFn(url, db.WithDriver("sqlite3"))

			conn, err := conFn()
			if err != nil {
				t.Errorf("Expected nil, got err %v", err)
			}

			if url == urls[0] {
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

	t.Run("connection params", func(t *testing.T) {
		cases := []string{
			":memory:",
			"file::memory:?cache=shared",
			t.TempDir() + "memory.db",
			t.TempDir() + "memory.db?mode=memory",
		}

		for _, tcase := range cases {
			t.Run(tcase, func(t *testing.T) {
				connFn := db.ConnectionFn(
					tcase,

					db.WithDriver("sqlite3"),
					db.Params("_cache_size", "54321"), // Applying cache size parameter to check it later.
				)

				conn, err := connFn()
				if err != nil {
					t.Errorf("Expected nil, got err %v", err)
				}
				defer conn.Close()

				rows, err := conn.Query("pragma cache_size;")
				if err != nil {
					t.Errorf("Expected nil, got err %v", err)
				}
				defer rows.Close()

				var size int
				for rows.Next() {
					if err := rows.Scan(&size); err != nil {
						t.Errorf("Expected nil, got err %v", err)
					}
				}

				if size != 54321 {
					t.Fatalf("Expected cache size to be 54321, got %v", size)
				}
			})
		}
	})
}
