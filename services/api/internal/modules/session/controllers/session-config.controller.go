package controllers

import (
	"net/http"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/domain"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/services"
	"github.com/gin-gonic/gin"
)

type SessionConfigController struct {
	svc services.ISessionConfigService
}

func NewSessionConfigController(svc services.ISessionConfigService) *SessionConfigController {
	return &SessionConfigController{svc: svc}
}

// HandleGet godoc
// @Summary      Get session config
// @Description  Retrieve the session configuration
// @Tags         session-config
// @Produce      json
// @Success      200  {object}  response.BaseResponse[res.SessionConfigRes]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /sessions/config [get]
func (ctrl *SessionConfigController) HandleGet(c *gin.Context) {
	model, appErr := ctrl.svc.Get(c.Request.Context())
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}
	configPayload, err := domain.ParseSessionConfig(model.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal()))
		return
	}
	c.JSON(http.StatusOK, response.Success(&res.SessionConfigRes{Config: *configPayload}))
}

// HandleUpdate godoc
// @Summary      Update session config
// @Description  Update the session configuration (admin only)
// @Security     Bearer
// @Tags         session-config
// @Accept       json
// @Produce      json
// @Param        body  body      req.SessionConfigReq  true  "Session config payload"
// @Success      200   {object}  response.BaseResponse[res.SessionConfigRes]
// @Failure      400   {object}  response.BaseResponse[any]
// @Failure      401   {object}  response.BaseResponse[any]
// @Failure      403   {object}  response.BaseResponse[any]
// @Failure      500   {object}  response.BaseResponse[any]
// @Router       /sessions/config [put]
func (ctrl *SessionConfigController) HandleUpdate(c *gin.Context) {
	var cfg req.SessionConfigReq

	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	model, appErr := ctrl.svc.Update(c.Request.Context(), &cfg)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}
	configPayload, err := domain.ParseSessionConfig(model.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal()))
		return
	}
	c.JSON(http.StatusOK, response.Success(&res.SessionConfigRes{Config: *configPayload}))
}
