package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go-tg-support-ticket/bot"
	"go-tg-support-ticket/config"
	"go-tg-support-ticket/form"
	"go-tg-support-ticket/internal/store"
	"go-tg-support-ticket/logger"
	"go-tg-support-ticket/webhook"
	"os"
	"path/filepath"
	"runtime"
)

var botCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Telegram Bot",
	Run: func(cmd *cobra.Command, args []string) {

		cfg, err := config.LoadConfig(configFilePath)
		if err != nil {
			color.Set(color.FgRed)
			cmd.PrintErrf("❌ Error loading configuration: %v\n", err)
			color.Unset()
			return
		}

		logger.Init(cfg.DebugMode)

		if formatFilePath == "" {
			color.Set(color.FgYellow)
			cmd.Println("⚠️ No format file path provided. Showing help...")
			color.Unset()
			cmd.Help()
			return
		}
		tf, err := form.LoadTicketFormat(formatFilePath)
		if err != nil {
			color.Set(color.FgRed)
			cmd.PrintErrf("❌ Error loading json format from %s: %v\n", formatFilePath, err)
			color.Unset()
			return
		}
		errs := tf.ValidateForm()
		showValidationErrors(cmd, errs)

		if len(errs) == 0 {
			color.Set(color.FgGreen)
			cmd.Println("✅ The JSON file is valid!")
			color.Unset()
		}

		if cfg.Database.Enable && cfg.Database.UseAdaptor != tf.DB {
			color.Set(color.FgRed)
			cmd.PrintErrf("❌ Configuration mismatch: config uses %s database adaptor, but format specifies %s\n", cfg.Database.UseAdaptor, tf.DB)
			color.Unset()
			return
		}

		if cfg.EnableMemoryLoad {
			color.Set(color.FgGreen)
			cmd.Println("🔄 Memory load enabled. Trying to load photos...")

			for i := range tf.Fields {
				if tf.Fields[i].Type == "photo" && tf.Fields[i].Location != "" {
					color.Set(color.FgCyan)
					cmd.Printf("🔄 Loading photo: %s...\n", tf.Fields[i].Location)
					color.Unset()

					data, err := os.ReadFile(filepath.Clean(tf.Fields[i].Location))
					if err != nil {
						color.Set(color.FgRed)
						cmd.PrintErrf("❌ Failed to load %s: %v\n", tf.Fields[i].Location, err)
						color.Unset()
						return
					}

					tf.Fields[i].PhotoData = data
					color.Set(color.FgGreen)
					cmd.Printf("✅ Successfully loaded %s\n", tf.Fields[i].Location)
					color.Unset()

					// Check memory usage after loading each photo
					if getMemoryUsageMB() >= cfg.MemoryLimitMB {
						color.Set(color.FgYellow)
						cmd.Printf("⚠️ Memory limit reached mid-load. Stopping further preloading!\n")
						color.Unset()
						break
					}
				}
			}
		}

		if cfg.Database.Enable {
			if err := store.Store.Open(cfg.Database); err != nil {
				color.Set(color.FgRed)
				cmd.PrintErrf("❌ Failed to connect to the database: %v\n", err)
				color.Unset()
				return
			} else {
				color.Set(color.FgGreen)
				cmd.Println("✅ Connected to the database successfully.")
				color.Unset()
			}
		}

		webhook.NewWebhookWorker(cfg.Webhook)

		b, err := bot.NewBot(cfg.Bot, tf)
		if err != nil {
			color.Set(color.FgRed)
			cmd.PrintErrf("❌ Error initializing bot: %v\n", err)
			color.Unset()
			return
		}

		color.Set(color.FgGreen)
		cmd.Println("✅ Bot started successfully!")
		color.Unset()

		b.Start()
	},
}

func init() {
	rootCmd.AddCommand(botCmd)
	botCmd.Flags().StringVarP(&formatFilePath, "file", "f", "", "Path to format JSON file")
	botCmd.Flags().StringVarP(&configFilePath, "config", "c", "config.yaml", "Path to config JSON file")
}

// Check memory usage
func getMemoryUsageMB() int64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return int64(memStats.Alloc) / (1024 * 1024) // Convert bytes to MB
}
