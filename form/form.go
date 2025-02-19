package form

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
	FormName      string `json:"form_name"`
	TableName     string `json:"table_name"`
	ReviewEnabled bool   `json:"review_enabled"`
	//SubmitMessage string  `json:"submit_message"`
	Messages Message `json:"messages"`
	Fields   []Field `json:"fields"`
	DB       string  `json:"db"`
}

type Message struct {
	Submit              string `json:"submit"`
	SubmitButton        string `json:"submit_button"`
	SkipButton          string `json:"skip_button"`
	Modify              string `json:"modify"`
	ModifyButton        string `json:"modify_button"`
	ChooseOption        string `json:"choose_option"`
	Review              string `json:"review"`
	FileUploadSuccess   string `json:"file_upload_success"`
	UploadAnother       string `json:"upload_another"`
	UploadAnotherButton string `json:"upload_another_button"`
	FinishUploadButton  string `json:"finish_upload_button"`
	FinishUpload        string `json:"finish_upload"`
	RequiredFile        string `json:"required_file"`
	RequiredSelect      string `json:"required_select"`
	RequiredInput       string `json:"required_input"`
	InvalidEmail        string `json:"invalid_email"`
	InvalidMaxNumber    string `json:"invalid_max_number"`
	InvalidMinNumber    string `json:"invalid_min_number"`
	InvalidNumber       string `json:"invalid_number"`
	InvalidFormat       string `json:"invalid_format"`
	ValidationError     string `json:"validation_error"`
	InvalidMaxLength    string `json:"invalid_max_length"`
	InvalidMinLength    string `json:"invalid_min_length"`
}

const (
	Submit              string = "🎉 Thank you for submitting the form! 🎉"
	SubmitButton        string = "✅ Submit"
	SkipButton          string = "⏭️ Skip"
	Modify              string = "Please enter a new value for %s:"
	ModifyButton        string = "✏️ Modify %s"
	ChooseOption        string = "Choose an option:"
	Review              string = "📝 <b>Review Your Inputs:</b>\n\n"
	FileUploadSuccess   string = "File uploaded successfully!"
	UploadAnotherButton string = "Upload another file"
	UploadAnother       string = "Please upload another file"
	FinishUploadButton  string = "Finish uploading"
	FinishUpload        string = "Do you want to upload another file or finish uploading?"
	RequiredFile        string = "Oops! A file is required for %s. Please upload a file."
	RequiredSelect      string = "Oops! A selection is required. The input for %s must be one of the following options: %s. Please choose one."
	RequiredInput       string = "Oops! This %s is required. Please provide a value."
	InvalidEmail        string = "Oops! This doesn't look like a valid email address. Please check and try again."
	InvalidMaxNumber    string = "Oops! The value for %s must be at most %d. Please provide a valid number."
	InvalidMinNumber    string = "Oops! The value for %s must be at least %d. Please provide a valid number."
	InvalidNumber       string = "Oops! Please enter a valid number for %s."
	InvalidFormat       string = "Oops! The input for %s doesn't match the required format. Please make sure it’s correct."
	ValidationError     string = "Oops! Something went wrong while validating your input for %s. Please try again."
	InvalidMaxLength    string = "Oops! This input for %s is too long. Please provide no more than %d characters."
	InvalidMinLength    string = "Oops! This input for %s is too short. Please provide at least %d characters."
)

// Expected format placeholders for each message key
var expectedPlaceholders = map[string]int{
	"Modify":           1, // Requires 1 %s
	"ModifyButton":     1, // Requires 1 %s
	"RequiredFile":     1, // Requires 1 %s
	"RequiredSelect":   2, // Requires 2 (%s, %s)
	"RequiredInput":    1, // Requires 1 %s
	"InvalidMaxNumber": 2, // Requires 1 %s and 1 %d
	"InvalidMinNumber": 2, // Requires 1 %s and 1 %d
	"InvalidNumber":    1, // Requires 1 %s
	"InvalidFormat":    1, // Requires 1 %s
	"ValidationError":  1, // Requires 1 %s
	"InvalidMaxLength": 2, // Requires 1 %s and 1 %d
	"InvalidMinLength": 2, // Requires 1 %s and 1 %d
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

	tf.DefaultMessages()

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
func (f *Form) ValidateForm() ([]error, []string) {
	var errs []error

	// 1. Ensure form name and table name are non-empty
	if f.FormName == "" {
		errs = append(errs, fmt.Errorf("form_name cannot be empty"))
	}
	if f.TableName == "" {
		errs = append(errs, fmt.Errorf("table_name cannot be empty"))
	}

	// 2. Ensure ReviewEnabled is a bool (automatic in Go)
	if f.ReviewEnabled != true && f.ReviewEnabled != false {
		errs = append(errs, fmt.Errorf("review_enabled must be a boolean type"))
	}

	// 3. Validate each field and collect any errors
	for _, field := range f.Fields {
		fieldErrors := validateField(field)
		errs = append(errs, fieldErrors...)
	}

	// 4. DB validation (if applicable)
	if f.DB != "" {
		var err error
		for i, field := range f.Fields {
			if field.DBType != "" {
				f.Fields[i].ActualDBType, err = validateDBType(f.DB, field.DBType)
				if err != nil {
					errs = append(errs, fmt.Errorf("DB validation failed for field '%s': %v", field.Name, err))
				}
			}
		}
	}

	warnings := ValidateMessagePlaceholders(f.Messages)

	return errs, warnings
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

// ValidateMessagePlaceholders checks if the messages contain the correct number of placeholders
func ValidateMessagePlaceholders(msg Message) []string {
	msgValue := reflect.ValueOf(msg)
	msgType := reflect.TypeOf(msg)

	// Regex to find placeholders (%s, %d, %v)
	placeholderRegex := regexp.MustCompile(`%[sdv]`)

	msgs := make([]string, 0)

	for i := 0; i < msgValue.NumField(); i++ {
		field := msgValue.Field(i)
		fieldName := msgType.Field(i).Name

		// Ensure the field is a string and can be accessed
		if field.Kind() == reflect.String {
			messageText := field.String()

			// Count occurrences of %s, %d, %v
			matches := placeholderRegex.FindAllString(messageText, -1)
			placeholderCount := len(matches)

			// Get expected count from the map
			expectedCount, exists := expectedPlaceholders[fieldName]

			// If this message has a defined expected count, compare it
			if exists && expectedCount != placeholderCount {
				msgs = append(msgs, fmt.Sprintf("Message '%s' expects %d placeholder(s), but found %d.", fieldName, expectedCount, placeholderCount))
			}
		}
	}
	return msgs
}

// DefaultMessages returns a Message struct with default values
func (f *Form) DefaultMessages() {
	if f.Messages.Submit == "" {
		f.Messages.Submit = Submit
	}
	if f.Messages.SubmitButton == "" {
		f.Messages.SubmitButton = SubmitButton
	}
	if f.Messages.SkipButton == "" {
		f.Messages.SkipButton = SkipButton
	}
	if f.Messages.Modify == "" {
		f.Messages.Modify = Modify
	}
	if f.Messages.ModifyButton == "" {
		f.Messages.ModifyButton = ModifyButton
	}
	if f.Messages.ChooseOption == "" {
		f.Messages.ChooseOption = ChooseOption
	}
	if f.Messages.Review == "" {
		f.Messages.Review = Review
	}
	if f.Messages.FileUploadSuccess == "" {
		f.Messages.FileUploadSuccess = FileUploadSuccess
	}
	if f.Messages.UploadAnotherButton == "" {
		f.Messages.UploadAnotherButton = UploadAnotherButton
	}
	if f.Messages.UploadAnother == "" {
		f.Messages.UploadAnother = UploadAnother
	}
	if f.Messages.FinishUploadButton == "" {
		f.Messages.FinishUploadButton = FinishUploadButton
	}
	if f.Messages.FinishUpload == "" {
		f.Messages.FinishUpload = FinishUpload
	}
	if f.Messages.RequiredFile == "" {
		f.Messages.RequiredFile = RequiredFile
	}
	if f.Messages.RequiredSelect == "" {
		f.Messages.RequiredSelect = RequiredSelect
	}
	if f.Messages.RequiredInput == "" {
		f.Messages.RequiredInput = RequiredInput
	}
	if f.Messages.InvalidEmail == "" {
		f.Messages.InvalidEmail = InvalidEmail
	}
	if f.Messages.InvalidMaxNumber == "" {
		f.Messages.InvalidMaxNumber = InvalidMaxNumber
	}
	if f.Messages.InvalidMinNumber == "" {
		f.Messages.InvalidMinNumber = InvalidMinNumber
	}
	if f.Messages.InvalidNumber == "" {
		f.Messages.InvalidNumber = InvalidNumber
	}
	if f.Messages.InvalidFormat == "" {
		f.Messages.InvalidFormat = InvalidFormat
	}
	if f.Messages.ValidationError == "" {
		f.Messages.ValidationError = ValidationError
	}
	if f.Messages.InvalidMaxLength == "" {
		f.Messages.InvalidMaxLength = InvalidMaxLength
	}
	if f.Messages.InvalidMinLength == "" {
		f.Messages.InvalidMinLength = InvalidMinLength
	}
}
