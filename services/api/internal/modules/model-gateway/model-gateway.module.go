package modelgateway

import (
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/controllers"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(router *gin.RouterGroup, db *gorm.DB, middleware ...gin.HandlerFunc) {
	serivce := services.NewModelGatewayService()
	controller := controllers.NewModelGatewayController(serivce)

	group := router.Group("/model-gateway")
	if len(middleware) > 0 {
		group.Use(middleware...)
	}

	group.POST("/start", controller.HandleStart)
	group.POST("/sessions/:sessionId/api/offer", controller.HandleOffer)
	group.PATCH("/sessions/:sessionId/api/offer", controller.HandleOffer)
}
