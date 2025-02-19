package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go-tg-support-ticket/form"
)

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&formatFilePath, "file", "f", "", "Path to format JSON file")
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a JSON file",
	Run: func(cmd *cobra.Command, args []string) {

		if formatFilePath == "" {
			color.Set(color.FgYellow)
			cmd.Println("⚠️ Format file path is missing. Showing help...")
			color.Unset()
			cmd.Help()
			return
		}

		tf, err := form.LoadTicketFormat(formatFilePath)
		if err != nil {
			color.Set(color.FgRed)
			cmd.PrintErrf("❌ Invalid JSON file: %v\n", err)
			color.Unset()
			return
		}

		errs, warnings := tf.ValidateForm()
		showValidationWarnings(cmd, warnings)
		showValidationErrors(cmd, errs)

		if len(errs) == 0 {
			color.Set(color.FgGreen)
			cmd.Println("✅ The JSON file is valid!")
			color.Unset()
		}

	},
}

func showValidationErrors(cmd *cobra.Command, errors []error) {
	if len(errors) > 0 {
		// Show all errors in a formatted way
		color.Set(color.FgRed)
		for _, err := range errors {
			cmd.PrintErrf("❌ %s\n", err)
		}
		color.Unset()
	} else {
		color.Set(color.FgGreen)
		cmd.Println("✅ All validations passed successfully!")
		color.Unset()
	}
}

func showValidationWarnings(cmd *cobra.Command, warnings []string) {
	if len(warnings) > 0 {
		// Show all warnings in a formatted way
		color.Set(color.FgYellow)
		for _, warning := range warnings {
			cmd.PrintErrf("⚠️ Warning: %s\n", warning)
		}
		color.Unset()
	}
}
