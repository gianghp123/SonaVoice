package controllers

import (
	"io"
	"net/http"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	_ "github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/gin-gonic/gin"
)

type ModelGatewayController struct {
	svc services.IModelGatewayService
}

func NewModelGatewayController(svc services.IModelGatewayService) *ModelGatewayController {
	return &ModelGatewayController{svc: svc}
}

// HandleCreateSession godoc
// @Summary      Create a new session
// @Description  Create a new session for an authenticated user. Use the returned session ID with /start.
// @Security     Bearer
// @Tags         model-gateway
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.BaseResponse[res.SessionRes]
// @Failure      401  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /model-gateway/sessions [post]
func (ctrl *ModelGatewayController) HandleCreateSession(c *gin.Context) {
	userID := utils.GetCtx[string](c.Request.Context(), enums.ContextKeyUserID)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.Fail(errors.Unauthorized("authentication required")))
		return
	}

	sessionRes, appErr := ctrl.svc.CreateSession(c.Request.Context())
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}
	c.JSON(http.StatusOK, response.Success(sessionRes))
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
	var req req.StartConnectionReq

	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	offer, appErr := ctrl.svc.StartConnection(c.Request.Context(), &req)

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
