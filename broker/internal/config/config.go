package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var (
	Cfg AppConfig
)

type AppConfig struct {
	App struct {
		Name            string `mapstructure:"name"`
		SqliteDirectory string `mapstructure:"sqlite_directory"`
		LogDirectory    string `mapstructure:"log_directory"`
		Environment     string `mapstructure:"environment"` // "debug" or "prod"
	} `mapstructure:"app"`
	Server struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"server"`
}

func Initialize() error {
	viper.SetConfigFile("./config.yaml")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("env")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("error reading config file, %w", err)
	}

	viper.BindEnv("app.name", "APP_NAME")

	err = viper.Unmarshal(&Cfg)
	if err != nil {
		return fmt.Errorf("unable to decode into struct, %v", err)
	}

	return nil
}
