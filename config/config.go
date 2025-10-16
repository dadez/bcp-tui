package config

// Config struct for your clusters list
type Config struct {
	Clusters []string `mapstructure:"clusters"`
	Actions  []string `mapstructure:"actions"`
}

// AppConfig exported package-level variable
var AppConfig *Config
