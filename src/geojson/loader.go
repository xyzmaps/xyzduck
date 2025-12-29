package geojson

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"org.xyzmaps.xyzduck/src/database"
)

// GeoJSON structures
type GeoJSON struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string                 `json:"type"`
	Geometry   json.RawMessage        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// Schema represents a table schema
type Schema struct {
	Columns []database.Column
}

// LoadGeoJSON loads a GeoJSON file into a DuckDB database table
func LoadGeoJSON(dbPath, geojsonPath, tableName string) (int, error) {
	// Get absolute paths
	absDBPath, err := filepath.Abs(dbPath)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve database path: %w", err)
	}

	absGeoJSONPath, err := filepath.Abs(geojsonPath)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve GeoJSON path: %w", err)
	}

	// Open database
	db, err := sql.Open("duckdb", absDBPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Ensure spatial extension is loaded
	if err := loadSpatialExtension(db); err != nil {
		return 0, err
	}

	// Check if table exists
	tableExists, err := database.TableExists(absDBPath, tableName)
	if err != nil {
		return 0, fmt.Errorf("failed to check if table exists: %w", err)
	}

	if !tableExists {
		// Infer schema from GeoJSON
		schema, err := inferSchemaFromGeoJSON(absGeoJSONPath)
		if err != nil {
			return 0, fmt.Errorf("failed to infer schema: %w", err)
		}

		// Create table
		if err := createTableFromSchema(db, tableName, schema); err != nil {
			return 0, fmt.Errorf("failed to create table: %w", err)
		}

		fmt.Printf("✓ Table '%s' created with %d columns\n", tableName, len(schema.Columns))
	}

	// Load data into table
	rowCount, err := loadDataIntoTable(db, absDBPath, tableName, absGeoJSONPath)
	if err != nil {
		return 0, fmt.Errorf("failed to load data: %w", err)
	}

	return rowCount, nil
}

// loadSpatialExtension ensures the spatial extension is loaded
func loadSpatialExtension(db *sql.DB) error {
	_, err := db.Exec("LOAD spatial;")
	if err != nil {
		return fmt.Errorf("failed to load spatial extension: %w", err)
	}
	return nil
}

// inferSchemaFromGeoJSON reads the first feature to infer the table schema
func inferSchemaFromGeoJSON(geojsonPath string) (Schema, error) {
	data, err := os.ReadFile(geojsonPath)
	if err != nil {
		return Schema{}, fmt.Errorf("failed to read GeoJSON file: %w", err)
	}

	var gj GeoJSON
	if err := json.Unmarshal(data, &gj); err != nil {
		return Schema{}, fmt.Errorf("failed to parse GeoJSON: %w", err)
	}

	if len(gj.Features) == 0 {
		return Schema{}, fmt.Errorf("GeoJSON file contains no features")
	}

	// Infer types from first feature
	firstFeature := gj.Features[0]
	var columns []database.Column

	for key, value := range firstFeature.Properties {
		colType := inferType(value)
		columns = append(columns, database.Column{
			Name: key,
			Type: colType,
		})
	}

	// Always add geometry column
	columns = append(columns, database.Column{
		Name: "geom",
		Type: "GEOMETRY",
	})

	return Schema{Columns: columns}, nil
}

// inferType infers DuckDB type from Go value
func inferType(value interface{}) string {
	switch v := value.(type) {
	case string:
		return "VARCHAR"
	case float64:
		// Check if it's an integer
		if v == float64(int64(v)) {
			return "BIGINT"
		}
		return "DOUBLE"
	case bool:
		return "BOOLEAN"
	case nil:
		return "VARCHAR" // Default for null
	default:
		return "VARCHAR" // Default fallback
	}
}

// createTableFromSchema creates a table with the inferred schema
func createTableFromSchema(db *sql.DB, tableName string, schema Schema) error {
	var colDefs []string
	for _, col := range schema.Columns {
		colDefs = append(colDefs, fmt.Sprintf("%s %s", col.Name, col.Type))
	}

	createSQL := fmt.Sprintf("CREATE TABLE %s (%s)", tableName, strings.Join(colDefs, ", "))
	_, err := db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to execute CREATE TABLE: %w", err)
	}

	return nil
}

// loadDataIntoTable loads GeoJSON features into the specified table
func loadDataIntoTable(db *sql.DB, dbPath, tableName, geojsonPath string) (int, error) {
	// First, create a temporary view of the GeoJSON file
	createTempSQL := fmt.Sprintf(`
		CREATE TEMPORARY TABLE temp_geojson AS
		SELECT * FROM read_json_auto('%s')
	`, geojsonPath)

	_, err := db.Exec(createTempSQL)
	if err != nil {
		return 0, fmt.Errorf("failed to read GeoJSON file: %w", err)
	}
	defer db.Exec("DROP TABLE IF EXISTS temp_geojson")

	// Get the column names from the target table (excluding geom)
	schema, err := database.GetTableSchema(dbPath, tableName)
	if err != nil {
		return 0, fmt.Errorf("failed to get table schema: %w", err)
	}

	// Build column list (excluding geometry)
	var propCols []string
	for _, col := range schema {
		if col.Name != "geom" {
			propCols = append(propCols, col.Name)
		}
	}

	// Build the SELECT part for properties
	var selectCols []string
	for _, colName := range propCols {
		selectCols = append(selectCols, fmt.Sprintf("properties->>'%s' as %s", colName, colName))
	}
	selectCols = append(selectCols, "ST_GeomFromGeoJSON(json(geometry)) as geom")

	// Build and execute INSERT statement
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s
		SELECT %s
		FROM (
			SELECT unnest(features) as feature
			FROM temp_geojson
		) sub,
		LATERAL (
			SELECT
				feature->'properties' as properties,
				feature->'geometry' as geometry
		) extracted
	`, tableName, strings.Join(selectCols, ", "))

	result, err := db.Exec(insertSQL)
	if err != nil {
		return 0, fmt.Errorf("failed to insert data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}
