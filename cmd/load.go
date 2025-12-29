package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"org.xyzmaps.xyzduck/src/database"
	"org.xyzmaps.xyzduck/src/geojson"
)

var (
	dbFlag    string
	tableFlag string
)

var loadCmd = &cobra.Command{
	Use:   "load <geojson-file>",
	Short: "Load GeoJSON file into DuckDB database",
	Long: `Load a GeoJSON file into a DuckDB table with automatic schema inference.

The table name is derived from the GeoJSON filename by default, but can be
overridden with the --table flag. If the table already exists, features will
be appended to it.`,
	Args: cobra.ExactArgs(1),
	RunE: runLoad,
}

func init() {
	loadCmd.Flags().StringVar(&dbFlag, "db", "", "Target database file (required)")
	loadCmd.MarkFlagRequired("db")
	loadCmd.Flags().StringVar(&tableFlag, "table", "", "Table name (default: derived from filename)")
	rootCmd.AddCommand(loadCmd)
}

func runLoad(cmd *cobra.Command, args []string) error {
	geojsonPath := args[0]

	// Validate GeoJSON file exists
	if !database.FileExists(geojsonPath) {
		return fmt.Errorf("GeoJSON file not found: %s", geojsonPath)
	}

	// Ensure database has .duckdb extension
	dbPath := database.EnsureDuckDBExtension(dbFlag)

	// Validate database exists
	if !database.FileExists(dbPath) {
		return fmt.Errorf("database not found: %s\nHint: Run 'xyzduck init %s' to create it", dbPath, dbFlag)
	}

	// Determine table name
	tableName := tableFlag
	if tableName == "" {
		// Derive from filename
		base := filepath.Base(geojsonPath)
		tableName = strings.TrimSuffix(base, filepath.Ext(base))
		// Clean up table name (replace invalid characters)
		tableName = strings.ReplaceAll(tableName, "-", "_")
		tableName = strings.ReplaceAll(tableName, " ", "_")
	}

	// Check if table exists
	tableExists, err := database.TableExists(dbPath, tableName)
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	if tableExists {
		fmt.Printf("Appending to existing table '%s' in %s...\n", tableName, dbPath)
	} else {
		fmt.Printf("Loading %s into %s...\n", filepath.Base(geojsonPath), dbPath)
	}

	// Load the GeoJSON file
	rowCount, err := geojson.LoadGeoJSON(dbPath, geojsonPath, tableName)
	if err != nil {
		return fmt.Errorf("failed to load GeoJSON: %w", err)
	}

	// Display success message
	fmt.Printf("✓ Loaded %d features into table '%s'\n", rowCount, tableName)

	// Show table schema
	schema, err := database.GetTableSchema(dbPath, tableName)
	if err == nil && len(schema) > 0 {
		var colNames []string
		for _, col := range schema {
			colNames = append(colNames, fmt.Sprintf("%s (%s)", col.Name, col.Type))
		}
		fmt.Printf("\nTable: %s\nColumns: %s\n", tableName, strings.Join(colNames, ", "))
	}

	return nil
}
