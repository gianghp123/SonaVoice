package user_profile

import (
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/controllers"
	"github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/repositories"
	"github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupModule(router *gin.RouterGroup, db *gorm.DB, authMiddleware gin.HandlerFunc) {
	profileRepo := repositories.NewUserProfileRepository(db)
	uow := transaction.NewUnitOfWork(db)
	profileService := services.NewUserProfileService(profileRepo, uow)
	profileController := controllers.NewUserProfileController(profileService)

	profileGroup := router.Group("/profile")
	profileGroup.GET("", authMiddleware, profileController.HandleGetProfile)
	profileGroup.POST("", authMiddleware, profileController.HandleOnboardProfile)
	profileGroup.PATCH("", authMiddleware, profileController.HandleUpdateProfile)
}
