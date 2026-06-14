package main

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"

	httpclient "github.com/gianghp123/SonaVoice/api/internal/clients/http-client"
	openaiclient "github.com/gianghp123/SonaVoice/api/internal/clients/openai-client"
	"github.com/gianghp123/SonaVoice/api/internal/configs"
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/middlewares"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning"
	message "github.com/gianghp123/SonaVoice/api/internal/modules/message"
	session "github.com/gianghp123/SonaVoice/api/internal/modules/session"
	user_profile "github.com/gianghp123/SonaVoice/api/internal/modules/user_profile"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/v3" // imported as openai
	"github.com/openai/openai-go/v3/option"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/gianghp123/SonaVoice/api/cmd/server/docs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
	gin.SetMode(cfg.Server.Mode)

	// Connect database
	db, err := gorm.Open(
		postgres.New(
			postgres.Config{
				DSN:                  cfg.Database.DatabaseUrl,
				PreferSimpleProtocol: true,
			}),
		&gorm.Config{},
	)
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

	// OpenAI SDK client
	sdkClient := openai.NewClient(
		option.WithAPIKey(cfg.OpenAI.APIKey),
		option.WithBaseURL(cfg.OpenAI.BaseURL),
	)

	openaiClient := openaiclient.NewOpenAIClient(
		&sdkClient,
		cfg.OpenAI.Model,
	)

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
		fmt.Sprintf("%s:rl:global", cfg.Redis.RedisKeyPrefix),
		"60-M",
	)
	sessionLimiter := middlewares.NewRateLimitMiddleware(
		redisClient,
		fmt.Sprintf("%s:rl:session", cfg.Redis.RedisKeyPrefix),
		"5-M",
	)
	auth := middlewares.ClerkAuthMiddleware()
	internalSecret := middlewares.InternalSecretMiddleware()

	// Init Gin
	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.Server.AllowUrls
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		// "X-Requested-With",
		// "X-CSRF-Token",
		// "svix-id",
		// "svix-timestamp",
		// "svix-signature",
	}
	corsConfig.AllowWebSockets = true

	router.Use(cors.New(corsConfig))

	// Sentry middleware - must be first to catch panics
	if cfg.Sentry.Dsn != "" {
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
	}

	// Enable rate limiting
	router.ForwardedByClientIP = true
	router.Use(globalLimiter)

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "/health ok"})
	})

	// Register modules
	httpClient := httpclient.NewHttpClient()
	session.SetupModule(router.Group("/"), db, httpClient, auth, sessionLimiter, internalSecret)
	message.SetupModule(router.Group("/"), db, auth, internalSecret)
	user_profile.SetupModule(router.Group("/"), db, auth)
	learning.SetupModule(router.Group("/"), db, openaiClient, auth)

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Server address
	addr := fmt.Sprintf(":%s", cfg.Server.Port)

	logger.Infof("session server running on %s", addr)
	logger.Infof("swagger running on %s/swagger/index.html", addr)

	// Start server
	if err := router.Run(addr); err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}
}
