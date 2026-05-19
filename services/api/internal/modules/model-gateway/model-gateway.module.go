package modelgateway

import (
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	httpclient "github.com/gianghp123/SonaVoice/api/internal/http-client"
	"github.com/gianghp123/SonaVoice/api/internal/middlewares"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/controllers"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/repositories"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(router *gin.RouterGroup, db *gorm.DB, httpClient httpclient.IHttpClient) {
	sessionRepo := repositories.NewSessionRepository(db)
	configRepo := repositories.NewGlobalConfigRepository(db)
	userQuotaRepo := repositories.NewUserQuotaRepository(db)

	uow := transaction.NewUnitOfWork(db)

	sessionService := services.NewSessionService(sessionRepo)
	speechProxyService := services.NewSpeechProxyService(httpClient)
	configService := services.NewGlobalConfigService(configRepo)
	startConnectionSvc := services.NewStartConnectionService(speechProxyService, uow)
	quotaService := services.NewQuotaService(userQuotaRepo)

	modelGatewayService := services.NewModelGatewayService(configService, sessionService, speechProxyService, startConnectionSvc, quotaService, uow)
	modelGatewayController := controllers.NewModelGatewayController(modelGatewayService)
	globalConfigController := controllers.NewGlobalConfigController(configService)

	mgGroup := router.Group("/model-gateway")
	mgGroup.POST("/sessions", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleCreateSession)
	mgGroup.POST("/sessions/:sessionId/start", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleStartConnection)
	mgGroup.POST("/sessions/:sessionId/api/offer", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleOffer)
	mgGroup.PATCH("/sessions/:sessionId/api/offer", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleOffer)
	mgGroup.POST("/sessions/:sessionId/cancel", middlewares.ClerkAuthMiddleware(), modelGatewayController.HandleCancelSession)
	mgGroup.POST("/sessions/:sessionId/close", modelGatewayController.HandleCloseSession)

	gcGroup := router.Group("/global-config")
	gcGroup.GET("", globalConfigController.HandleGet)
	gcGroup.PUT("", middlewares.ClerkAuthMiddleware(), middlewares.RequireRole(enums.UserRoleAdmin), globalConfigController.HandleUpdate)
}
