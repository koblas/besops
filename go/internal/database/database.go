package database

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "modernc.org/sqlite" // register sqlite driver
)

//go:embed migrations/*.sql
var migrations embed.FS

// Open parses a database URL and returns a connected *sql.DB.
//
// Supported URL schemes:
//   - sqlite:///path/to/file.db or sqlite://./relative/path.db
//   - mysql://user:pass@host:port/dbname
//   - postgres://user:pass@host:port/dbname?sslmode=disable
func Open(databaseURL string) (*sql.DB, error) {
	scheme, dsn, err := parseDatabaseURL(databaseURL)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(scheme, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if scheme == "sqlite" {
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("set WAL mode: %w", err)
		}
		if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
			db.Close()
			return nil, fmt.Errorf("enable foreign keys: %w", err)
		}
	}

	return db, nil
}

// Scheme extracts the database type from a DATABASE_URL.
func Scheme(databaseURL string) string {
	scheme, _, _ := parseDatabaseURL(databaseURL)
	return scheme
}

func Migrate(db *sql.DB, databaseURL string) error {
	m, err := newMigrator(db, databaseURL)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func MigrateDown(db *sql.DB, databaseURL string) error {
	m, err := newMigrator(db, databaseURL)
	if err != nil {
		return err
	}

	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("rollback migration: %w", err)
	}

	return nil
}

func newMigrator(db *sql.DB, databaseURL string) (*migrate.Migrate, error) {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("create migration source: %w", err)
	}

	scheme := Scheme(databaseURL)

	switch scheme {
	case "sqlite":
		driver, err := sqlite.WithInstance(db, &sqlite.Config{})
		if err != nil {
			return nil, fmt.Errorf("create sqlite driver: %w", err)
		}
		m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
		if err != nil {
			return nil, fmt.Errorf("create migrator: %w", err)
		}
		return m, nil
	case "mysql":
		mysqlDriver, err := mysqlMigrateDriver(db)
		if err != nil {
			return nil, err
		}
		m, err := migrate.NewWithInstance("iofs", source, "mysql", mysqlDriver)
		if err != nil {
			return nil, fmt.Errorf("create migrator: %w", err)
		}
		return m, nil
	case "postgres":
		pgDriver, err := postgresMigrateDriver(db)
		if err != nil {
			return nil, err
		}
		m, err := migrate.NewWithInstance("iofs", source, "postgres", pgDriver)
		if err != nil {
			return nil, fmt.Errorf("create migrator: %w", err)
		}
		return m, nil
	default:
		return nil, fmt.Errorf("unsupported database type for migration: %s", scheme)
	}
}

// parseDatabaseURL returns (driverName, dsn, error).
//
//   - sqlite://./data/besops.db  → ("sqlite", "./data/besops.db?_pragma=busy_timeout(10000)")
//   - mysql://user:pass@host:3306/db → ("mysql", "user:pass@tcp(host:3306)/db?parseTime=true&multiStatements=true")
//   - postgres://user:pass@host:5432/db → ("postgres", "postgres://user:pass@host:5432/db")
func parseDatabaseURL(rawURL string) (string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("parse database URL: %w", err)
	}

	switch u.Scheme {
	case "sqlite", "sqlite3":
		path := u.Host + u.Path
		if path == "" {
			return "", "", fmt.Errorf("sqlite URL must include a file path")
		}
		dsn := path + "?_pragma=busy_timeout(10000)"
		return "sqlite", dsn, nil

	case "mysql":
		password, _ := u.User.Password()
		host := u.Host
		if !strings.Contains(host, ":") {
			host += ":3306"
		}
		dbName := strings.TrimPrefix(u.Path, "/")
		params := "parseTime=true&multiStatements=true"
		if u.RawQuery != "" {
			params += "&" + u.RawQuery
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
			u.User.Username(), password, host, dbName, params)
		return "mysql", dsn, nil

	case "postgres", "postgresql":
		return "postgres", rawURL, nil

	default:
		return "", "", fmt.Errorf("unsupported database scheme: %s", u.Scheme)
	}
}
