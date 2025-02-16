package bot

import (
	"go-tg-support-ticket/form"
	"strings"
	"testing"
)

func TestValidateField(t *testing.T) {
	tests := []struct {
		name        string
		field       form.Field
		userValue   string
		expectError bool
	}{
		{
			name: "Valid text field",
			field: form.Field{
				Name: "name",
				Type: "text",
				Validation: form.Validation{
					MinLength: 3,
					MaxLength: 100,
				},
			},
			userValue:   "John Doe",
			expectError: false,
		},
		{
			name: "Text field too short",
			field: form.Field{
				Name: "name",
				Type: "text",
				Validation: form.Validation{
					MinLength: 3,
				},
			},
			userValue:   "Jo",
			expectError: true,
		},
		{
			name: "Text field too long",
			field: form.Field{
				Name: "name",
				Type: "text",
				Validation: form.Validation{
					MaxLength: 100,
				},
			},
			userValue:   strings.Repeat("a", 101),
			expectError: true,
		},
		{
			name: "Text field with regex validation",
			field: form.Field{
				Name: "name",
				Type: "text",
				Validation: form.Validation{
					Regex: "^[a-zA-Z ]+$",
				},
			},
			userValue:   "John Doe",
			expectError: false,
		},
		{
			name: "Text field with invalid regex",
			field: form.Field{
				Name: "name",
				Type: "text",
				Validation: form.Validation{
					Regex: "[",
				},
			},
			userValue:   "John Doe",
			expectError: true,
		},
		{
			name: "Number field valid",
			field: form.Field{
				Name: "age",
				Type: "number",
				Validation: form.Validation{
					Min: 18,
					Max: 100,
				},
			},
			userValue:   "25",
			expectError: false,
		},
		{
			name: "Number field invalid",
			field: form.Field{
				Name: "age",
				Type: "number",
				Validation: form.Validation{
					Min: 18,
					Max: 100,
				},
			},
			userValue:   "101",
			expectError: true,
		},
		{
			name: "Select field valid",
			field: form.Field{
				Name:    "color",
				Type:    "select",
				Options: []string{"red", "green", "blue"},
			},
			userValue:   "green",
			expectError: false,
		},
		{
			name: "Select field invalid",
			field: form.Field{
				Name:    "color",
				Type:    "select",
				Options: []string{"red", "green", "blue"},
			},
			userValue:   "yellow",
			expectError: true,
		},
		{
			name: "File field required",
			field: form.Field{
				Name:     "file",
				Type:     "file",
				Required: true,
			},
			userValue:   "",
			expectError: true,
		},
		{
			name: "File field not required",
			field: form.Field{
				Name:     "file",
				Type:     "file",
				Required: false,
			},
			userValue:   "",
			expectError: false,
		},
		{
			name: "Text field with no validation rules",
			field: form.Field{
				Name:       "name",
				Type:       "text",
				UserValue:  "John Doe",
				Validation: form.Validation{},
			},
			userValue:   "John Doe",
			expectError: false,
		},
		{
			name: "Text field with empty value and not required",
			field: form.Field{
				Name:      "name",
				Type:      "text",
				UserValue: "",
				Validation: form.Validation{
					MinLength: 3,
				},
				Skippable: true,
			},
			userValue:   "",
			expectError: false,
		},
		{
			name: "Text field with empty value and required",
			field: form.Field{
				Name:      "name",
				Type:      "text",
				UserValue: "",
				Validation: form.Validation{
					MinLength: 3,
				},
				Required: true,
			},
			userValue:   "",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field := test.field
			_, err := ValidateField(field, test.userValue)
			if test.expectError && err == nil {
				t.Errorf("expected error, got nil")
			} else if !test.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
