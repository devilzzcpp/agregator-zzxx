package config

import (
	"fmt"

	"github.com/devilzzcpp/agregator-zzxx/common"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	//сервер
	Host string
	Port string

	//бд
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	TimeZone string
}

// DSN собирает строку подключения к postgreSQL для GORM
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode, c.TimeZone,
	)
}

var Cfg *Config

func setDefaults() {
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "8080")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")

	viper.SetDefault("APP_TIMEZONE", "Europe/Samara")
}

func LoadConfig() (*Config, error) {
	setDefaults()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	usedDefaults := false
	if err := viper.ReadInConfig(); err != nil {
		usedDefaults = true
		if common.Logger != nil {
			common.Logger.Warn(".env не прочитан, используется env + значения по умолчанию", zap.Error(err))
		}
	}

	cfg := &Config{
		Host: viper.GetString("HOST"),
		Port: viper.GetString("PORT"),

		DBHost:     viper.GetString("DB_HOST"),
		DBPort:     viper.GetString("DB_PORT"),
		DBUser:     viper.GetString("DB_USER"),
		DBPassword: viper.GetString("DB_PASSWORD"),
		DBName:     viper.GetString("DB_NAME"),
		DBSSLMode:  viper.GetString("DB_SSLMODE"),

		TimeZone: viper.GetString("APP_TIMEZONE"),
	}

	if cfg.DBUser == "" || cfg.DBName == "" {
		err := fmt.Errorf("config: обязательные поля DB_USER и DB_NAME должны быть заданы")
		if common.Logger != nil {
			common.Logger.Error("конфигурация невалидна", zap.Error(err))
		}
		return nil, err
	}

	Cfg = cfg

	if common.Logger != nil {
		if usedDefaults {
			common.Logger.Info("конфиг загружен с fallback-значениями", zap.String("host", cfg.Host), zap.String("port", cfg.Port))
		} else {
			common.Logger.Info("конфиг успешно загружен из .env", zap.String("host", cfg.Host), zap.String("port", cfg.Port))
		}
	}

	return cfg, nil
}
