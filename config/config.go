package config

type (
	Clusters []string
	Actions  []string
)

type Commands struct {
	Name    string `mapstructure:"name"`
	Command string `mapstructure:"command"`
}

type Urls struct {
	Name string `mapstructure:"name"`
	URL  string `mapstructure:"url"`
}
type credentials struct {
	Username     string `mapstructure:"username"`
	Gopass       string `mapstructure:"username"`
	PassowrdPath string `mapstructure:"passowrdPath"`
}

type Config struct {
	Clusters    Clusters    `mapstructure:"clusters"`
	Actions     Actions     `mapstructure:"actions"`
	Commands    []Commands  `mapstructure:"commands"`
	Urls        []Urls      `mapstructure:"urls"`
	credentials credentials `mapstructure:"credentials"`
}

// AppConfig exported package-level variable
var AppConfig *Config
