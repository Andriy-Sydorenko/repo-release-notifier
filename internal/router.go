package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, cfg *Config) *gin.Engine {
	router := gin.Default()

	repo := NewRepository(db)
	github := NewGitHubClient(cfg.GitHubToken)
	notifier := NewNotifier(cfg)
	service := NewService(repo, github, notifier)
	handler := NewHandler(service)

	RegisterRoutes(router, handler, cfg.APIKey)

	return router
}

func RegisterRoutes(router *gin.Engine, h *Handler, apiKey string) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	api.Use(APIKeyAuth(apiKey))
	{
		api.POST("/subscribe", h.Subscribe)
		api.GET("/confirm/:token", h.ConfirmSubscription)
	}
}
