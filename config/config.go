package config

import (
	"fmt"
	"github.com/spf13/viper"
	"go-tg-support-ticket/bot"
	"go-tg-support-ticket/webhook"

	"go-tg-support-ticket/internal/database"
)

type Config struct {
	DebugMode bool `mapstructure:"debug_mode"`

	EnableMemoryLoad bool  `mapstructure:"enable_memory_load"`
	MemoryLimitMB    int64 `mapstructure:"memory_limit_mb"`

	Bot      *bot.Config      `mapstructure:"bot" validate:"required"`
	Database *database.Config `mapstructure:"database"`
	Webhook  *webhook.Config  `mapstructure:"webhook"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &cfg, nil
}
