//go:build mysql || all
// +build mysql all

package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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

func (a *adaptor) Open(dsn string) error {
	db, err := sqlx.Connect("mysql", dsn)
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
	return "mysql"
}
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

	log.Println("✅ MySQL Table Created:", schema.TableName)
	return nil
}

// Check if table exists in MySQL
func (a *adaptor) tableExists(tableName string) (bool, error) {
	query := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	var result string
	err := a.db.QueryRow(query).Scan(&result)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // Table does not exist
	} else if err != nil {
		return false, err // Other errors
	}
	return true, nil // Table exists
}

// BuildCreateTableQuery generates a CREATE TABLE SQL statement dynamically
func buildCreateTableQuery(schema *form.Form) (string, error) {
	for _, f := range schema.Fields {
		fmt.Println("name ", f.Name)
		fmt.Println("dbtype ", f.DBType)
		fmt.Println("actualdbtype ", f.ActualDBType)
	}

	if schema.TableName == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}

	// Ensure at least one field has `db_type`
	var columns []string
	for _, field := range schema.Fields {
		if field.Name != "" && field.ActualDBType != "" {
			column := fmt.Sprintf("%s %s", field.Name, field.ActualDBType)
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
	columns = append([]string{"id VARCHAR(36) PRIMARY KEY"}, columns...)

	// Build the final SQL query
	query := fmt.Sprintf("CREATE TABLE %s (%s);", schema.TableName, strings.Join(columns, ", "))

	return query, nil
}

// buildInsertQuery generates an INSERT query for the given table and fields.
// It returns the query and the corresponding values.
func buildInsertQuery(tableName string, fields []form.Field) (string, []interface{}, error) {
	if tableName == "" {
		return "", nil, fmt.Errorf("table name is empty")
	}

	var columns []string
	var values []interface{}

	uid, _ := uuid.NewV7()
	columns = append(columns, "id")
	values = append(values, uid.String())

	for _, field := range fields {
		if field.ActualDBType != "" { // Only include fields with user input
			columns = append(columns, strings.ToLower(field.Name))
			values = append(values, field.UserValue)
		}
	}

	if len(columns) == 0 {
		return "", nil, fmt.Errorf("no user inputs to insert")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Repeat("?, ", len(columns)-1)+"?",
	)

	return query, values, nil
}

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
