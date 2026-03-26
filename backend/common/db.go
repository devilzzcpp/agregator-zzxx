package common

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/devilzzcpp/agregator-zzxx/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB открывает соединение с PostgreSQL через GORM,
// проверяет его через Ping и настраивает базовый пул соединений.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("db: config is nil")
	}

	dsn := cfg.DSN() // строка DSN для подключения к базе данных

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: buildGormLogger(cfg.Logger.Level),
	})
	if err != nil {
		return nil, fmt.Errorf("db: failed to open connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("db: failed to get sql.DB: %w", err)
	}

	configureConnectionPool(sqlDB) // настраиваем пул соединений для оптимальной работы с PostgreSQL

	// проверяем соединение с базой данных, выполняя Ping

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("db: ping failed: %w", err)
	}

	if Logger != nil {
		Logger.Info("подключение к базе данных успешно установлено",
			zap.String("db_host", cfg.DBHost),
			zap.String("db_port", cfg.DBPort),
			zap.String("db_name", cfg.DBName),
		)
	}

	DB = db

	return db, nil
}

// configureConnectionPool настраивает параметры пула соединений для sql.DB
func configureConnectionPool(sqlDB *sql.DB) {
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(15 * time.Minute)
}

// buildGormLogger создает адаптер для логгера GORM на основе уровня логирования приложения
func buildGormLogger(appLogLevel string) gormlogger.Interface {
	level := gormlogger.Warn

	switch appLogLevel {
	case "debug":
		level = gormlogger.Info
	case "error":
		level = gormlogger.Error
	case "silent":
		level = gormlogger.Silent
	}

	return gormlogger.Default.LogMode(level)
}

// CloseDB закрывает соединение с базой данных, если оно открыто, и логирует результат
func CloseDB() error {
	if DB == nil {
		return nil
	}

	// получаем sql.DB из GORM для закрытия соединения
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("db: failed to get sql.DB for close: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("db: close failed: %w", err)
	}

	if Logger != nil {
		Logger.Info("соединение с базой данных закрыто")
	}

	DB = nil
	return nil
}
