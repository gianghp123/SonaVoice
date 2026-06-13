package services

import (
	"context"
	"encoding/json"

	openaiclient "github.com/gianghp123/SonaVoice/api/internal/clients/openai-client"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/prompts"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/schemas"
	"gorm.io/datatypes"
)

type IGrammarService interface {
	Analyze(ctx context.Context, messageID string, explanationLanguage string) (*models.GrammarAnalysis, *errors.AppError)
	AnalyzeText(ctx context.Context, body *req.GrammarAnalyzeTextReq) (*schemas.GrammarAnalysisOutput, *errors.AppError)
	GetBySessionID(ctx context.Context, sessionID string) ([]*models.GrammarAnalysis, *errors.AppError)
}

type grammarService struct {
	messageRepo  repository_interfaces.IMessageRepository
	grammarRepo  repository_interfaces.IGrammarAnalysisRepository
	openaiClient openaiclient.IOpenAIClient
}

func NewGrammarService(
	openaiclient openaiclient.IOpenAIClient,
	messageRepo repository_interfaces.IMessageRepository,
	grammarRepo repository_interfaces.IGrammarAnalysisRepository,
) IGrammarService {
	return &grammarService{
		openaiClient: openaiclient,
		messageRepo:  messageRepo,
		grammarRepo:  grammarRepo,
	}
}

func (s *grammarService) AnalyzeText(ctx context.Context, body *req.GrammarAnalyzeTextReq) (*schemas.GrammarAnalysisOutput, *errors.AppError) {
	logger := zapLogger.S()

	logger.Infow("sending request to openai client", "explanation_language", body.ExplanationLanguage)

	prompt, err := prompts.BuildGrammarAnalysisPrompt(body.Transcript, body.ExplanationLanguage, body.ConversationContext)
	if err != nil {
		logger.Errorw("failed to build grammar analysis prompt", "error", err)
		return nil, errors.Internal(err.Error())
	}

	var result schemas.GrammarAnalysisOutput

	_, err = openaiclient.
		NewResponse(ctx, s.openaiClient).
		Prompt(prompt).
		WithStructuredOutput(
			&result,
			schemas.GrammarAnalysisOutputSchemaName,
			schemas.GrammarAnalysisOutputSchema,
		).
		Do()

	if err != nil {
		logger.Errorw("failed to analyze grammar", "error", err)
		return nil, errors.Internal(err.Error())
	}

	return &result, nil
}

func (s *grammarService) Analyze(ctx context.Context, messageID string, explanationLanguage string) (*models.GrammarAnalysis, *errors.AppError) {
	logger := zapLogger.S()

	logger.Infow("analyzing grammar for message", "message_id", messageID, "explanation_language", explanationLanguage)
	message, err := s.messageRepo.GetByID(ctx, messageID)

	if err != nil {
		logger.Errorw("failed to get message", "error", err)
		return nil, errors.MapRepoError(err)
	}

	recentMessages, err := s.messageRepo.ListBySessionID(
		ctx,
		message.SessionID,
		database.NewQuery().
			SetOrderBy("created_at DESC").
			SetLimit(5),
	)

	if err != nil {
		logger.Errorw("failed to get recent messages", "error", err)
		return nil, errors.MapRepoError(err)
	}

	conversationContext := make([]req.ConversationTurn, len(recentMessages.Data))
	for i, msg := range recentMessages.Data {
		conversationContext[i] = req.ConversationTurn{
			Role: msg.Role,
			Text: msg.Transcript,
		}
	}

	analyzeTextBody := &req.GrammarAnalyzeTextReq{
		Transcript:          message.Transcript,
		ExplanationLanguage: explanationLanguage,
		ConversationContext: conversationContext,
	}

	aiResult, appErr := s.AnalyzeText(ctx, analyzeTextBody)
	if appErr != nil {
		return nil, appErr
	}

	resultJSON, err := json.Marshal(aiResult)
	if err != nil {
		logger.Errorw("failed to marshal grammar result", "error", err)
		return nil, errors.Internal(err.Error())
	}

	model := &models.GrammarAnalysis{
		SessionID:    message.SessionID,
		MessageID:    messageID,
		OriginalText: message.Transcript,
		Result:       datatypes.JSON(resultJSON),
	}

	if err := s.grammarRepo.Upsert(ctx, model); err != nil {
		logger.Errorw("failed to save grammar analysis", "error", err)
		return nil, errors.MapRepoError(err)
	}

	return model, nil
}

func (s *grammarService) GetBySessionID(ctx context.Context, sessionID string) ([]*models.GrammarAnalysis, *errors.AppError) {
	logger := zapLogger.S()

	logger.Infow("fetching grammar analyses for session", "session_id", sessionID)
	analyses, err := s.grammarRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		logger.Errorw("failed to get grammar analyses by session", "error", err, "session_id", sessionID)
		return nil, errors.MapRepoError(err)
	}

	return analyses, nil
}
