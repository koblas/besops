package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/go-sql-driver/mysql" // register mysql driver
	_ "github.com/lib/pq"              // register postgres driver
)

func mysqlMigrateDriver(db *sql.DB) (database.Driver, error) {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("create mysql migration driver: %w", err)
	}
	return driver, nil
}

func postgresMigrateDriver(db *sql.DB) (database.Driver, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("create postgres migration driver: %w", err)
	}
	return driver, nil
}
