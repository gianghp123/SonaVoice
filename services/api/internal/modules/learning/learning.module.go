package learning

import (
	openaiclient "github.com/gianghp123/SonaVoice/api/internal/clients/openai-client"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/controllers"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/repositories"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/services"
	messagerepo "github.com/gianghp123/SonaVoice/api/internal/modules/message/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(router *gin.RouterGroup, db *gorm.DB, openaiClient openaiclient.IOpenAIClient, authMiddlware gin.HandlerFunc) {
	messageRepo := messagerepo.NewMessageRepository(db)
	grammarRepo := repositories.NewGrammarAnalysisRepository(db)
	grammarSvc := services.NewGrammarService(openaiClient, messageRepo, grammarRepo)
	grammarCtrl := controllers.NewGrammarController(grammarSvc)

	analyzeGroup := router.Group("/learning/grammar")
	analyzeGroup.POST("/analyze", authMiddlware, grammarCtrl.HandleAnalyzeText)

	messageGroup := router.Group("/learning/grammar/messages/:messageId")
	messageGroup.POST("", authMiddlware, grammarCtrl.HandleAnalyze)

	sessionGroup := router.Group("/learning/grammar/sessions/:sessionId")
	sessionGroup.GET("", authMiddlware, grammarCtrl.HandleGetBySession)
}
