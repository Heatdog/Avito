package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

type ConfigStorage struct {
	Server      ServerListen   `mapstructure:"server_listen"`
	Postgre     PostgreStorage `mapstructure:"postgre_settings"`
	PasswordKey string         `mapstructure:"password_key"`
}

type ServerListen struct {
	IP   string `mapstructure:"ip"`
	Port string `mapstructure:"port"`
}

type PostgreStorage struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func NewConfigStorage(logger *slog.Logger) *ConfigStorage {
	logger.Debug("reading log file")
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		logger.Error("config file reading failed", slog.Any("error", err))
	}

	res := &ConfigStorage{}
	logger.Debug("unmarshaling log file")
	if err := viper.Unmarshal(res); err != nil {
		logger.Error("unmarshaling config file failed", slog.Any("error", err))
	}
	return res
}
