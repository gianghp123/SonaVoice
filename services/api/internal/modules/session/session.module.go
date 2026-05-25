package session

import (
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	httpclient "github.com/gianghp123/SonaVoice/api/internal/http-client"
	"github.com/gianghp123/SonaVoice/api/internal/middlewares"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/controllers"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/repositories"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(
	router *gin.RouterGroup,
	db *gorm.DB, httpClient httpclient.IHttpClient,
	authMiddleware gin.HandlerFunc,
	sessionLimiter gin.HandlerFunc,
) {
	sessionRepo := repositories.NewSessionRepository(db)
	configRepo := repositories.NewSessionConfigRepository(db)
	userQuotaRepo := repositories.NewUserQuotaRepository(db)

	uow := transaction.NewUnitOfWork(db)

	sessionService := services.NewSessionService(sessionRepo)
	speechProxyService := services.NewSpeechProxyService(httpClient)
	configService := services.NewSessionConfigService(configRepo)
	startConnectionSvc := services.NewStartConnectionService(speechProxyService, uow)
	quotaService := services.NewQuotaService(userQuotaRepo)

	orchestratorService := services.NewOrchestratorService(configService, sessionService, speechProxyService, startConnectionSvc, quotaService, uow)
	sessionController := controllers.NewSessionController(orchestratorService)
	sessionConfigController := controllers.NewSessionConfigController(configService)

	sessGroup := router.Group("/sessions")
	sessGroup.POST("", sessionLimiter, authMiddleware, sessionController.HandleCreateSession)
	sessGroup.POST("/:sessionId/start", authMiddleware, sessionController.HandleStartConnection)
	sessGroup.POST("/:sessionId/api/offer", authMiddleware, sessionController.HandleOffer)
	sessGroup.PATCH("/:sessionId/api/offer", authMiddleware, sessionController.HandleOffer)
	sessGroup.GET("", authMiddleware, sessionController.HandleListSessions)
	sessGroup.GET("/:sessionId", authMiddleware, sessionController.HandleGetSession)
	sessGroup.POST("/:sessionId/cancel", authMiddleware, sessionController.HandleCancelSession)
	sessGroup.POST("/:sessionId/close", sessionController.HandleCloseSession)

	scGroup := sessGroup.Group("/config")
	scGroup.GET("", sessionConfigController.HandleGet)
	scGroup.PUT("", authMiddleware, middlewares.RequireRole(enums.UserRoleAdmin), sessionConfigController.HandleUpdate)
}
