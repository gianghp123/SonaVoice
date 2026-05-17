package modelgateway

import (
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/quota"
	httpclient "github.com/gianghp123/SonaVoice/api/internal/http-client"
	"github.com/gianghp123/SonaVoice/api/internal/middlewares"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/controllers"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/repositories"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	redisClient "github.com/gianghp123/SonaVoice/api/internal/redis-client"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(router *gin.RouterGroup, db *gorm.DB, httpClient httpclient.IHttpClient, redis redisClient.IRedisClient) {
	sessionRepo := repositories.NewSessionRepository(db)
	configRepo := repositories.NewGlobalConfigRepository(db)

	quotaService := quota.NewQuotaService(redis)
	sessionLockService := services.NewSessionLockService(redis)
	sessionService := services.NewSessionService(sessionRepo)
	speechProxyService := services.NewSpeechProxyService(httpClient)
	configService := services.NewGlobalConfigService(configRepo)

	modelGatewayService := services.NewModelGatewayService(configService, sessionService, speechProxyService, quotaService, sessionLockService)
	modelGatewayController := controllers.NewModelGatewayController(modelGatewayService, sessionService)
	globalConfigController := controllers.NewGlobalConfigController(configService)

	mgGroup := router.Group("/model-gateway")
	mgGroup.POST("/sessions", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleCreateSession)
	mgGroup.GET("/sessions", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleListSessions)
	mgGroup.POST("/sessions/:sessionId/resume", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleResumeSession)
	mgGroup.POST("/sessions/:sessionId/api/offer", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleOffer)
	mgGroup.PATCH("/sessions/:sessionId/api/offer", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleOffer)
	mgGroup.POST("/sessions/:sessionId/close", modelGatewayController.HandleCloseSession)

	gcGroup := router.Group("/global-config")
	gcGroup.GET("", globalConfigController.HandleGet)
	gcGroup.PUT("", middlewares.ClerkAuthMiddleware(), middlewares.RequireRole(enums.UserRoleAdmin), globalConfigController.HandleUpdate)
}
