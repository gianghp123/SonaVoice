package globalconfig

import (
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/middlewares"
	"github.com/gianghp123/SonaVoice/api/internal/modules/global-config/controllers"
	"github.com/gianghp123/SonaVoice/api/internal/modules/global-config/repositories"
	"github.com/gianghp123/SonaVoice/api/internal/modules/global-config/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(router *gin.RouterGroup, db *gorm.DB) {
	repo := repositories.NewGlobalConfigRepository(db)
	service := services.NewGlobalConfigService(repo)
	controller := controllers.NewGlobalConfigController(service)

	group := router.Group("/global-config")

	group.GET("", controller.HandleGet)
	group.PUT("", middlewares.ClerkAuthMiddleware(), middlewares.RequireRole(enums.UserRoleAdmin), controller.HandleUpdate)
}
