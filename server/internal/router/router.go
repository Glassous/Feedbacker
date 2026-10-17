// Package router 组装路由与中间件。
package router

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"feedbacker/internal/config"
	"feedbacker/internal/handler"
	"feedbacker/internal/middleware"
	"feedbacker/internal/service"
)

// New 构建 HTTP 路由。
func New(cfg *config.Config, logger *slog.Logger, db *mongo.Database, authSvc *service.AuthService, feedbackSvc *service.FeedbackService, version string) *gin.Engine {
	if cfg.LogLevel > slog.LevelDebug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.RedirectTrailingSlash = false

	if cfg.TrustProxy {
		_ = engine.SetTrustedProxies([]string{"0.0.0.0/0", "::/0"})
	} else {
		_ = engine.SetTrustedProxies(nil)
	}

	engine.Use(
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.Recovery(logger),
		middleware.SecurityHeaders(),
		middleware.CORS(cfg.CORSOrigins),
	)

	engine.GET("/healthz", healthHandler(db, version))
	engine.GET("/api/meta", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"siteName": cfg.SiteName,
			"version":  version,
		})
	})

	publicHandler := handler.NewPublicHandler(feedbackSvc, logger)
	authHandler := handler.NewAuthHandler(authSvc, logger)
	adminHandler := handler.NewAdminHandler(feedbackSvc, logger)

	// 公开接口：限流 + 可选 API Key，供其他项目接入。
	public := engine.Group("/api/v1")
	public.POST("/feedback",
		middleware.RateLimit(cfg.RateLimitPerMinute, time.Minute),
		middleware.APIKey(cfg.FeedbackAPIKey),
		publicHandler.Submit,
	)

	// 管理接口：登录本身也做限流，避免暴力破解。
	admin := engine.Group("/api/admin")
	admin.POST("/login", middleware.RateLimit(10, time.Minute), authHandler.Login)

	authorized := admin.Group("", middleware.Auth(authSvc))
	authorized.GET("/me", authHandler.Me)
	authorized.POST("/password", authHandler.ChangePassword)
	authorized.POST("/logout", authHandler.Logout)

	guarded := authorized.Group("", middleware.RequirePasswordChanged())
	guarded.GET("/feedback", adminHandler.List)
	guarded.GET("/feedback/:id", adminHandler.Get)
	guarded.PATCH("/feedback/:id", adminHandler.UpdateStatus)
	guarded.DELETE("/feedback/:id", adminHandler.Delete)
	guarded.GET("/projects", adminHandler.Projects)

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "not_found",
			"message": "接口不存在",
		})
	})

	return engine
}

func healthHandler(db *mongo.Database, version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.Client().Ping(ctx, nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "degraded",
				"version": version,
				"error":   "database unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": version,
		})
	}
}
