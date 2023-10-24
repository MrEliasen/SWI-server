package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/libsql/libsql-client-go/libsql"
	"github.com/mreliasen/swi-server/internal/logger"
	"github.com/pterm/pterm"
)

var dbConnection *sql.DB

func Connect(dbUrl string) (*sql.DB, bool) {
	if dbConnection != nil {
		return dbConnection, true
	}

	logger.Logger.Info("Connecting to DB..")

	db, err := sql.Open("libsql", dbUrl)
	if err != nil {
		logger.Logger.Fatal(pterm.Sprintf("Failed to connect to db %s: %s", dbUrl, err))
		return dbConnection, false
	}

	dbConnection = db
	logger.Logger.Info("DB connected.")
	return dbConnection, true
}

func RunMigration(conn *sql.DB) {
	driver, err := sqlite.WithInstance(conn, &sqlite.Config{})
	if err != nil {
		logger.Logger.Fatal("Failed to run migrations")
		fmt.Println(err)
		return
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"swi",
		driver,
	)
	if err != nil {
		logger.Logger.Fatal("Failed to find migrations")
		fmt.Println(err)
		return
	}

	m.Up()
}
