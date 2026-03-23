package app

import (
	"github.com/devilzzcpp/agregator-zzxx/internal/subscription"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), Middleware(), CORSMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	v1 := r.Group("/api/v1")

	subscriptionRepository := subscription.NewRepository(db)
	subscriptionService := subscription.NewService(subscriptionRepository)
	subscriptionHandler := subscription.NewHandler(subscriptionService)

	subscriptions := v1.Group("/subscriptions")
	{
		subscriptions.POST("", subscriptionHandler.Create)
		subscriptions.GET("", subscriptionHandler.List)
		subscriptions.GET("/:id", subscriptionHandler.GetByID)
		subscriptions.PUT("/:id", subscriptionHandler.Update)
		subscriptions.DELETE("/:id", subscriptionHandler.Delete)
		subscriptions.GET("/total", subscriptionHandler.Total)
	}

	return r
}
