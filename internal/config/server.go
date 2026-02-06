package config

type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	JWT  JWT    `mapstructure:"jwt"`
}
