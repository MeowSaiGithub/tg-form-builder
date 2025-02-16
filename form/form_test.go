package form

import (
	"testing"
)

// Test case structure
type testCase struct {
	name    string
	form    Form
	wantErr bool
}

// Unit tests for validateForm function
func TestValidateForm(t *testing.T) {
	testCases := []testCase{
		{
			name: "Valid Form",
			form: Form{
				FormName:      "Customer Survey",
				TableName:     "survey_responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name:        "welcome_message",
						Type:        "photo",
						Description: "Welcome to our survey! 🎉",
						Location:    "welcome.jpg",
						Buttons: []Button{
							{Text: "Start Survey 🚀", Data: "start_survey"},
						},
					},
					{
						Name:     "rating",
						Label:    "Rate our service",
						Type:     "select",
						DBType:   "VARCHAR(5)",
						Required: true,
						Options:  []string{"1", "2", "3", "4", "5"},
						Buttons: []Button{
							{Text: "⭐️ 1", Data: "rating_1"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing Form Name",
			form: Form{
				TableName:     "survey_responses",
				ReviewEnabled: true,
				Fields:        []Field{},
			},
			wantErr: true,
		},
		{
			name: "Missing Table Name",
			form: Form{
				FormName:      "Customer Survey",
				ReviewEnabled: true,
				Fields:        []Field{},
			},
			wantErr: true,
		},
		{
			name: "Location file does not exist",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name:     "image",
						Type:     "photo",
						Location: "nonexistent.jpg", // File does not exist
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Button without text",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name: "action",
						Type: "text",
						Buttons: []Button{
							{Text: "", Data: "valid_data"}, // Text missing
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Button without data",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name: "action",
						Type: "text",
						Buttons: []Button{
							{Text: "Click Me", Data: ""}, // Data missing
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Button data contains invalid characters",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name: "action",
						Type: "text",
						Buttons: []Button{
							{Text: "Click Me", Data: "invalid data!"}, // Invalid characters
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Select field without options",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name:    "selection",
						Type:    "select",
						Options: []string{}, // No options provided
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Min greater than Max",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name: "age",
						Type: "number",
						Validation: Validation{
							Min: 50,
							Max: 30, // Invalid constraint
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid regex pattern",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name: "email",
						Type: "text",
						Validation: Validation{
							Regex: "^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9.-]+$", // Valid regex
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid regex pattern",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name: "email",
						Type: "text",
						Validation: Validation{
							Regex: "[a-zA-Z0-9+_.-", // Invalid regex
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid Form with db_type",
			form: Form{
				FormName:      "Customer Survey",
				TableName:     "survey_responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name:     "name",
						Type:     "text",
						DBType:   "VARCHAR(255)", // Valid db_type
						Required: true,
					},
					{
						Name:     "age",
						Type:     "number",
						DBType:   "INT", // Valid db_type
						Required: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing db_type for field",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name:     "name",
						Type:     "text",
						Required: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid db_type for PostgreSQL",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				DB:            "postgres",
				Fields: []Field{
					{
						Name:     "email",
						Type:     "text",
						DBType:   "INVALID_TYPE", // Invalid db_type
						Required: true,
					},
				},
			},
			wantErr: true, // Error expected due to invalid db_type for PostgreSQL
		},
		{
			name: "Valid db_type for PostgreSQL",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name:     "email",
						Type:     "text",
						DBType:   "VARCHAR(255)", // Valid db_type for PostgreSQL
						Required: true,
					},
				},
			},
			wantErr: false, // No error expected for valid db_type
		},
		{
			name: "Valid db_type for MySQL",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name:     "email",
						Type:     "text",
						DBType:   "VARCHAR(255)", // Valid db_type for MySQL
						Required: true,
					},
				},
			},
			wantErr: false, // No error expected for valid db_type
		},
		{
			name: "Invalid db_type for MySQL",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				DB:            "mysql",
				Fields: []Field{
					{
						Name:     "email",
						Type:     "text",
						DBType:   "UNKNOWN_TYPE", // Invalid db_type for MySQL
						Required: true,
					},
				},
			},
			wantErr: true, // Error expected due to invalid db_type for MySQL
		},
		{
			name: "Valid db_type for MongoDB",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name:     "email",
						Type:     "text",
						DBType:   "STRING", // Valid db_type for MongoDB
						Required: true,
					},
				},
			},
			wantErr: false, // No error expected for valid db_type for MongoDB
		},
		{
			name: "Invalid db_type for MongoDB",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				DB:            "mongo",
				Fields: []Field{
					{
						Name:     "email",
						Type:     "text",
						DBType:   "UNKNOWN_TYPE", // Invalid db_type for MongoDB
						Required: true,
					},
				},
			},
			wantErr: true, // Error expected due to invalid db_type for MongoDB
		},
		{
			name: "Valid db_type with custom PostgreSQL field",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name:     "custom_field",
						Type:     "text",
						DBType:   "JSONB", // Valid custom db_type for PostgreSQL
						Required: true,
					},
				},
			},
			wantErr: false, // No error expected for custom db_type
		},
		{
			name: "Valid db_type for MongoDB with custom data",
			form: Form{
				FormName:      "Survey",
				TableName:     "responses",
				ReviewEnabled: true,
				Fields: []Field{
					{
						Name:     "custom_field",
						Type:     "text",
						DBType:   "OBJECT", // Valid custom db_type for MongoDB
						Required: true,
					},
				},
			},
			wantErr: false, // No error expected for custom db_type for MongoDB
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.form.ValidateForm()
			if (err != nil) != tc.wantErr {
				t.Errorf("Test case '%s' failed: expected error = %v, got error = %v", tc.name, tc.wantErr, err)
			}
		})
	}
}
