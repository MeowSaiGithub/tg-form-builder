package postgresql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go-tg-support-ticket/form"
	"strings"
	"testing"
)

func TestBuildCreateTableQuery(t *testing.T) {
	tests := []struct {
		name          string
		schema        form.Form
		expectedQuery string
		expectError   bool
	}{
		{
			name: "Valid table with multiple fields",
			schema: form.Form{
				TableName: "users",
				Fields: []form.Field{
					{Name: "name", DBType: "TEXT", Required: true},
					{Name: "email", DBType: "VARCHAR(255)", Required: false},
					{Name: "age", DBType: "INTEGER", Required: true},
				},
			},
			expectedQuery: `CREATE TABLE users ("id" UUID PRIMARY KEY DEFAULT gen_random_uuid(), "name" TEXT NOT NULL, "email" VARCHAR(255), "age" INTEGER NOT NULL);`,
			expectError:   false,
		},
		{
			name: "Table name is empty",
			schema: form.Form{
				TableName: "",
				Fields: []form.Field{
					{Name: "name", DBType: "TEXT"},
				},
			},
			expectedQuery: "",
			expectError:   true,
		},
		{
			name: "No fields provided",
			schema: form.Form{
				TableName: "empty_table",
				Fields:    []form.Field{},
			},
			expectedQuery: "",
			expectError:   true,
		},
		{
			name: "Some fields missing db_type",
			schema: form.Form{
				TableName: "partial_fields",
				Fields: []form.Field{
					{Name: "valid_field", DBType: "TEXT"},
					{Name: "invalid_field", DBType: ""},
				},
			},
			expectedQuery: `CREATE TABLE partial_fields ("id" UUID PRIMARY KEY DEFAULT gen_random_uuid(), "valid_field" TEXT);`,
			expectError:   false,
		},
		{
			name: "Handles uppercase field names correctly",
			schema: form.Form{
				TableName: "case_test",
				Fields: []form.Field{
					{Name: "FullName", DBType: "TEXT", Required: true},
					{Name: "EMAIL", DBType: "VARCHAR(255)", Required: false},
				},
			},
			expectedQuery: `CREATE TABLE case_test ("id" UUID PRIMARY KEY DEFAULT gen_random_uuid(), "fullname" TEXT NOT NULL, "email" VARCHAR(255));`,
			expectError:   false,
		},
		{
			name: "Only primary key field should exist when no db_type fields",
			schema: form.Form{
				TableName: "only_pk",
				Fields: []form.Field{
					{Name: "no_db_type_field", DBType: ""},
				},
			},
			expectedQuery: "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert field names to lowercase before testing
			for i := range tt.schema.Fields {
				tt.schema.Fields[i].Name = strings.ToLower(tt.schema.Fields[i].Name)
			}

			query, err := buildCreateTableQuery(&tt.schema)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
				assert.Empty(t, query, "Expected empty query on error")
				return
			}

			assert.NoError(t, err, "Did not expect an error")
			assert.Equal(t, tt.expectedQuery, query, fmt.Sprintf("Expected query:\n%s\nGot:\n%s", tt.expectedQuery, query))
		})
	}
}

func TestBuildInsertQuery(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		fields         []form.Field
		expectedQuery  string
		expectedValues []interface{}
		expectError    bool
	}{
		{
			name:      "All fields have db_type",
			tableName: "survey_responses",
			fields: []form.Field{
				{Name: "name", DBType: "TEXT", UserValue: "John Doe"},
				{Name: "email", DBType: "TEXT", UserValue: "john.doe@example.com"},
				{Name: "age", DBType: "INTEGER", UserValue: "30"},
			},
			expectedQuery:  "INSERT INTO survey_responses (name, email, age) VALUES ($1, $2, $3)",
			expectedValues: []interface{}{"John Doe", "john.doe@example.com", "30"},
			expectError:    false,
		},
		{
			name:      "Some fields missing db_type",
			tableName: "survey_responses",
			fields: []form.Field{
				{Name: "name", DBType: "TEXT", UserValue: "Jane Doe"},
				{Name: "email", DBType: "", UserValue: "jane.doe@example.com"}, // Ignored
				{Name: "age", DBType: "INTEGER", UserValue: "25"},
			},
			expectedQuery:  "INSERT INTO survey_responses (name, age) VALUES ($1, $2)",
			expectedValues: []interface{}{"Jane Doe", "25"},
			expectError:    false,
		},
		{
			name:      "No fields with db_type",
			tableName: "survey_responses",
			fields: []form.Field{
				{Name: "name", DBType: "", UserValue: "John Doe"},
				{Name: "email", DBType: "", UserValue: "john.doe@example.com"},
			},
			expectedQuery:  "",
			expectedValues: nil,
			expectError:    true,
		},
		{
			name:      "Empty table name",
			tableName: "",
			fields: []form.Field{
				{Name: "name", DBType: "TEXT", UserValue: "John Doe"},
			},
			expectedQuery:  "",
			expectedValues: nil,
			expectError:    true,
		},
		{
			name:      "Single field with db_type",
			tableName: "survey_responses",
			fields: []form.Field{
				{Name: "name", DBType: "TEXT", UserValue: "John Doe"},
			},
			expectedQuery:  "INSERT INTO survey_responses (name) VALUES ($1)",
			expectedValues: []interface{}{"John Doe"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, values, err := buildInsertQuery(tt.tableName, tt.fields)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if query != tt.expectedQuery {
				t.Errorf("Expected query: %s, got: %s", tt.expectedQuery, query)
			}

			if len(values) != len(tt.expectedValues) {
				t.Errorf("Expected %d values, got %d", len(tt.expectedValues), len(values))
				return
			}

			for i, value := range values {
				if value != tt.expectedValues[i] {
					t.Errorf("Expected value at index %d: %v, got: %v", i, tt.expectedValues[i], value)
				}
			}
		})
	}
}
