package message

import (
	"github.com/gianghp123/SonaVoice/api/internal/middlewares"
	"github.com/gianghp123/SonaVoice/api/internal/modules/message/controllers"
	"github.com/gianghp123/SonaVoice/api/internal/modules/message/repositories"
	"github.com/gianghp123/SonaVoice/api/internal/modules/message/services"
	sessionrepos "github.com/gianghp123/SonaVoice/api/internal/modules/session/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(router *gin.RouterGroup, db *gorm.DB) {
	messageRepo := repositories.NewMessageRepository(db)
	sessionRepo := sessionrepos.NewSessionRepository(db)
	messageService := services.NewMessageService(messageRepo, sessionRepo)
	messageController := controllers.NewMessageController(messageService)

	msgGroup := router.Group("/sessions/:sessionId/messages")
	msgGroup.GET("", middlewares.ClerkAuthMiddleware(), messageController.HandleListMessages)
	msgGroup.POST("", messageController.HandleCreateMessages)
}
