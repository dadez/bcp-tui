// Package config
package config

import "github.com/spf13/viper"

// Load reads the config file and sets AppConfig
func Load(path string) error {
	viper.SetConfigFile(path)

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
