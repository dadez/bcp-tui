package config

type (
	Clusters []string
)

type Commands struct {
	Name    string `mapstructure:"name"`
	Command string `mapstructure:"command"`
}

type credentials struct {
	Username     string `mapstructure:"username"`
	Gopass       string `mapstructure:"username"`
	PassowrdPath string `mapstructure:"passowrdPath"`
}

type Config struct {
	Clusters    Clusters    `mapstructure:"clusters"`
	Commands    []Commands  `mapstructure:"commands"`
	credentials credentials `mapstructure:"credentials"`
}

// AppConfig exported package-level variable
var AppConfig *Config
