package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devilzzcpp/agregator-zzxx/common"
	"github.com/devilzzcpp/agregator-zzxx/config"
	"github.com/devilzzcpp/agregator-zzxx/internal/app"
	"go.uber.org/zap"
)

func main() {
	cfg, meta, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "не удалось загрузить конфиг: %v\n", err)
		os.Exit(1)
	}

	common.InitLogger(common.LoggerConfig{
		Level:      cfg.Logger.Level,
		File:       cfg.Logger.File,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		TimeZone:   cfg.TimeZone,
	})
	defer common.Sync()

	if common.Logger != nil {
		fields := []zap.Field{
			zap.String("host", cfg.Host),
			zap.String("port", cfg.Port),
		}

		if meta != nil {
			fields = append(fields,
				zap.String("config_source", meta.ConfigSource),
				zap.String("config_priority", meta.Priority),
				zap.Bool("env_file_loaded", meta.EnvFileLoaded),
			)

			if meta.FallbackReason != "" {
				fields = append(fields, zap.String("fallback_reason", meta.FallbackReason))
			}
		}

		if meta != nil && !meta.EnvFileLoaded {
			common.Logger.Warn(".env не загружен, приложение работает на переменных окружения и значениях по умолчанию", fields...)
		} else {
			common.Logger.Info("конфиг успешно загружен", fields...)
		}
	}

	db, err := common.InitDB(cfg)
	if err != nil {
		if common.Logger != nil {
			common.Logger.Fatal("не удалось подключиться к базе данных", zap.Error(err))
		}
		os.Exit(1)
	}

	defer func() {
		if err := common.CloseDB(); err != nil && common.Logger != nil {
			common.Logger.Error("ошибка при закрытии соединения с базой данных", zap.Error(err))
		}
	}()

	r := app.NewRouter(db)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if common.Logger != nil {
		common.Logger.Info("сервер запускается", zap.String("addr", addr))
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if common.Logger != nil {
				common.Logger.Fatal("ошибка запуска HTTP сервера", zap.Error(err))
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if common.Logger != nil {
		common.Logger.Info("получен сигнал остановки, завершаем сервер")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		if common.Logger != nil {
			common.Logger.Error("ошибка graceful shutdown", zap.Error(err))
		}
	}

	if common.Logger != nil {
		common.Logger.Info("сервер остановлен")
	}
}
