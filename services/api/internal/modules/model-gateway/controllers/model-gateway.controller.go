package controllers

import (
	"io"
	"net/http"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/gin-gonic/gin"
)

type ModelGatewayController struct {
	svc            services.IModelGatewayService
	sessionService services.ISessionService
}

func NewModelGatewayController(svc services.IModelGatewayService, sessionService services.ISessionService) *ModelGatewayController {
	return &ModelGatewayController{svc: svc, sessionService: sessionService}
}

// HandleCreateSession godoc
// @Summary      Create new session
// @Description  Create a new session and start a WebRTC connection with the speech service
// @Security     Bearer
// @Tags         model-gateway
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.BaseResponse[res.CreateSessionRes]
// @Failure      401  {object}  response.BaseResponse[any]
// @Failure      403  {object}  response.BaseResponse[any]
// @Failure      409  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /model-gateway/sessions [post]
func (ctrl *ModelGatewayController) HandleCreateSession(c *gin.Context) {
	offer, appErr := ctrl.svc.CreateSession(c.Request.Context())
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}
	c.JSON(http.StatusOK, response.Success(offer))
}

// HandleResumeSession godoc
// @Summary      Resume an existing session
// @Description  Resume an inactive session with a new speech engine connection
// @Security     Bearer
// @Tags         model-gateway
// @Accept       json
// @Produce      json
// @Param        sessionId  path      string  true  "Session ID"
// @Success      200  {object}  response.BaseResponse[res.CreateSessionRes]
// @Failure      400  {object}  response.BaseResponse[any]
// @Failure      403  {object}  response.BaseResponse[any]
// @Failure      409  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /model-gateway/sessions/{sessionId}/resume [post]
func (ctrl *ModelGatewayController) HandleResumeSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	offer, appErr := ctrl.svc.ResumeSession(c.Request.Context(), sessionID)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}
	c.JSON(http.StatusOK, response.Success(offer))
}

// HandleListSessions godoc
// @Summary      List resumable sessions
// @Description  List all resumable (inactive) sessions for the authenticated user
// @Security     Bearer
// @Tags         model-gateway
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.BaseResponse[[]res.SessionListItemRes]
// @Failure      401  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /model-gateway/sessions [get]
func (ctrl *ModelGatewayController) HandleListSessions(c *gin.Context) {
	requesterID := utils.GetCtx[string](c.Request.Context(), enums.ContextKeyUserID)
	sessions, appErr := ctrl.sessionService.FindResumableByUserID(c.Request.Context(), requesterID)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}
	var dtos []*res.SessionListItemRes
	if err := utils.MapToDTOs(sessions, &dtos); err != nil {
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal()))
		return
	}
	c.JSON(http.StatusOK, response.Success(dtos))
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

	c.Data(statusCode, "application/json", respBody)
}

// HandleCloseSession handles internal session close callbacks from the speech engine.
// No user auth — this is an internal endpoint on a private network.
func (ctrl *ModelGatewayController) HandleCloseSession(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var reqBody req.CloseSessionReq
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(errors.BadRequest("invalid request body")))
		return
	}
	reqBody.SessionID = sessionID

	if appErr := ctrl.svc.CloseSession(c.Request.Context(), &reqBody); appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	c.JSON(http.StatusOK, response.Success[any](nil))
}
