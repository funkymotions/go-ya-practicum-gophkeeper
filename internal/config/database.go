package config

import "fmt"

type Database struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBName      string `mapstructure:"dbname"`
	ConnTimeout int    `mapstructure:"timeout"`
}

func (d *Database) GetDSN() string {
	dsn := "postgres://" + d.User + ":" + d.Password + "@" + d.Host + ":" + fmt.Sprint(d.Port) + "/" + d.DBName + "?sslmode=disable"
	fmt.Printf("DSN: %s\n", dsn)
	return dsn
}
