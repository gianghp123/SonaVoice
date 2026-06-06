package controllers

import (
	"net/http"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/services"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserProfileController struct {
	svc services.IUserProfileService
}

func NewUserProfileController(svc services.IUserProfileService) *UserProfileController {
	return &UserProfileController{svc: svc}
}

func (ctrl *UserProfileController) HandleGetProfile(c *gin.Context) {
	userID := utils.GetCtx[string](c.Request.Context(), enums.ContextKeyUserID)

	profile, appErr := ctrl.svc.GetByUserID(c.Request.Context(), userID)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	result := res.UserProfileRes{
		ID:           profile.ID,
		UserID:       profile.UserID,
		DisplayName:  profile.DisplayName,
		EnglishLevel: profile.EnglishLevel,
		Preferences:  profile.Preferences,
		CreatedAt:    profile.CreatedAt,
		UpdatedAt:    profile.UpdatedAt,
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func (ctrl *UserProfileController) HandleOnboardProfile(c *gin.Context) {
	userID := utils.GetCtx[string](c.Request.Context(), enums.ContextKeyUserID)

	var body req.UpsertProfileReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(errors.BadRequest("invalid request body")))
		return
	}

	appErr := ctrl.svc.Onboard(c.Request.Context(), userID, &body)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	c.JSON(http.StatusOK, response.Success(true))
}

func (ctrl *UserProfileController) HandleUpdateProfile(c *gin.Context) {
	userID := utils.GetCtx[string](c.Request.Context(), enums.ContextKeyUserID)

	var body req.UpdateProfileReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(errors.BadRequest("invalid request body")))
		return
	}

	appErr := ctrl.svc.Update(c.Request.Context(), userID, &body)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	c.JSON(http.StatusOK, response.Success(true))
}
