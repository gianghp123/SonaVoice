package controllers

import (
	"net/http"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/domain"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/gin-gonic/gin"
)

type GlobalConfigController struct {
	svc services.IGlobalConfigService
}

func NewGlobalConfigController(svc services.IGlobalConfigService) *GlobalConfigController {
	return &GlobalConfigController{svc: svc}
}

// HandleGet godoc
// @Summary      Get global config
// @Description  Retrieve the global configuration
// @Tags         global-config
// @Produce      json
// @Success      200  {object}  response.BaseResponse[res.GlobalConfigRes]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /global-config [get]
func (ctrl *GlobalConfigController) HandleGet(c *gin.Context) {
	model, appErr := ctrl.svc.Get(c.Request.Context())
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}
	configPayload, err := domain.ParseGlobalConfig(model.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal()))
		return
	}
	c.JSON(http.StatusOK, response.Success(&res.GlobalConfigRes{Config: *configPayload}))
}

// HandleUpdate godoc
// @Summary      Update global config
// @Description  Update the global configuration (admin only)
// @Security     Bearer
// @Tags         global-config
// @Accept       json
// @Produce      json
// @Param        body  body      req.GlobalConfigReq  true  "Global config payload"
// @Success      200   {object}  response.BaseResponse[res.GlobalConfigRes]
// @Failure      400   {object}  response.BaseResponse[any]
// @Failure      401   {object}  response.BaseResponse[any]
// @Failure      403   {object}  response.BaseResponse[any]
// @Failure      500   {object}  response.BaseResponse[any]
// @Router       /global-config [put]
func (ctrl *GlobalConfigController) HandleUpdate(c *gin.Context) {
	var cfg req.GlobalConfigReq

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
	configPayload, err := domain.ParseGlobalConfig(model.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal()))
		return
	}
	c.JSON(http.StatusOK, response.Success(&res.GlobalConfigRes{Config: *configPayload}))
}
