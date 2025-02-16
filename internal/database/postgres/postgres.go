//go:build postgres || all
// +build postgres all

package postgresql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go-tg-support-ticket/form"
	"go-tg-support-ticket/internal/store"
	"log"
	"strings"
	"time"
)

func init() {
	store.RegisterAdaptor(&adaptor{})
}

type adaptor struct {
	db *sqlx.DB
}

// Open establishes a connection to PostgreSQL.
func (a *adaptor) Open(dsn string) error {
	db, err := sqlx.Connect("postgres", dsn) // Using `pgx` for PostgreSQL
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	a.db = db
	return nil
}

func (a *adaptor) GetName() string {
	return "postgres"
}

// Migrate checks if a table exists, then creates it if necessary.
func (a *adaptor) Migrate(schema *form.Form) error {
	// Check if table exists
	exists, err := a.tableExists(schema.TableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if exists {
		log.Println("⚠️ Table already exists:", schema.TableName)
		return nil // No migration needed
	}

	// Generate CREATE TABLE query
	query, err := buildCreateTableQuery(schema)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	// Execute the query
	_, err = a.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	log.Println("✅ PostgreSQL Table Created:", schema.TableName)
	return nil
}

// Check if a table exists in PostgreSQL
func (a *adaptor) tableExists(tableName string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1);`
	var exists bool
	err := a.db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, err // Other errors
	}
	return exists, nil // Table exists
}

// BuildCreateTableQuery generates a CREATE TABLE SQL statement dynamically for PostgreSQL.
func buildCreateTableQuery(schema *form.Form) (string, error) {
	if schema.TableName == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}

	// Ensure at least one field has `db_type`
	var columns []string
	for _, field := range schema.Fields {
		if field.Name != "" && field.ActualDBType != "" {
			column := fmt.Sprintf(`"%s" %s`, field.Name, field.ActualDBType)
			if field.Required {
				column += " NOT NULL"
			}
			columns = append(columns, column)
		}
	}

	// Ensure we have valid columns
	if len(columns) == 0 {
		return "", fmt.Errorf("no valid fields with db_type found")
	}

	// Add primary key column
	columns = append([]string{`"id" UUID PRIMARY KEY DEFAULT gen_random_uuid()`}, columns...)

	// Build the final SQL query
	query := fmt.Sprintf(`CREATE TABLE %s (%s);`, schema.TableName, strings.Join(columns, ", "))

	return query, nil
}

// buildInsertQuery generates an INSERT query for the given table and fields in PostgreSQL.
func buildInsertQuery(tableName string, fields []form.Field) (string, []interface{}, error) {
	if tableName == "" {
		return "", nil, fmt.Errorf("table name is empty")
	}

	var columns []string
	var values []interface{}

	for _, field := range fields {
		if field.ActualDBType != "" { // Only include fields with user input
			columns = append(columns, strings.ToLower(field.Name)) // Ensure column names are properly quoted
			values = append(values, field.UserValue)
		}
	}

	if len(columns) == 0 {
		return "", nil, fmt.Errorf("no user inputs to insert")
	}

	// Generate placeholders dynamically ($1, $2, etc.)
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return query, values, nil
}

// InsertUserInputs inserts data into PostgreSQL.
func (a *adaptor) InsertUserInputs(tableName string, fields []form.Field) error {
	// Build the INSERT query and get the values
	query, values, err := buildInsertQuery(tableName, fields)
	if err != nil {
		return fmt.Errorf("failed to build INSERT query: %w", err)
	}

	// Execute the INSERT query
	_, err = a.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to execute INSERT query: %w", err)
	}

	return nil
}
