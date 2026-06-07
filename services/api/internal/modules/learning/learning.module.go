package learning

import (
	openaiclient "github.com/gianghp123/SonaVoice/api/internal/clients/openai-client"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(router *gin.RouterGroup, db *gorm.DB, openaiClient openaiclient.IOpenAIClient, authMiddlware gin.HandlerFunc) {

}
