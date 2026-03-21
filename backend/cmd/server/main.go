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
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	common.InitLogger()
	defer common.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		if common.Logger != nil {
			common.Logger.Fatal("не удалось загрузить конфиг", zap.Error(err))
		}
		os.Exit(1)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

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
