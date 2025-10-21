package config

type (
	Clusters []string
)

type Commands struct {
	Name    string `mapstructure:"name"`
	Command string `mapstructure:"command"`
}

type Config struct {
	Clusters Clusters   `mapstructure:"clusters"`
	Commands []Commands `mapstructure:"commands"`
}
