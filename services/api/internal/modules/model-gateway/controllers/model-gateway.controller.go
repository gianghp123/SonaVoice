package controllers

import (
	"io"
	"net/http"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	_ "github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/gin-gonic/gin"
)

type ModelGatewayController struct {
	svc services.IModelGatewayService
}

func NewModelGatewayController(svc services.IModelGatewayService) *ModelGatewayController {
	return &ModelGatewayController{svc: svc}
}

// HandleStart godoc
// @Summary      Create WebRTC connection
// @Description  Start a new WebRTC session with the speech service and return connection information
// @Security     Bearer
// @Tags         model-gateway
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.BaseResponse[res.WebRTCConnectionRes]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /model-gateway/start [post]
func (ctrl *ModelGatewayController) HandleStart(c *gin.Context) {
	offer, appErr := ctrl.svc.GetConnnection(c.Request.Context())
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}
	c.JSON(http.StatusOK, response.Success(offer))
}

// HandleOffer godoc
// @Summary      Proxy WebRTC offer
// @Description  Proxy a WebRTC offer request to the speech service by session ID
// @Security     Bearer
// @Tags         model-gateway
// @Accept       json
// @Produce      json
// @Param        sessionId  path      string  true  "Session ID"
// @Param        body       body      any     true  "WebRTC offer request body"
// @Success      200        {object}  response.BaseResponse[any]
// @Failure      400        {object}  response.BaseResponse[any]
// @Failure      500        {object}  response.BaseResponse[any]
// @Router       /model-gateway/sessions/{sessionId}/offer [post]
// @Router       /model-gateway/sessions/{sessionId}/offer [patch]
func (ctrl *ModelGatewayController) HandleOffer(c *gin.Context) {
	sessionId := c.Param("sessionId")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(errors.BadRequest("failed to read request body")))
		return
	}

	respBody, statusCode, appErr := ctrl.svc.ProxyOffer(c.Request.Context(), sessionId, c.Request.Method, body)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	// ✅ Return raw response directly, no wrapping
	c.Data(statusCode, "application/json", respBody)
}
