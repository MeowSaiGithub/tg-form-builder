package form

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Button struct {
	Text string `json:"text"`
	Data string `json:"data"`
}

type Field struct {
	Name         string     `json:"name"`
	Label        string     `json:"label"`
	Type         string     `json:"type"`
	DBType       string     `json:"db_type"`
	ActualDBType string     `json:"-"`
	Required     bool       `json:"required"`
	Skippable    bool       `json:"skippable"`
	Description  string     `json:"description"`
	Formatting   string     `json:"formatting"`
	Location     string     `json:"location"`
	PhotoData    []byte     `json:"-"` // Field-specific photo data internal use only
	Buttons      []Button   `json:"buttons"`
	Options      []string   `json:"options,omitempty"`
	UserValue    string     `json:"user_value"`
	Validation   Validation `json:"validation,omitempty"`
}

type Validation struct {
	MinLength int    `json:"min_length,omitempty"`
	MaxLength int    `json:"max_length,omitempty"`
	Regex     string `json:"regex,omitempty"`
	Min       int    `json:"min,omitempty"`
	Max       int    `json:"max,omitempty"`
}

type Form struct {
	FormName      string  `json:"form_name"`
	TableName     string  `json:"table_name"`
	ReviewEnabled bool    `json:"review_enabled"`
	SubmitMessage string  `json:"submit_message"`
	Fields        []Field `json:"fields"`
	DB            string  `json:"db"`
}

func LoadTicketFormat(path string) (*Form, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ticket.json file: %w", err)
	}

	var tf Form
	if err := json.Unmarshal(file, &tf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ticket.json: %w", err)
	}

	return &tf, nil
}

// Check if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// Validate button data format (Ensure it only contains alphanumeric, underscores, and dashes)
func isValidButtonData(data string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return re.MatchString(data)
}

// validateField a single field
// Collects and returns all errors for a single field validation
func validateField(field Field) []error {
	var errs []error

	// 1. If type is "photo", "document" and "video", the file must exist
	if field.Type == "photo" || field.Type == "document" || field.Type == "video" {
		if field.Location == "" {
			errs = append(errs, fmt.Errorf("field '%s' location is empty", field.Name))
		} else if !fileExists(filepath.Clean(field.Location)) {
			errs = append(errs, fmt.Errorf("field '%s' has a missing media file: %s", field.Name, field.Location))
		}
	}

	// 2. Buttons must have both text and data, and data must be valid
	for _, button := range field.Buttons {
		if button.Text == "" || button.Data == "" {
			errs = append(errs, fmt.Errorf("button in field '%s' must have both text and data", field.Name))
		}
		if !isValidButtonData(button.Data) {
			errs = append(errs, fmt.Errorf("button data '%s' in field '%s' contains invalid characters", button.Data, field.Name))
		}
	}

	// 3. Ensure options exist for `select` type fields
	if field.Type == "select" && len(field.Options) == 0 {
		errs = append(errs, fmt.Errorf("select field '%s' must have options", field.Name))
	}

	// 4. Validate numerical constraints (Min < Max)
	if field.Validation.Min > field.Validation.Max {
		errs = append(errs, fmt.Errorf("field '%s' has invalid min/max constraints", field.Name))
	}
	if field.Validation.MinLength > field.Validation.MaxLength {
		errs = append(errs, fmt.Errorf("field '%s' has invalid min/max length constraints", field.Name))
	}

	// 5. Ensure regex is valid
	if field.Validation.Regex != "" {
		_, err := regexp.Compile(field.Validation.Regex)
		if err != nil {
			errs = append(errs, fmt.Errorf("field '%s' has an invalid regex pattern", field.Name))
		}
	}

	return errs
}

// ValidateForm the whole form
// Collects and returns all errors for form validation
func (form *Form) ValidateForm() []error {
	var errs []error

	// 1. Ensure form name and table name are non-empty
	if form.FormName == "" {
		errs = append(errs, fmt.Errorf("form_name cannot be empty"))
	}
	if form.TableName == "" {
		errs = append(errs, fmt.Errorf("table_name cannot be empty"))
	}

	// 2. Ensure ReviewEnabled is a bool (automatic in Go)
	if form.ReviewEnabled != true && form.ReviewEnabled != false {
		errs = append(errs, fmt.Errorf("review_enabled must be a boolean type"))
	}

	// 3. Validate each field and collect any errors
	for _, field := range form.Fields {
		fieldErrors := validateField(field)
		errs = append(errs, fieldErrors...)
	}

	// 4. DB validation (if applicable)
	if form.DB != "" {
		var err error
		for i, field := range form.Fields {
			if field.DBType != "" {
				form.Fields[i].ActualDBType, err = validateDBType(form.DB, field.DBType)
				if err != nil {
					errs = append(errs, fmt.Errorf("DB validation failed for field '%s': %v", field.Name, err))
				}
			}
		}
	}

	return errs
}

// Define a mapping of custom types to actual DB types
var dbTypeMapping = map[string]map[string]string{
	"mysql": {
		"STRING":   "VARCHAR(255)",
		"TEXT":     "TEXT",
		"NUMBER":   "INT",
		"BOOLEAN":  "BOOLEAN",
		"DATETIME": "DATETIME",
		"JSON":     "JSON",
	},
	"postgres": {
		"STRING":   "VARCHAR(255)",
		"TEXT":     "TEXT",
		"NUMBER":   "INTEGER",
		"BOOLEAN":  "BOOLEAN",
		"DATETIME": "TIMESTAMP",
		"JSON":     "JSONB",
	},
	"mongo": {
		"STRING":   "string",
		"TEXT":     "string",
		"NUMBER":   "int",
		"BOOLEAN":  "bool",
		"DATETIME": "date",
		"OBJECT":   "object",
	},
	"sqlite": {
		"STRING":   "VARCHAR(255)", // Or "TEXT", VARCHAR is usually treated as TEXT in SQLite, but kept for consistency with other SQL types STRING mapping
		"TEXT":     "TEXT",
		"NUMBER":   "INTEGER", // Or "NUMERIC" if you want to store both integers and floats, but INTEGER is common for "NUMBER" type mapping
		"BOOLEAN":  "INTEGER", // SQLite doesn't have a dedicated BOOLEAN type, INTEGER with 0 and 1 is common practice
		"DATETIME": "TEXT",    // SQLite best practice for DATETIME is to store as TEXT in ISO8601 format, or INTEGER as Unix Time
		"JSON":     "TEXT",    // SQLite stores JSON as TEXT using JSON1 extension (you need to ensure JSON1 extension is enabled in your SQLite build if you intend to use JSON functions)
	},
}

// Validate and map DB type
func validateDBType(db string, userType string) (string, error) {
	userType = strings.ToUpper(userType) // Normalize input

	// Try mapping user type to DB-specific type
	if mappedTypes, ok := dbTypeMapping[db]; ok {
		if dbType, exists := mappedTypes[userType]; exists {
			return dbType, nil
		}
	}

	// Check if the userType is already a valid direct type
	validTypes := getValidDBTypes(db)
	if contains(validTypes, userType) {
		return userType, nil // Use as is
	}

	// Special case: Validate VARCHAR(N) for MySQL & PostgreSQL
	if (db == "mysql" || db == "postgres" || db == "sqlite") && isValidVarchar(userType) {
		return userType, nil
	}

	// If invalid, return an error with suggestions
	return "", fmt.Errorf("❌ Invalid db_type: '%s' for %s. Allowed: %v", userType, db, validTypes)
}

// Get list of valid types for a database
func getValidDBTypes(db string) []string {
	switch db {
	case "mysql":
		return []string{"VARCHAR(255)", "TEXT", "INT", "BOOLEAN", "DATETIME", "JSON"}
	case "postgres":
		return []string{"TEXT", "VARCHAR(255)", "INTEGER", "BOOLEAN", "TIMESTAMP", "JSONB"}
	case "mongo":
		return []string{"string", "int", "bool", "date", "object"}
	}
	return nil
}

// Helper function to check if value exists in slice
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Validate VARCHAR(N) format for MySQL & PostgreSQL
func isValidVarchar(userType string) bool {
	varcharRegex := regexp.MustCompile(`^VARCHAR\(\d+\)$`)
	return varcharRegex.MatchString(userType)
}
