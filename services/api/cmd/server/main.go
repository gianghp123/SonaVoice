package main

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"

	"github.com/gianghp123/SonaVoice/api/internal/configs"
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	httpclient "github.com/gianghp123/SonaVoice/api/internal/http-client"
	"github.com/gianghp123/SonaVoice/api/internal/middlewares"
	message "github.com/gianghp123/SonaVoice/api/internal/modules/message"
	session "github.com/gianghp123/SonaVoice/api/internal/modules/session"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/gianghp123/SonaVoice/api/cmd/server/docs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getServerConfig() *configs.ServerConfig {
	return &configs.ServerConfig{
		Mode: utils.GetEnv("MODE", "debug"),
		Port: utils.GetEnv("PORT", "8080"),
	}
}

// @title           Session API
// @version         1.0
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	cfg := configs.Load()

	// Init Sentry
	if cfg.Sentry.Dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.Sentry.Dsn,
			Environment: cfg.Sentry.Environment,
			EnableLogs:  true,
			Debug:       true,
		}); err != nil {
			panic(fmt.Sprintf("sentry initialization failed: %v", err))
		}
		defer sentry.Flush(2 * time.Second)
	}

	// Init logger
	zapLogger.Init(cfg.Logger, cfg.Sentry)
	defer zapLogger.Sync()
	logger := zapLogger.S()

	// Set gin mode
	serverCfg := getServerConfig()
	gin.SetMode(serverCfg.Mode)

	// Connect database
	db, err := gorm.Open(postgres.Open(cfg.Database.DatabaseUrl), &gorm.Config{})
	if err != nil {
		logger.Fatalf("failed to connect database: %v", err)
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatalf("failed to get sql.DB: %v", err)
		return
	}
	if err := sqlDB.Ping(); err != nil {
		logger.Fatalf("failed to ping database: %v", err)
		return
	}
	logger.Info("database connected successfully")

	// Redis client
	redisOpts, err := redis.ParseURL(cfg.Redis.RedisUrl)
	if err != nil {
		logger.Fatalf("failed to parse redis url: %v", err)
		return
	}
	redisClient := redis.NewClient(redisOpts)

	//middlewares
	globalLimiter := middlewares.NewRateLimitMiddleware(
		redisClient,
		"rl:global",
		"60-M",
	)
	sessionLimiter := middlewares.NewRateLimitMiddleware(
		redisClient,
		"rl:session",
		"5-M",
	)
	auth := middlewares.ClerkAuthMiddleware()

	// Init Gin
	router := gin.Default()

	// Sentry middleware - must be first to catch panics
	router.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	// Attach user info to Sentry scope when available
	router.Use(func(c *gin.Context) {
		if hub := sentrygin.GetHubFromContext(c); hub != nil {
			userID := utils.GetCtx[string](c, enums.ContextKeyUserID)
			if userID != "" {
				hub.Scope().SetUser(sentry.User{ID: userID})
			}
			role := utils.GetCtx[enums.UserRole](c, enums.ContextKeyUserRole)
			if role != "" {
				hub.Scope().SetTag("role", string(role))
			}
		}
		c.Next()
	})

	// Enable rate limiting
	router.ForwardedByClientIP = true
	router.Use(globalLimiter)

	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Register modules
	httpClient := httpclient.NewHttpClient()
	session.SetupModule(router.Group("/"), db, httpClient, auth, sessionLimiter)
	message.SetupModule(router.Group("/"), db, auth)

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Server address
	addr := fmt.Sprintf(":%s", serverCfg.Port)

	logger.Infof("session server running on %s", addr)
	logger.Infof("swagger running on %s/swagger/index.html", addr)

	// Start server
	if err := router.Run(addr); err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}
}
