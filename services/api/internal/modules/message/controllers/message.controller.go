package controllers

import (
	"net/http"

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
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal("failed to map messages")))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(messages, result.Meta))
}

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
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal("failed to map messages")))
		return
	}

	c.JSON(http.StatusCreated, response.Success(result))
}
