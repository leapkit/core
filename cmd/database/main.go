package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/leapkit/core/db"

	// Load environment variables
	_ "github.com/leapkit/core/tools/envload"
	// sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
	// postgres driver
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tools database <command>")
		fmt.Println("Available commands:")
		fmt.Println(" - migrate")
		fmt.Println(" - create")
		fmt.Println(" - drop")

		return
	}

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Println("[error] DATABASE_URL is not set")

		return
	}

	switch os.Args[1] {
	case "migrate":
		conn, err := db.Connect(url)
		if err != nil {
			fmt.Println(err)
		}

		err = db.RunMigrationsDir(filepath.Join("internal", "migrations"), conn)
		if err != nil {
			fmt.Println(err)

			return
		}

		fmt.Println("✅ Migrations ran successfully")
	case "create":
		err := db.Create(url)
		if err != nil {
			fmt.Println(err)

			return
		}

		fmt.Println("✅ Database created successfully")

	case "drop":
		err := db.Drop(url)
		if err != nil {
			fmt.Println(err)

			return
		}

		fmt.Println("✅ Database dropped successfully")

	default:
		fmt.Println("command not found")

		return
	}
}
