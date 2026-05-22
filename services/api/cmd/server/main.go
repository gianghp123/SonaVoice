package main

import (
	"fmt"

	"github.com/gianghp123/SonaVoice/api/internal/configs"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	httpclient "github.com/gianghp123/SonaVoice/api/internal/http-client"
	"github.com/gianghp123/SonaVoice/api/internal/utils"

	session "github.com/gianghp123/SonaVoice/api/internal/modules/session"
	message "github.com/gianghp123/SonaVoice/api/internal/modules/message"

	"github.com/gin-gonic/gin"
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

	// Init logger
	zapLogger.Init(cfg.Logger)
	defer zapLogger.Sync()

	logger := zapLogger.S()

	serverCfg := getServerConfig()
	// Set gin mode
	gin.SetMode(serverCfg.Mode)

	// Connect database
	db, err := gorm.Open(postgres.Open(cfg.Database.DatabaseUrl), &gorm.Config{})
	if err != nil {
		logger.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatalf("failed to get sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		logger.Fatalf("failed to ping database: %v", err)
	}

	logger.Info("database connected successfully")

	httpClient := httpclient.NewHttpClient()

	// Init Gin
	router := gin.Default()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Register modules
	session.SetupModule(router.Group("/"), db, httpClient)
	message.SetupModule(router.Group("/"), db)

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
