package controllers

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/services"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/gin-gonic/gin"
)

type GrammarController struct {
	svc services.IGrammarService
}

func NewGrammarController(svc services.IGrammarService) *GrammarController {
	return &GrammarController{svc: svc}
}

// HandleAnalyze godoc
// @Summary      Analyze grammar for a message
// @Description  Analyze a message transcript for grammar and naturalness, save the result
// @Security     BearerAuth
// @Tags         learning
// @Accept       json
// @Produce      json
// @Param        messageId           path   string true  "Message ID"
// @Param        explanationLanguage query  string false "Language for the explanation (e.g. vietnamese)"
// @Success      200  {object}  response.BaseResponse[res.GrammarAIResult]
// @Failure      400  {object}  response.BaseResponse[any]
// @Failure      401  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /learning/grammar/messages/{messageId} [post]
func (ctrl *GrammarController) HandleAnalyze(c *gin.Context) {
	messageID := c.Param("messageId")

	var query req.GrammarAnalyzeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(errors.BadRequest("invalid query params")))
		return
	}

	model, appErr := ctrl.svc.Analyze(c.Request.Context(), messageID, query.ExplanationLanguage)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	var dto res.GrammarAIResult
	if err := utils.MapToDTO(model, &dto); err != nil {
		sentry.CaptureException(err)
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal()))
		return
	}

	c.JSON(http.StatusOK, response.Success(&dto))
}

// HandleAnalyzeText godoc
// @Summary      Analyze grammar for text
// @Description  Analyze raw transcript text for grammar and naturalness without saving to DB
// @Security     BearerAuth
// @Tags         learning
// @Accept       json
// @Produce      json
// @Param        body  body  req.GrammarAnalyzeBody  true  "Transcript text to analyze"
// @Success      200  {object}  response.BaseResponse[res.GrammarAIResult]
// @Failure      400  {object}  response.BaseResponse[any]
// @Failure      401  {object}  response.BaseResponse[any]
// @Failure      500  {object}  response.BaseResponse[any]
// @Router       /learning/grammar/analyze [post]
func (ctrl *GrammarController) HandleAnalyzeText(c *gin.Context) {
	var body req.GrammarAnalyzeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(errors.BadRequest("invalid request body")))
		return
	}

	model, appErr := ctrl.svc.AnalyzeText(c.Request.Context(), body.Transcript, body.ExplanationLanguage)
	if appErr != nil {
		c.JSON(appErr.Code, response.Fail(appErr))
		return
	}

	var dto res.GrammarAIResult
	if err := utils.MapToDTO(model, &dto); err != nil {
		sentry.CaptureException(err)
		c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal()))
		return
	}

	c.JSON(http.StatusOK, response.Success(&dto))
}
