package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go-tg-support-ticket/form"
	"go-tg-support-ticket/internal/store"
	"go-tg-support-ticket/logger"
	"go-tg-support-ticket/webhook"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Token string `mapstructure:"token" validate:"required"`
}

type Bot struct {
	api                   *tgbotapi.BotAPI
	format                *form.Form
	userStates            sync.Map // Stores user step (int64 -> int)
	userModificationState sync.Map // Tracks modifying field (int64 -> string)
	userTimers            sync.Map // Stores user inactivity timers (int64 -> *time.Timer)
	sessionTimeout        time.Duration
	authLinks             sync.Map // Authentication links (int64 -> string)
	userAuthStatus        sync.Map // Auth status (int64 -> bool)
	mu                    sync.Mutex
}

func NewBot(cfg *Config, format *form.Form) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new bot: %w", err)
	}

	b := &Bot{
		api:            api,
		format:         format,
		sessionTimeout: 30 * time.Minute,
	}

	if err := b.SetCommands(); err != nil {
		return nil, fmt.Errorf("failed to set commands: %w", err)
	}
	return b, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			if update.Message.IsCommand() {
				command := update.Message.Command()
				switch command {
				case "start":
					b.userStates.Store(update.Message.Chat.ID, 0)
					b.generateFormStep(update.Message.Chat.ID)
				case "end":
					b.endSession(update.Message.Chat.ID)
				case "help":
					b.sendHelpMessage(update.Message.Chat.ID)
				default:
					b.handleUserInput(update)
				}
			} else {
				b.handleUserInput(update)
			}
		} else if update.CallbackQuery != nil { // If we got a callback query
			b.handleCallbackQuery(update)
		}
	}
}

func (b *Bot) sendHelpMessage(chatID int64) {
	helpText := `Welcome to the bot! Here are the available commands:
/start - Start a new session
/end - End the current session
/help - Show this help message`
	if _, err := b.api.Send(tgbotapi.NewMessage(chatID, helpText)); err != nil {
		logger.PrintLog(chatID, "failed to send help message", err)
	}
}

func (b *Bot) SetCommands() error {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Start a new session"},
		{Command: "end", Description: "End the current session"},
		{Command: "help", Description: "Show help message"},
	}

	_, err := b.api.Request(tgbotapi.NewSetMyCommands(commands...))
	return err
}

func (b *Bot) generateFormStep(chatID int64) {
	stepA, _ := b.userStates.LoadOrStore(chatID, 0)
	step := stepA.(int)

	if step >= len(b.format.Fields) {
		if b.format.ReviewEnabled {
			b.sendReviewMessage(chatID, b.format)
		} else {
			b.submitForm(chatID)
		}
		return
	}

	field := b.format.Fields[step]

	// Function to apply message formatting
	applyFormatting := func(msg *tgbotapi.MessageConfig) {
		if field.Formatting == "Markdown" {
			msg.ParseMode = tgbotapi.ModeMarkdown
		} else if field.Formatting == "HTML" {
			msg.ParseMode = tgbotapi.ModeHTML
		}
	}

	// Send media based on type (photo, video, or document)
	var err error
	switch field.Type {
	case "photo":
		err = b.sendPhoto(chatID, field)
	case "video":
		err = b.sendVideo(chatID, field)
	case "document":
		err = b.sendDocument(chatID, field)
	default:
		// Send a text message if no valid type found
		msg := tgbotapi.NewMessage(chatID, field.Description)
		applyFormatting(&msg)

		_, err = b.api.Send(msg)
	}

	if err != nil {
		logger.PrintLog(chatID, "failed to send media/text message", err)
		// Notify the user about the error
		errorMsg := tgbotapi.NewMessage(chatID, "Failed to send the content. Please try again later.")
		if _, sendErr := b.api.Send(errorMsg); sendErr != nil {
			logger.PrintLog(chatID, "failed to send error message", sendErr)
		}
		b.clearUserSession(chatID)
		return
	}

	// Add buttons for the field
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, button := range field.Buttons {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(button.Text, button.Data),
		))
	}

	// Add a "Skip" button if the field is skippable
	if field.Skippable {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(b.format.Messages.SkipButton, "skip"),
		))
	}

	// Send inline buttons if there are any
	if len(rows) > 0 {
		msg := tgbotapi.NewMessage(chatID, b.format.Messages.ChooseOption)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		if _, err := b.api.Send(msg); err != nil {
			logger.PrintLog(chatID, "failed to send inline keyboard", err)
		}
	}
}

func (b *Bot) sendReviewMessage(chatID int64, config *form.Form) {
	var reviewText strings.Builder
	reviewText.WriteString(b.format.Messages.Review + "\n\n")
	for _, field := range config.Fields {
		value := field.UserValue
		if value == "" {
			value = "Not provided"
		}
		reviewText.WriteString(fmt.Sprintf("<b>%s:</b> %s\n", field.Label, value))
	}

	// Add buttons for each field to allow modification
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, field := range config.Fields {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf(b.format.Messages.ModifyButton, field.Label), fmt.Sprintf("modify_%s", field.Name)),
		))
	}

	// Add a submit button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		//tgbotapi.NewInlineKeyboardButtonData("✅ Submit", "submit"),
		tgbotapi.NewInlineKeyboardButtonData(b.format.Messages.SubmitButton, "submit"),
	))

	msg := tgbotapi.NewMessage(chatID, reviewText.String())
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	if _, err := b.api.Send(msg); err != nil {
		logger.PrintLog(chatID, "failed to send review message", err)
	}
}

func (b *Bot) submitForm(chatID int64) {

	if err := store.Tickets.Create(b.format.TableName, b.format.Fields); err != nil {
		logger.PrintLog(chatID, "failed to create form", err)
	}

	//text := "🎉 Thank you for submitting the form! 🎉"
	text := b.format.Messages.Submit

	// Process the form submission (e.g., save to database)
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		logger.PrintLog(chatID, "failed to send submitting message", err)
	}

	if webhook.Workers != nil {
		webhook.Workers.Enqueue(b.format)
	}

	// Clear the user session after submission
	b.clearUserSession(chatID)
}

func (b *Bot) clearUserSession(chatID int64) {
	b.userStates.Delete(chatID)
	b.userModificationState.Delete(chatID)
	if timer, ok := b.userTimers.Load(chatID); ok {
		timer.(*time.Timer).Stop()
		b.userTimers.Delete(chatID)
	}
}

func (b *Bot) resetInactivityTimer(chatID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if timer, ok := b.userTimers.Load(chatID); ok {
		timer.(*time.Timer).Stop()
	}
	timer := time.AfterFunc(b.sessionTimeout, func() {
		b.endSession(chatID)
	})
	b.userTimers.Store(chatID, timer)
}

// validateField validates the user input for a field based on its validation rules.
func (b *Bot) validateField(field form.Field, text string) (string, error) {
	value := text

	// Skip validation if the field is skippable and the value is empty
	if field.Skippable && value == "" {
		return "", nil
	}

	// Check if the field is required and the value is empty
	if field.Required && value == "" {
		//userMsg := fmt.Sprintf("Oops! The input for %s is required. Please provide a value.", field.Name)
		userMsg := fmt.Sprintf(b.format.Messages.RequiredInput, field.Name)
		logMsg := fmt.Errorf("validation error for %s: input is required but was not provided", field.Name)
		return userMsg, logMsg
	}

	// Validate based on the field type
	switch field.Type {
	case "text":
		return b.validateTextField(field, value)
	case "number":
		return b.validateNumberField(field, value)
	case "email":
		return b.validateEmailField(field, value)
	case "select":
		return b.validateSelectField(field, value)
	case "file":
		return b.validateFileField(field, value)
	default:
		userMsg := fmt.Sprintf("Oops! Unsupport field type")
		logMsg := fmt.Errorf("unsupported field type: %s", field.Type)
		return userMsg, logMsg
	}
}

// handleUserInput handles the input for each field
func (b *Bot) handleUserInput(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	// Reset the inactivity timer for the user
	b.resetInactivityTimer(chatID)

	// Handle file uploads (photos, documents, videos)
	if update.Message.Document != nil || update.Message.Photo != nil || update.Message.Video != nil {
		b.handleFileUpload(update)
		return
	}

	text := update.Message.Text

	// Check if the user is modifying a field
	if fieldName, ok := b.userModificationState.Load(chatID); ok {
		for i, field := range b.format.Fields {
			if field.Name == fieldName {

				// Validate the user input
				if msg, err := b.validateField(field, text); err != nil {
					msg := tgbotapi.NewMessage(chatID, msg)
					if _, err := b.api.Send(msg); err != nil {
						logger.PrintLog(chatID, "failed to send error message", err)
					}
					return
				}

				// Store the validated input
				b.format.Fields[i].UserValue = text
				break
			}
		}
		b.userModificationState.Delete(chatID) // Clear modification state
		b.sendReviewMessage(chatID, b.format)
		return
	}

	// Handle input for the current step
	stepA, _ := b.userStates.Load(chatID)
	step := stepA.(int)
	if step < len(b.format.Fields) {
		field := b.format.Fields[step]

		// Validate the user input
		if msg, err := b.validateField(field, text); err != nil {
			logger.PrintLog(chatID, "user input validation", err)
			msg := tgbotapi.NewMessage(chatID, msg)
			if _, err := b.api.Send(msg); err != nil {
				logger.PrintLog(chatID, "failed to send error message", err)
			}
			return
		}

		// Store the validated input
		b.format.Fields[step].UserValue = text
		b.userStates.Store(chatID, step+1)
		b.generateFormStep(chatID)
	}
}

// endSession allows users to end their current session
func (b *Bot) endSession(chatID int64) {
	b.clearUserSession(chatID)
	msg := tgbotapi.NewMessage(chatID, "Your session has been ended. Thank you for using the bot.")
	if _, err := b.api.Send(msg); err != nil {
		logger.PrintLog(chatID, "failed to send end session message", err)
	}
}

func (b *Bot) handleCallbackQuery(update tgbotapi.Update) {

	query := update.CallbackQuery
	chatID := query.Message.Chat.ID

	// Reset the inactivity timer for the user
	b.resetInactivityTimer(chatID)

	switch {
	case query.Data == "skip":
		if stepValue, ok := b.userStates.Load(chatID); ok {
			step := stepValue.(int)
			if step < len(b.format.Fields) {
				b.format.Fields[step].UserValue = "skipped"
				b.userStates.Store(chatID, step+1)
				b.generateFormStep(chatID)
			}
		}
	case query.Data == "submit":
		b.submitForm(chatID)
	case strings.HasPrefix(query.Data, "modify_"):
		fieldName := strings.TrimPrefix(query.Data, "modify_")
		b.userModificationState.Store(chatID, fieldName) // Set the field to modify
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Please enter a new value for %s:", fieldName))
		if _, err := b.api.Send(msg); err != nil {
			logger.PrintLog(chatID, "failed to send modify message", err)
		}
	case query.Data == "upload_another":
		// User wants to upload another file, so we continue the current form step
		//msg := tgbotapi.NewMessage(chatID, "Please upload another file.")
		msg := tgbotapi.NewMessage(chatID, b.format.Messages.UploadAnother)
		if _, err := b.api.Send(msg); err != nil {
			logger.PrintLog(chatID, "failed to send upload prompt message", err)
		}
	case query.Data == "finish_uploading":
		// User finished uploading, continue with the form
		st, _ := b.userStates.Load(chatID)
		b.userStates.Store(chatID, st.(int)+1) // Move to the next step
		b.generateFormStep(chatID)
	default:
		if stepValue, ok := b.userStates.Load(chatID); ok {
			step := stepValue.(int)
			if step < len(b.format.Fields) {
				b.format.Fields[step].UserValue = query.Data
				b.userStates.Store(chatID, step+1)
				b.generateFormStep(chatID)
			}
		}
	}
}

// validateTextField validates a text field with both user and log messages.
func (b *Bot) validateTextField(field form.Field, value string) (string, error) {
	// Check min and max length (if specified)
	if field.Validation.MinLength > 0 && len(value) < field.Validation.MinLength {
		//userMsg := fmt.Sprintf("Oops! The input for %s is too short. Please provide at least %d characters.", field.Label, field.Validation.MinLength)
		userMsg := fmt.Sprintf(b.format.Messages.InvalidMinLength, field.Label, field.Validation.MinLength)
		logMsg := fmt.Errorf("validation error for %s: input is too short. Expected at least %d characters, got %d characters", field.Label, field.Validation.MinLength, len(value))
		return userMsg, logMsg
	}
	if field.Validation.MaxLength > 0 && len(value) > field.Validation.MaxLength {
		//userMsg := fmt.Sprintf("Oops! The input for %s is too long. Please provide no more than %d characters.", field.Label, field.Validation.MaxLength)
		userMsg := fmt.Sprintf(b.format.Messages.InvalidMaxLength, field.Label, field.Validation.MaxLength)
		logMsg := fmt.Errorf("validation error for %s: input is too long. Expected no more than %d characters, got %d characters", field.Label, field.Validation.MaxLength, len(value))
		return userMsg, logMsg
	}

	// Check regex pattern (if specified)
	if field.Validation.Regex != "" {
		matched, err := regexp.MatchString(field.Validation.Regex, value)
		if err != nil {
			//userMsg := fmt.Sprintf("Oops! Something went wrong while validating your input for %s. Please try again.", field.Label)
			userMsg := fmt.Sprintf(b.format.Messages.ValidationError, field.Label)
			logMsg := fmt.Errorf("regex error for %s: %v", field.Label, err)
			return userMsg, logMsg
		}
		if !matched {
			//userMsg := fmt.Sprintf("Oops! The input for %s doesn't match the required format. Please make sure it’s correct.", field.Label)
			userMsg := fmt.Sprintf(b.format.Messages.InvalidFormat, field.Label)
			logMsg := fmt.Errorf("validation error for %s: input '%s' does not match required regex '%s'", field.Label, value, field.Validation.Regex)
			return userMsg, logMsg
		}
	}

	return "", nil
}

// validateNumberField validates a number field with both user and log messages.
func (b *Bot) validateNumberField(field form.Field, value string) (string, error) {
	// Convert the value to a number
	num, err := strconv.Atoi(value)
	if err != nil {
		//userMsg := fmt.Sprintf("Oops! The input for %s must be a valid number. Please provide a valid number.", field.Label)
		userMsg := fmt.Sprintf(b.format.Messages.InvalidNumber, field.Label)
		logMsg := fmt.Errorf("validation error for %s: failed to convert '%s' to a number. Error: %v", field.Label, value, err)
		return userMsg, logMsg
	}

	// Check min and max values (if specified)
	if field.Validation.Min > 0 && num < field.Validation.Min {
		//userMsg := fmt.Sprintf("Oops! The input for %s must be at least %d. Please provide a valid number.", field.Label, field.Validation.Min)
		userMsg := fmt.Sprintf(b.format.Messages.InvalidMinNumber, field.Label, field.Validation.Min)
		logMsg := fmt.Errorf("validation error for %s: input %d is less than the minimum %d", field.Label, num, field.Validation.Min)
		return userMsg, logMsg
	}
	if field.Validation.Max > 0 && num > field.Validation.Max {
		//userMsg := fmt.Sprintf("Oops! The input for %s must be at most %d. Please provide a valid number.", field.Label, field.Validation.Max)
		userMsg := fmt.Sprintf(b.format.Messages.InvalidMaxNumber, field.Label, field.Validation.Max)
		logMsg := fmt.Errorf("validation error for %s: input %d exceeds the maximum %d", field.Label, num, field.Validation.Max)
		return userMsg, logMsg
	}

	return "", nil
}

// validateEmailField validates an email field with both user and log messages.
func (b *Bot) validateEmailField(field form.Field, value string) (string, error) {
	// Simple email regex for validation
	emailRegex := `^[a-z0-9]+@[a-z0-9]+\.[a-z]{2,3}$`
	matched, err := regexp.MatchString(emailRegex, value)
	if err != nil {
		//userMsg := fmt.Sprintf("Oops! Something went wrong while validating your email. Please try again.")
		userMsg := fmt.Sprintf(b.format.Messages.InvalidEmail)
		logMsg := fmt.Errorf("email validation error for %s: regex failed with error: %v", field.Label, err)
		return userMsg, logMsg
	}
	if !matched {
		//userMsg := fmt.Sprintf("Oops! The input for %s doesn't look like a valid email address. Please check and try again.", field.Label)
		userMsg := fmt.Sprintf(b.format.Messages.InvalidEmail)
		logMsg := fmt.Errorf("validation error for %s: email '%s' does not match valid format", field.Label, value)
		return userMsg, logMsg
	}
	return "", nil
}

// validateSelectField validates a select field with both user and log messages.
func (b *Bot) validateSelectField(field form.Field, value string) (string, error) {
	// Check if the value is one of the allowed options
	for _, option := range field.Options {
		if value == option {
			return "", nil
		}
	}
	//userMsg := fmt.Sprintf("Oops! The input for %s must be one of the following options: %s. Please choose one.", field.Label, strings.Join(field.Options, ", "))
	userMsg := fmt.Sprintf(b.format.Messages.ChooseOption, field.Label, strings.Join(field.Options, ", "))
	logMsg := fmt.Errorf("validation error for %s: invalid option '%s'. Expected one of: %s", field.Label, value, strings.Join(field.Options, ", "))
	return userMsg, logMsg
}

// validateFileField validates a file field with both user and log messages.
func (b *Bot) validateFileField(field form.Field, value string) (string, error) {
	// For file fields, we can check if the value is a valid file path or URL
	if value == "" && field.Required {
		//userMsg := fmt.Sprintf("Oops! The input for %s is required. Please upload a file.", field.Label)
		userMsg := fmt.Sprintf(b.format.Messages.RequiredFile, field.Label)
		logMsg := fmt.Errorf("validation error for %s: file is required but was not provided", field.Label)
		return userMsg, logMsg
	}
	return "", nil
}

// handleFileUpload handles the file upload logic for multiple files
func (b *Bot) handleFileUpload(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	// Get the current step
	stepA, _ := b.userStates.Load(chatID)
	step := stepA.(int)
	if step >= len(b.format.Fields) {
		return
	}

	field := b.format.Fields[step]

	// Ensure the field is of type "file"
	if field.Type != "file" {
		logger.PrintLog(chatID, "invalid field type for file upload", fmt.Errorf("expected file type, got %s", field.Type))
		return
	}

	// Initialize a variable to store the highest resolution file URL
	var fileURL string

	// Process documents
	if update.Message.Document != nil {
		// Handle a single document (document is always one file)
		fileURL, _ = b.api.GetFileDirectURL(update.Message.Document.FileID)
	}

	// Process photos (get the highest resolution)
	if len(update.Message.Photo) > 0 {
		// Telegram provides multiple photo resolutions, so we take the highest one (last one)
		// Location array is sorted by resolution (first is the smallest, last is the largest)
		photo := update.Message.Photo[len(update.Message.Photo)-1] // Take the last one (highest resolution)
		fileURL, _ = b.api.GetFileDirectURL(photo.FileID)
	}

	// Process videos (handle video uploads)
	if update.Message.Video != nil {
		// Handle a single video (only one file)
		fileURL, _ = b.api.GetFileDirectURL(update.Message.Video.FileID)
	}

	// If no valid files are found, log an error and return
	if fileURL == "" {
		logger.PrintLog(chatID, "no valid file found in the message", fmt.Errorf("no valid file found"))
		return
	}

	// Append the file URL to the current field's user value (multiple files allowed)
	if field.UserValue == "" {
		field.UserValue = fileURL
	} else {
		field.UserValue += "," + fileURL // Separate multiple file URLs with a comma
	}

	// Update the field in the form
	b.format.Fields[step].UserValue = field.UserValue

	// Notify the user that the file was uploaded successfully
	//msg := tgbotapi.NewMessage(chatID, "File uploaded successfully!")
	msg := tgbotapi.NewMessage(chatID, b.format.Messages.FileUploadSuccess)
	if _, err := b.api.Send(msg); err != nil {
		logger.PrintLog(chatID, "failed to send success message", err)
	}

	// Prompt the user to either upload another file or finish uploading
	// Add buttons for the field
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		//tgbotapi.NewInlineKeyboardButtonData("Upload another", "upload_another"),
		//tgbotapi.NewInlineKeyboardButtonData("Finish uploading", "finish_uploading"),
		tgbotapi.NewInlineKeyboardButtonData(b.format.Messages.UploadAnotherButton, "upload_another"),
		tgbotapi.NewInlineKeyboardButtonData(b.format.Messages.FinishUploadButton, "finish_uploading"),
	))

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	// Send the prompt message with buttons
	//msgPrompt := tgbotapi.NewMessage(chatID, "Do you want to upload another file or finish uploading?")
	msgPrompt := tgbotapi.NewMessage(chatID, b.format.Messages.FinishUpload)
	msgPrompt.ReplyMarkup = inlineKeyboard
	if _, err := b.api.Send(msgPrompt); err != nil {
		logger.PrintLog(chatID, "failed to send upload prompt", err)
	}
}

func (b *Bot) sendPhoto(chatID int64, field form.Field) error {
	var photoBytes []byte
	var err error

	// Read photo from the provided location
	if field.Location != "" {
		photoBytes, err = os.ReadFile(field.Location)
		if err != nil {
			return fmt.Errorf("failed to load photo from location: %s; %w", field.Location, err)
		}
	}

	photoFile := tgbotapi.FileBytes{Name: field.Location, Bytes: photoBytes}
	photoMsg := tgbotapi.NewPhoto(chatID, photoFile)
	photoMsg.Caption = field.Description

	if field.Formatting == "Markdown" {
		photoMsg.ParseMode = tgbotapi.ModeMarkdown
	} else if field.Formatting == "HTML" {
		photoMsg.ParseMode = tgbotapi.ModeHTML
	}

	_, err = b.api.Send(photoMsg)
	return err
}

func (b *Bot) sendVideo(chatID int64, field form.Field) error {
	videoFile := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(field.Location))
	videoFile.Caption = field.Description

	if field.Formatting == "Markdown" {
		videoFile.ParseMode = tgbotapi.ModeMarkdown
	} else if field.Formatting == "HTML" {
		videoFile.ParseMode = tgbotapi.ModeHTML
	}

	_, err := b.api.Send(videoFile)
	if err != nil {
		return fmt.Errorf("failed to send video from location: %s; error: %w", field.Location, err)
	}
	return err
}

func (b *Bot) sendDocument(chatID int64, field form.Field) error {
	docFile := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(field.Location))
	docFile.Caption = field.Description

	if field.Formatting == "Markdown" {
		docFile.ParseMode = tgbotapi.ModeMarkdown
	} else if field.Formatting == "HTML" {
		docFile.ParseMode = tgbotapi.ModeHTML
	}

	_, err := b.api.Send(docFile)
	if err != nil {
		return fmt.Errorf("failed to send document from location: %s; error %w", field.Location, err)
	}
	return nil
}
