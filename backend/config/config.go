package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type LoggerConfig struct {
	Level      string
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

type LoadMeta struct {
	EnvFileLoaded  bool
	ConfigSource   string
	FallbackReason string
	Priority       string
}

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
	Logger   LoggerConfig

	// swagger basic auth
	SwaggerLogin    string
	SwaggerPassword string
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

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FILE", "./storage/logs/app.log")
	viper.SetDefault("LOG_SIZE", 50)
	viper.SetDefault("LOG_BACKUP", 5)
	viper.SetDefault("LOG_AGE", 30)

	viper.SetDefault("APP_TIMEZONE", "Europe/Samara")
}

func LoadConfig() (*Config, *LoadMeta, error) {
	setDefaults()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// загружаем конфиг и собираем метаинформацию о процессе загрузки
	meta := &LoadMeta{
		EnvFileLoaded: false,
		ConfigSource:  "переменные окружения и значения по умолчанию", 
		Priority:      "переменные окружения переопределяют значения по умолчанию",
	}

	if err := viper.ReadInConfig(); err != nil {
		meta.FallbackReason = err.Error() // сохраняем причину, по которой не удалось загрузить .env
	} else {
		meta.EnvFileLoaded = true
		meta.ConfigSource = "файл .env, переменные окружения и значения по умолчанию"
		meta.Priority = "переменные окружения переопределяют значения .env, значения .env переопределяют значения по умолчанию"

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

		SwaggerLogin:    viper.GetString("SVG_LOGIN"),
		SwaggerPassword: viper.GetString("SVG_PASSWORD"),

		Logger: LoggerConfig{
			Level:      viper.GetString("LOG_LEVEL"),
			File:       viper.GetString("LOG_FILE"),
			MaxSize:    viper.GetInt("LOG_SIZE"),
			MaxBackups: viper.GetInt("LOG_BACKUP"),
			MaxAge:     viper.GetInt("LOG_AGE"),
		},
	}

	if cfg.DBUser == "" || cfg.DBName == "" {
		return nil, meta, fmt.Errorf("config: обязательные поля DB_USER и DB_NAME должны быть заданы")
	}

	Cfg = cfg
	return cfg, meta, nil
}
