// @title           Subscription Aggregator API
// @version         1.0
// @description     REST-сервис агрегации онлайн-подписок пользователей

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/devilzzcpp/agregator-zzxx/common"
	"github.com/devilzzcpp/agregator-zzxx/config"
	_ "github.com/devilzzcpp/agregator-zzxx/docs"
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

	// логируем источник конфигурации
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
			// если была причина fallback, логируем её
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

	// гарантируем закрытие соединения с базой данных при завершении работы приложения
	defer func() {
		if err := common.CloseDB(); err != nil && common.Logger != nil {
			common.Logger.Error("ошибка при закрытии соединения с базой данных", zap.Error(err))
		}
	}()

	r := app.NewRouter(db)
	
	// оборачиваем роутер в middleware для защиты Swagger UI базовой аутентификацией
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		
		// проверяем, если запрос к Swagger UI, то требуем базовую аутентификацию
		if strings.HasPrefix(req.URL.Path, "/swagger") {
			auth := req.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Basic ") {
				w.Header().Set("WWW-Authenticate", `Basic realm="Swagger UI"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// декодируем базовую аутентификацию и проверяем логин/пароль
			payload, err := base64.StdEncoding.DecodeString(auth[6:])
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// проверяем логин и пароль
			parts := strings.SplitN(string(payload), ":", 2)
			if len(parts) != 2 || parts[0] != cfg.SwaggerLogin || parts[1] != cfg.SwaggerPassword {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		
		r.ServeHTTP(w, req)
	})

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	srv := &http.Server{
		Addr:              addr,
		Handler:           protectedHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if common.Logger != nil {
		common.Logger.Info("сервер запускается", zap.String("addr", addr))
	}

	// запускаем сервер в отдельной горутине, чтобы не блокировать основной поток для graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if common.Logger != nil {
				common.Logger.Fatal("ошибка запуска HTTP сервера", zap.Error(err))
			}
		}
	}()
	
	// ждем сигнала остановки (например, Ctrl+C) и выполняем graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // ловим сигналы остановки
	<-quit

	if common.Logger != nil {
		common.Logger.Info("получен сигнал остановки, завершаем сервер")
	}

	// создаем контекст с таймаутом для graceful shutdown, чтобы не ждать бесконечно
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
