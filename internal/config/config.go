package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
}

var confBindings = []string{
	"server.host",
	"server.port",
	"database.host",
	"database.port",
	"database.user",
	"database.password",
	"database.dbname",
	"database.timeout",
	"server.jwt.secret",
}

func NewConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := bindEnvs(); err != nil {
		return nil, err
	}

	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func bindEnvs() error {
	for _, key := range confBindings {
		if err := viper.BindEnv(key); err != nil {
			return err
		}
	}

	return nil
}
