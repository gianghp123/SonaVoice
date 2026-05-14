package modelgateway

import (
	httpclient "github.com/gianghp123/SonaVoice/api/internal/http-client"
	"github.com/gianghp123/SonaVoice/api/internal/middlewares"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/controllers"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/repositories"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(router *gin.RouterGroup, db *gorm.DB, httpClient httpclient.IHttpClient) {
	sessionRepo := repositories.NewSessionRepository(db)
	service := services.NewModelGatewayService(httpClient, sessionRepo)
	controller := controllers.NewModelGatewayController(service)

	group := router.Group("/model-gateway")

	group.Use(middlewares.OptionalClerkAuthMiddleware())

	group.POST("/start", controller.HandleStart)
	group.POST("/sessions/:sessionId/api/offer", controller.HandleOffer)
	group.PATCH("/sessions/:sessionId/api/offer", controller.HandleOffer)
}
