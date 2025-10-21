// Package config
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var AppConfig *Config

// Load reads the config file and sets AppConfig
func Load(configPath string) error {
	if configPath != "" {
		if filepath.Ext(configPath) != "" {
			viper.SetConfigFile(configPath)
		} else {
			viper.AddConfigPath(configPath)
			viper.SetConfigName("config")
		}
	} else {
		// No flag passed, defaults
		viper.SetConfigName("config")
		viper.AddConfigPath(".")

		// XDG config path if set
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			viper.AddConfigPath(filepath.Join(xdg, "bcp-tui"))
		} else {
			viper.AddConfigPath("$HOME/.config/bcp-tui")
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	AppConfig = &cfg // <- THIS IS VALID because AppConfig is defined in the same package
	return nil
}
