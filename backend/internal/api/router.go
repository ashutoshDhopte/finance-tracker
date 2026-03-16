package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ash/finance-tracker/backend/internal/api/handlers"
	"github.com/ash/finance-tracker/backend/internal/api/middleware"
	"github.com/ash/finance-tracker/backend/internal/config"
	gmailsvc "github.com/ash/finance-tracker/backend/internal/services/gmail"
	"github.com/ash/finance-tracker/backend/internal/services/parser"
)

func NewRouter(cfg *config.Config, pool *pgxpool.Pool, parserSvc *parser.Service, gmailService *gmailsvc.Service) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authHandler := handlers.NewAuthHandler(pool, cfg)
	txnHandler := handlers.NewTransactionHandler(pool)
	catHandler := handlers.NewCategoryHandler(pool)
	accHandler := handlers.NewAccountHandler(pool)
	reportHandler := handlers.NewReportHandler(pool)
	ingestHandler := handlers.NewIngestHandler(pool, parserSvc)
	alertHandler := handlers.NewAlertHandler(pool)
	syncHandler := handlers.NewSyncHandler(gmailService)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", authHandler.Login)

		auth := v1.Group("")
		auth.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			auth.GET("/transactions", txnHandler.List)
			auth.GET("/transactions/:id", txnHandler.Get)
			auth.POST("/transactions", txnHandler.Create)
			auth.PUT("/transactions/:id", txnHandler.Update)
			auth.DELETE("/transactions/:id", txnHandler.Delete)
			auth.POST("/transactions/rehash", txnHandler.Rehash)

			auth.GET("/categories", catHandler.List)
			auth.POST("/categories", catHandler.Create)
			auth.PUT("/categories/:id", catHandler.Update)
			auth.DELETE("/categories/:id", catHandler.Delete)

			auth.GET("/accounts", accHandler.List)
			auth.POST("/accounts", accHandler.Create)
			auth.PUT("/accounts/:id", accHandler.Update)
			auth.DELETE("/accounts/:id", accHandler.Delete)

			auth.GET("/reports/monthly", reportHandler.Monthly)
			auth.GET("/reports/biweekly", reportHandler.Biweekly)
			auth.GET("/reports/categories", reportHandler.Categories)
			auth.GET("/reports/trends", reportHandler.Trends)

			auth.POST("/transactions/import/csv", ingestHandler.ImportCSV)

			auth.GET("/alerts", alertHandler.List)
			auth.POST("/alerts", alertHandler.Create)
			auth.PUT("/alerts/:id", alertHandler.Update)
			auth.DELETE("/alerts/:id", alertHandler.Delete)
			auth.GET("/alerts/check", alertHandler.Check)

			auth.POST("/sync/gmail", syncHandler.SyncGmail)
		}
	}

	return r
}
