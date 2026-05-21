package controllers

import (
	"io"
	"net/http"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/req"
	_ "github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/services"
	"github.com/gin-gonic/gin"
)

type SessionController struct {
	svc services.IOrchestratorService
}

func NewSessionController(svc services.IOrchestratorService) *SessionController {
	return &SessionController{svc: svc}
}

// HandleCreateSession godoc
// @Summary      Create new session and start connection
// @Description  Create a new session and start a WebRTC connection with the speech service
// @Security     Bearer
// @Tags         session
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.BaseResponse[res.CreateSessionRes]
// @Failure      401  {object}  response.BaseResponse[any]
// @Failure      403  {object}  response.BaseResponse[any]
// @Failure      409  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /sessions [post]
func (ctrl *SessionController) HandleCreateSession(c *gin.Context) {
	offer, appErr := ctrl.svc.CreateSession(c.Request.Context())
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}
	c.JSON(http.StatusOK, response.Success(offer))
}

// HandleStartConnection godoc
// @Summary      Start connection for an existing session
// @Description  Start a WebRTC connection for a pending session
// @Security     Bearer
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        sessionId  path      string  true  "Session ID"
// @Success      200  {object}  response.BaseResponse[res.CreateSessionRes]
// @Failure      400  {object}  response.BaseResponse[any]
// @Failure      403  {object}  response.BaseResponse[any]
// @Failure      409  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /sessions/{sessionId}/start [post]
func (ctrl *SessionController) HandleStartConnection(c *gin.Context) {
	sessionID := c.Param("sessionId")
	offer, appErr := ctrl.svc.StartConnection(c.Request.Context(), sessionID)
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
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        sessionId  path      string  true  "Session ID"
// @Param        body       body      any     true  "WebRTC offer request body"
// @Success      200        {object}  response.BaseResponse[any]
// @Failure      400        {object}  response.BaseResponse[any]
// @Failure      500        {object}  response.BaseResponse[any]
// @Router       /sessions/{sessionId}/offer [post]
// @Router       /sessions/{sessionId}/offer [patch]
func (ctrl *SessionController) HandleOffer(c *gin.Context) {
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
func (ctrl *SessionController) HandleCloseSession(c *gin.Context) {
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

// HandleCancelSession allows a user to cancel their own pending or active session.
// This releases all reserved quota and marks the session inactive.
func (ctrl *SessionController) HandleCancelSession(c *gin.Context) {
	sessionID := c.Param("sessionId")

	if appErr := ctrl.svc.CancelSession(c.Request.Context(), sessionID); appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	c.JSON(http.StatusOK, response.Success[any](nil))
}
