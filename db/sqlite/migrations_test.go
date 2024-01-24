package sqlite

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestSetup(t *testing.T) {
	td := t.TempDir()
	conn, err := sqlx.Connect("sqlite3", filepath.Join(td, "database.db"))
	if err != nil {
		t.Fatal(err)
	}

	adapter := New(conn)
	err = adapter.Setup()
	if err != nil {
		t.Fatal(err)
	}

	result := struct{ Name string }{}
	err = conn.Get(&result, "SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations';")
	if err != nil {
		t.Fatal("schema_migrations table not found")
	}

	if result.Name != "schema_migrations" {
		t.Fatal("schema_migrations table not found")
	}
}

func TestRun(t *testing.T) {
	t.Run("migration not found", func(t *testing.T) {
		td := t.TempDir()
		conn, err := sqlx.Connect("sqlite3", filepath.Join(td, "database.db"))
		if err != nil {
			t.Fatal(err)
		}

		adapter := New(conn)
		err = adapter.Setup()
		if err != nil {
			t.Fatal(err)
		}

		err = adapter.Run("20210101000000", "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);")
		if err != nil {
			t.Fatal(err)
		}

		result := struct{ Name string }{}
		err = conn.Get(&result, "SELECT name FROM sqlite_master WHERE type='table' AND name='users';")
		if err != nil {
			t.Fatal("users table not found")
		}
	})

	t.Run("migration found", func(t *testing.T) {
		td := t.TempDir()
		conn, err := sqlx.Connect("sqlite3", filepath.Join(td, "database.db"))
		if err != nil {
			t.Fatal(err)
		}

		adapter := New(conn)
		err = adapter.Setup()
		if err != nil {
			t.Fatal(err)
		}

		err = adapter.Run("20210101000000", "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);")
		if err != nil {
			t.Fatal(err)
		}

		err = adapter.Run("20210101000000", "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);")
		if err != nil {
			t.Fatal(err)
		}

		result := struct{ Name string }{}
		err = conn.Get(&result, "SELECT name FROM sqlite_master WHERE type='table' AND name='users';")
		if err != nil {
			t.Fatal("users table not found")
		}
	})
}
