package sqlite

import (
	"go-tg-support-ticket/form"
	"testing"
)

// Test for buildCreateTableQuery function
func TestBuildCreateTableQuery(t *testing.T) {
	tests := []struct {
		name       string
		schema     form.Form
		wantQuery  string
		shouldFail bool
	}{
		{
			name: "Valid Table Creation",
			schema: form.Form{
				TableName: "users",
				Fields: []form.Field{
					{Name: "name", ActualDBType: "TEXT", Required: true},
					{Name: "email", ActualDBType: "TEXT", Required: false},
				},
			},
			wantQuery:  "CREATE TABLE users (id TEXT PRIMARY KEY, name TEXT NOT NULL, email TEXT);",
			shouldFail: false,
		},
		{
			name: "Empty Table Name",
			schema: form.Form{
				TableName: "",
				Fields:    []form.Field{},
			},
			wantQuery:  "",
			shouldFail: true,
		},
		{
			name: "No Fields Provided",
			schema: form.Form{
				TableName: "empty_table",
				Fields:    []form.Field{},
			},
			wantQuery:  "",
			shouldFail: true,
		},
		{
			name: "Table with Numeric and Boolean Fields",
			schema: form.Form{
				TableName: "products",
				Fields: []form.Field{
					{Name: "price", ActualDBType: "REAL", Required: true},
					{Name: "in_stock", ActualDBType: "BOOLEAN", Required: false},
				},
			},
			wantQuery:  "CREATE TABLE products (id TEXT PRIMARY KEY, price REAL NOT NULL, in_stock BOOLEAN);",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, err := buildCreateTableQuery(&tt.schema)
			if (err != nil) != tt.shouldFail {
				t.Fatalf("Expected error: %v, got: %v", tt.shouldFail, err)
			}
			if !tt.shouldFail && gotQuery != tt.wantQuery {
				t.Errorf("Expected query:\n%s\ngot:\n%s", tt.wantQuery, gotQuery)
			}
		})
	}
}

func TestBuildInsertQuery(t *testing.T) {
	tests := []struct {
		name       string
		tableName  string
		fields     []form.Field
		wantQuery  string
		wantValues []interface{}
		shouldFail bool
	}{
		{
			name:      "Valid Insert Query",
			tableName: "users",
			fields: []form.Field{
				{Name: "name", ActualDBType: "TEXT", UserValue: "John Doe"},
				{Name: "email", ActualDBType: "TEXT", UserValue: "john@example.com"},
			},
			wantQuery:  "INSERT INTO users (id, name, email) VALUES (?, ?, ?)",
			wantValues: []interface{}{"John Doe", "john@example.com"},
			shouldFail: false,
		},
		{
			name:       "Empty Table Name",
			tableName:  "",
			fields:     []form.Field{},
			wantQuery:  "",
			wantValues: nil,
			shouldFail: true,
		},
		{
			name:      "Insert with Numeric and Boolean Values",
			tableName: "products",
			fields: []form.Field{
				{Name: "product_name", ActualDBType: "TEXT", UserValue: "Laptop"},
				{Name: "price", ActualDBType: "REAL", UserValue: "1200.50"},
				{Name: "in_stock", ActualDBType: "BOOLEAN", UserValue: "true"},
			},
			wantQuery:  "INSERT INTO products (id, product_name, price, in_stock) VALUES (?, ?, ?, ?)",
			wantValues: []interface{}{"Laptop", "1200.50", "true"},
			shouldFail: false,
		},
		{
			name:      "Insert with Null Values for Optional Fields",
			tableName: "users",
			fields: []form.Field{
				{Name: "name", ActualDBType: "TEXT", UserValue: "Alice"},
				{Name: "nickname", ActualDBType: "TEXT", UserValue: ""},
			},
			wantQuery:  "INSERT INTO users (id, name, nickname) VALUES (?, ?, ?)",
			wantValues: []interface{}{"Alice", ""},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotValues, err := buildInsertQuery(tt.tableName, tt.fields)
			if (err != nil) != tt.shouldFail {
				t.Fatalf("Expected error: %v, got: %v", tt.shouldFail, err)
			}

			if !tt.shouldFail {
				// Check the query
				if gotQuery != tt.wantQuery {
					t.Errorf("Expected query:\n%s\ngot:\n%s", tt.wantQuery, gotQuery)
				}

				// Check the values (skip the first value, which is the UUID)
				if len(gotValues) != len(tt.wantValues)+1 {
					t.Errorf("Expected %d values, got %d", len(tt.wantValues)+1, len(gotValues))
				}

				// Compare the values (skip the first value, which is the UUID)
				for i, val := range tt.wantValues {
					if gotValues[i+1] != val {
						t.Errorf("Expected value at index %d: %v, got: %v", i, val, gotValues[i+1])
					}
				}
			}
		})
	}
}
