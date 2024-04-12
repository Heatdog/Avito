package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

type Settings struct {
	Server      ServerListen    `mapstructure:"server_listen"`
	Postgre     PostgreSettings `mapstructure:"postgre_settings"`
	Cache       CacheSettings   `mapstructure:"cache_settings"`
	Redis       RedisSettings   `mapstructure:"redis_settings"`
	PasswordKey string          `mapstructure:"password_key"`
}

type ServerListen struct {
	IP   string `mapstructure:"ip"`
	Port int    `mapstructure:"port"`
}

type PostgreSettings struct {
	Host     string `mapstructure:"host"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`

	TimePrepare int `mapstructure:"time_prepare"`
	TimeWait    int `mapstructure:"time_wait"`
}

type CacheSettings struct {
	Size int `mapstructure:"host"`
	TTL  int `mapstructure:"ttl_in_minutes"`
}

type RedisSettings struct {
	Host        string `mapstructure:"host"`
	Password    string `mapstructure:"password"`
	Database    string `mapstructure:"database"`
	Port        int    `mapstructure:"port"`
	TimePrepare int    `mapstructure:"time_prepare"`
}

func NewConfigStorage(logger *slog.Logger) *Settings {
	logger.Debug("reading log file")
	viper.SetConfigFile("config.yaml")

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("config file reading failed", slog.Any("error", err))
	}

	res := &Settings{}

	logger.Debug("unmarshaling log file")

	if err := viper.Unmarshal(res); err != nil {
		logger.Error("unmarshaling config file failed", slog.Any("error", err))
	}

	return res
}
