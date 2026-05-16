package modelgateway

import (
	httpclient "github.com/gianghp123/SonaVoice/api/internal/http-client"
	"github.com/gianghp123/SonaVoice/api/internal/middlewares"
	globalConfigRepository "github.com/gianghp123/SonaVoice/api/internal/modules/global-config/repositories"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/controllers"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/repositories"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	redisClient "github.com/gianghp123/SonaVoice/api/internal/redis-client"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(router *gin.RouterGroup, db *gorm.DB, httpClient httpclient.IHttpClient, redis redisClient.IRedisClient) {
	sessionRepo := repositories.NewSessionRepository(db)
	globalConfigRepo := globalConfigRepository.NewGlobalConfigRepository(db)
	quoteService := services.NewQuoteService(redis)
	modelGatewayService := services.NewModelGatewayService(httpClient, sessionRepo, globalConfigRepo, quoteService)
	controller := controllers.NewModelGatewayController(modelGatewayService)

	group := router.Group("/model-gateway")

	group.POST("/sessions", middlewares.ClerkAuthMiddleware(), controller.HandleCreateSession)
	group.POST("/start", middlewares.ClerkAuthMiddleware(), controller.HandleStart)
	group.POST("/sessions/:sessionId/api/offer", middlewares.ClerkAuthMiddleware(), controller.HandleOffer)
	group.PATCH("/sessions/:sessionId/api/offer", middlewares.ClerkAuthMiddleware(), controller.HandleOffer)
}
