package controllers

import (
	"net/http"

	"github.com/getsentry/sentry-go"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/modules/message/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/message/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/message/services"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/gin-gonic/gin"
)

type MessageController struct {
	svc services.IMessageService
}

func NewMessageController(svc services.IMessageService) *MessageController {
	return &MessageController{svc: svc}
}

// HandleListMessages godoc
// @Summary      List messages in a session
// @Description  Get paginated list of messages for a session
// @Security     BearerAuth
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        sessionId  path  string  true  "Session ID"
// @Param        page       query int     false "Page number (default 1)"
// @Param        limit      query int     false "Items per page (default 10, max 100)"
// @Param        order      query string  false "Sort order (asc/desc)"
// @Success      200  {object}  response.BaseResponse[[]res.MessageRes]
// @Failure      400  {object}  response.BaseResponse[any]
// @Failure      401  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /sessions/{sessionId}/messages [get]
func (ctrl *MessageController) HandleListMessages(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var query req.MessageListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(errors.BadRequest("invalid query params")))
		return
	}

	result, appErr := ctrl.svc.List(c.Request.Context(), sessionID, query)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	var messages []res.MessageRes
	if err := utils.MapToDTOs(result.Data, &messages); err != nil {
		sentry.CaptureException(err)
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal("failed to map messages")))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(messages, result.Meta))
}

// HandleCreateMessages godoc
// @Summary      Create messages in a session
// @Description  Create new messages for a session (internal endpoint)
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        sessionId  path  string              true  "Session ID"
// @Param        body       body  req.CreateMessagesReq true "Messages to create"
// @Success      201  {object}  response.BaseResponse[[]res.MessageRes]
// @Failure      400  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /sessions/{sessionId}/messages [post]
func (ctrl *MessageController) HandleCreateMessages(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var body req.CreateMessagesReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(errors.BadRequest("invalid request body")))
		return
	}

	msgs, appErr := ctrl.svc.Create(c.Request.Context(), sessionID, &body)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	var result []res.MessageRes
	if err := utils.MapToDTOs(msgs, &result); err != nil {
		sentry.CaptureException(err)
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal("failed to map messages")))
		return
	}

	c.JSON(http.StatusCreated, response.Success(result))
}
