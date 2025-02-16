package mysql

import (
	"go-tg-support-ticket/form"
	"testing"
)

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
			name:      "All fields have values",
			tableName: "survey_responses",
			fields: []form.Field{
				{Name: "name", UserValue: "John Doe"},
				{Name: "email", UserValue: "john.doe@example.com"},
				{Name: "age", UserValue: "30"},
			},
			expectedQuery:  "INSERT INTO survey_responses (name, email, age) VALUES (?, ?, ?)",
			expectedValues: []interface{}{"John Doe", "john.doe@example.com", "30"},
			expectError:    false,
		},
		{
			name:      "Some fields have no values",
			tableName: "survey_responses",
			fields: []form.Field{
				{Name: "name", UserValue: "Jane Doe"},
				{Name: "email", UserValue: ""}, // Empty value
				{Name: "age", UserValue: "25"},
			},
			expectedQuery:  "INSERT INTO survey_responses (name, age) VALUES (?, ?)",
			expectedValues: []interface{}{"Jane Doe", "25"},
			expectError:    false,
		},
		{
			name:      "No fields have values",
			tableName: "survey_responses",
			fields: []form.Field{
				{Name: "name", UserValue: ""},
				{Name: "email", UserValue: ""},
				{Name: "age", UserValue: ""},
			},
			expectedQuery:  "",
			expectedValues: nil,
			expectError:    true,
		},
		{
			name:      "Empty table name",
			tableName: "",
			fields: []form.Field{
				{Name: "name", UserValue: "John Doe"},
			},
			expectedQuery:  "",
			expectedValues: nil,
			expectError:    true,
		},
		{
			name:      "Single field with value",
			tableName: "survey_responses",
			fields: []form.Field{
				{Name: "name", UserValue: "John Doe"},
			},
			expectedQuery:  "INSERT INTO survey_responses (name) VALUES (?)",
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
