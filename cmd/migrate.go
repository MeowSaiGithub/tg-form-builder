package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go-tg-support-ticket/config"
	"go-tg-support-ticket/form"
	"go-tg-support-ticket/internal/store"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
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
			cmd.PrintErrf("❌ Error loading ticket format from %s: %v\n", formatFilePath, err)
			color.Unset()
			return
		}

		cfg, err := config.LoadConfig(configFilePath)
		if err != nil {
			color.Set(color.FgRed)
			cmd.PrintErrf("❌ Error loading configuration: %v\n", err)
			color.Unset()
			return
		}

		if cfg.Database.Enable && cfg.Database.UseAdaptor != tf.DB {
			color.Set(color.FgRed)
			cmd.PrintErrf("❌ Database configuration mismatch: config uses %s adaptor, but format specifies %s DB\n", cfg.Database.UseAdaptor, tf.DB)
			color.Unset()
			return
		}

		if !cfg.Database.Enable {
			color.Set(color.FgYellow)
			cmd.Println("⚠️ Database is disabled in the configuration.")
			color.Unset()
			return
		}

		color.Set(color.FgGreen)
		cmd.Println("✅ Starting database connection for migration...")
		color.Unset()

		if err := store.Store.Open(cfg.Database); err != nil {
			color.Set(color.FgRed)
			cmd.PrintErrf("❌ Failed to connect to database: %v\n", err)
			color.Unset()
			return
		}

		color.Set(color.FgGreen)
		cmd.Println("✅ Database connected successfully.")
		color.Unset()

		color.Set(color.FgGreen)
		cmd.Println("🔄 Running database migration...")
		color.Unset()

		errs := tf.ValidateForm()
		if errs != nil {
			showValidationErrors(cmd, errs)
			return
		}

		if err := store.Store.Migrate(tf); err != nil {
			color.Set(color.FgRed)
			cmd.PrintErrf("❌ Migration failed: %v\n", err)
			color.Unset()
			return
		}

		color.Set(color.FgGreen)
		cmd.Println("✅ Database migration completed successfully!")
		color.Unset()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&formatFilePath, "file", "f", "", "Path to format JSON file")
	migrateCmd.Flags().StringVarP(&configFilePath, "config", "c", "config.yaml", "Path to config JSON file")
}
