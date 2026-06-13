package controllers

import (
	"io"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/services"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/gin-gonic/gin"
)

type SessionController struct {
	svc services.ISessionSevice
}

func NewSessionController(svc services.ISessionSevice) *SessionController {
	return &SessionController{svc: svc}
}

// HandleCreateSession godoc
// @Summary      Create new session and start connection
// @Description  Create a new session and start a WebRTC connection with the speech service
// @Security     BearerAuth
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
	c.JSON(http.StatusCreated, response.Success(offer))
}

// HandleStartConnection godoc
// @Summary      Start connection for an existing session
// @Description  Start a WebRTC connection for a pending session
// @Security     BearerAuth
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
// @Security     BearerAuth
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        sessionId  path      string  true  "Session ID"
// @Param        body       body      any     true  "WebRTC offer request body"
// @Success      200        {object}  response.BaseResponse[any]
// @Failure      400        {object}  response.BaseResponse[any]
// @Failure      500        {object}  response.BaseResponse[any]
// @Router       /sessions/{sessionId}/api/offer [post]
// @Router       /sessions/{sessionId}/api/offer [patch]
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

// HandleFinalizeSession handles internal session finalize callbacks from the speech engine.
// No user auth — this is an internal endpoint on a private network.
func (ctrl *SessionController) HandleFinalizeSession(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var reqBody req.FinalizeSessionReq
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(errors.BadRequest("invalid request body")))
		return
	}
	reqBody.SessionID = sessionID

	if appErr := ctrl.svc.FinalizeSession(c.Request.Context(), &reqBody); appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	c.JSON(http.StatusOK, response.Success[any](nil))
}

// HandleCancelSession godoc
// @Summary      Cancel a session
// @Description  Cancel a pending or active session, releasing all reserved quota
// @Security     BearerAuth
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        sessionId  path  string  true  "Session ID"
// @Success      200  {object}  response.BaseResponse[any]
// @Failure      401  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /sessions/{sessionId}/cancel [post]
func (ctrl *SessionController) HandleCancelSession(c *gin.Context) {
	sessionID := c.Param("sessionId")

	if appErr := ctrl.svc.CancelSession(c.Request.Context(), sessionID); appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	c.JSON(http.StatusOK, response.Success[any](nil))
}

// HandleListSessions godoc
// @Summary      List user sessions
// @Description  Get paginated list of sessions belonging to the authenticated user
// @Security     BearerAuth
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        page   query  int  false  "Page number (default 1)"
// @Param        limit  query  int  false  "Items per page (default 10, max 100)"
// @Success      200  {object}  response.BaseResponse[[]res.SessionListItemRes]
// @Failure      400  {object}  response.BaseResponse[any]
// @Failure      401  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /sessions [get]
func (ctrl *SessionController) HandleListSessions(c *gin.Context) {
	var query req.SessionListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(errors.BadRequest("invalid query params")))
		return
	}

	result, appErr := ctrl.svc.ListSessions(c.Request.Context(), query)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	var dto []*res.SessionListItemRes
	err := utils.MapToDTOs(result.Data, &dto)
	if err != nil {
		sentry.CaptureException(err)
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(dto, result.Meta))
}

// HandleGetSession godoc
// @Summary      Get session details
// @Description  Get a single session by ID (must belong to the authenticated user)
// @Security     BearerAuth
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        sessionId  path  string  true  "Session ID"
// @Success      200  {object}  response.BaseResponse[res.SessionRes]
// @Failure      400  {object}  response.BaseResponse[any]
// @Failure      401  {object}  response.BaseResponse[any]
// @Failure      403  {object}  response.BaseResponse[any]
// @Failure      404  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /sessions/{sessionId} [get]
func (ctrl *SessionController) HandleGetSession(c *gin.Context) {
	sessionID := c.Param("sessionId")

	session, appErr := ctrl.svc.GetSession(c.Request.Context(), sessionID)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	var dto res.SessionRes
	if err := utils.MapToDTO(session, &dto); err != nil {
		sentry.CaptureException(err)
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal()))
		return
	}

	c.JSON(http.StatusOK, response.Success(dto))
}
