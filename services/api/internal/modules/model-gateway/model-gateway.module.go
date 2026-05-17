package modelgateway

import (
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
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

	quoteService := services.NewQuoteService(redis)
	sessionService := services.NewSessionService(sessionRepo)
	speechProxyService := services.NewSpeechProxyService(httpClient)
	configService := services.NewGlobalConfigService(configRepo)

	modelGatewayService := services.NewModelGatewayService(configService, sessionService, speechProxyService, quoteService)
	modelGatewayController := controllers.NewModelGatewayController(modelGatewayService)
	globalConfigController := controllers.NewGlobalConfigController(configService)

	mgGroup := router.Group("/model-gateway")
	mgGroup.POST("/start", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleStart)
	mgGroup.POST("/sessions/:sessionId/api/offer", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleOffer)
	mgGroup.PATCH("/sessions/:sessionId/api/offer", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleOffer)

	gcGroup := router.Group("/global-config")
	gcGroup.GET("", globalConfigController.HandleGet)
	gcGroup.PUT("", middlewares.ClerkAuthMiddleware(), middlewares.RequireRole(enums.UserRoleAdmin), globalConfigController.HandleUpdate)
}
