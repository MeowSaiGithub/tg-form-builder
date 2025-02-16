package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
)

var formatFilePath string
var configFilePath string

var rootCmd = &cobra.Command{
	Use:   "gotgbot",
	Short: "A CLI tool for creating dynamic form telegram bot",
	Long:  "A CLI tool for creating dynamic form telegram bot based on JSON configuration",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Set(color.FgRed)
		fmt.Printf("❌ Error executing command: %v\n", err)
		color.Unset()
		os.Exit(1)
	}
}
