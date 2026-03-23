package app

import (
	"time"

	"github.com/devilzzcpp/agregator-zzxx/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RequestIDKey = "request_id"

// Middleware генерирует RequestID для каждого запроса, логирует метод, путь, статус и длительность
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// генерируем или берём RequestID из заголовка для трейсинга
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		// логируем начало
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// логируем завершение с RequestID, методом, путём, статусом и длительностью
		if common.Logger != nil {
			common.Logger.Info(
				c.Request.Method+" "+c.Request.URL.Path,
				zap.String("request_id", requestID),
				zap.Int("status", c.Writer.Status()),
				zap.String("duration", duration.String()),
			)
		}
	}
}

// КОРС что бы был, на будущее может
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
