package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestMigrateUpDown(t *testing.T) {
	dir := t.TempDir()
	dbURL := "sqlite://" + filepath.Join(dir, "test.db")

	db, err := Open(dbURL)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := Migrate(db, dbURL); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	tables := queryTables(t, db)
	expected := []string{
		"user", "monitor", "heartbeat", "notification",
		"tag", "status_page", "maintenance", "setting",
		"api_key", "proxy",
	}
	for _, table := range expected {
		if !slices.Contains(tables, table) {
			t.Errorf("expected table %q not found in %v", table, tables)
		}
	}

	if err := MigrateDown(db, dbURL); err != nil {
		t.Fatalf("migrate down: %v", err)
	}
}

func TestOpenSQLite(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	dbURL := "sqlite://" + dbPath

	db, err := Open(dbURL)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("expected database file at %s", dbPath)
	}
}

func TestParseDatabaseURL(t *testing.T) {
	tests := []struct {
		url            string
		expectScheme   string
		expectContains string
	}{
		{"sqlite://./data/besops.db", "sqlite", "./data/besops.db"},
		{"sqlite:///tmp/test.db", "sqlite", "/tmp/test.db"},
		{"mysql://root:pass@localhost:3306/mydb", "mysql", "root:pass@tcp(localhost:3306)/mydb"},
		{"postgres://user:pw@host:5432/db?sslmode=disable", "postgres", "postgres://"},
	}

	for _, tt := range tests {
		scheme, dsn, err := parseDatabaseURL(tt.url)
		if err != nil {
			t.Errorf("parseDatabaseURL(%q): %v", tt.url, err)
			continue
		}
		if scheme != tt.expectScheme {
			t.Errorf("parseDatabaseURL(%q) scheme = %q, want %q", tt.url, scheme, tt.expectScheme)
		}
		if !strings.Contains(dsn, tt.expectContains) {
			t.Errorf("parseDatabaseURL(%q) dsn = %q, want to contain %q", tt.url, dsn, tt.expectContains)
		}
	}
}

func queryTables(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'schema_migrations'")
	if err != nil {
		t.Fatalf("query tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		tables = append(tables, name)
	}
	return tables
}
