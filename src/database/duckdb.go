package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/duckdb/duckdb-go/v2"
)

// EnsureDuckDBExtension adds .duckdb extension if not present
func EnsureDuckDBExtension(filename string) string {
	if !strings.HasSuffix(filename, ".duckdb") {
		return filename + ".duckdb"
	}
	return filename
}

// FileExists checks if a file exists at the given path
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// CreateOrOpenDatabase creates a new DuckDB database or opens an existing one
func CreateOrOpenDatabase(filename string) error {
	// Get absolute path for better error messages
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Open/create the database
	db, err := sql.Open("duckdb", absPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return nil
}

// InitSpatialExtension installs and loads the spatial extension
func InitSpatialExtension(filename string) error {
	// Get absolute path
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Open the database
	db, err := sql.Open("duckdb", absPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Install spatial extension
	_, err = db.Exec("INSTALL spatial;")
	if err != nil {
		return fmt.Errorf("failed to install spatial extension: %w", err)
	}

	// Load spatial extension
	_, err = db.Exec("LOAD spatial;")
	if err != nil {
		return fmt.Errorf("failed to load spatial extension: %w", err)
	}

	return nil
}

// Column represents a database table column
type Column struct {
	Name string
	Type string
}

// TableExists checks if a table exists in the database
func TableExists(dbPath, tableName string) (bool, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	db, err := sql.Open("duckdb", absPath)
	if err != nil {
		return false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	var exists bool
	query := `
		SELECT COUNT(*) > 0
		FROM information_schema.tables
		WHERE table_name = ?
	`
	err = db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return exists, nil
}

// GetTableSchema returns the schema of a table
func GetTableSchema(dbPath, tableName string) ([]Column, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	db, err := sql.Open("duckdb", absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	query := `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = ?
		ORDER BY ordinal_position
	`
	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(&col.Name, &col.Type); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return columns, nil
}
