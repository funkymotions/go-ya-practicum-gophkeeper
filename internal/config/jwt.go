package config

type JWT struct {
	SecretKey string `mapstructure:"secret"`
}
