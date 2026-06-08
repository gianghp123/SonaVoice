package services

import (
	"context"
	"encoding/json"

	openaiclient "github.com/gianghp123/SonaVoice/api/internal/clients/openai-client"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/prompts"
	"github.com/gianghp123/SonaVoice/api/internal/modules/learning/schemas"
)

type IGrammarService interface {
	Analyze(ctx context.Context, messageID string, explanationLanguage string) (*models.GrammarAnalysis, *errors.AppError)
	AnalyzeText(ctx context.Context, transcript string, explanationLanguage string) (*models.GrammarAnalysis, *errors.AppError)
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

func (s *grammarService) AnalyzeText(ctx context.Context, transcript string, explanationLanguage string) (*models.GrammarAnalysis, *errors.AppError) {
	logger := zapLogger.S()

	prompt, err := prompts.BuildGrammarAnalysisPrompt(transcript, explanationLanguage)
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

	metadataJSON, err := json.Marshal(result.Metadata)
	if err != nil {
		logger.Errorw("failed to marshal metadata", "error", err)
		return nil, errors.Internal(err.Error())
	}

	model := &models.GrammarAnalysis{
		OriginalText:     result.OriginalText,
		CorrectedText:    result.CorrectedText,
		Explanation:      result.Explanation,
		HasCorrection:    result.HasCorrection,
		Severity:         result.Severity,
		PracticeSentence: result.PracticeSentence,
		PracticeFocus:    result.PracticeFocus,
		PracticeReason:   result.PracticeReason,
		Metadata:         metadataJSON,
	}

	return model, nil
}

func (s *grammarService) Analyze(ctx context.Context, messageID string, explanationLanguage string) (*models.GrammarAnalysis, *errors.AppError) {
	logger := zapLogger.S()

	logger.Infow("analyzing grammar for message", "message_id", messageID, "explanation_language", explanationLanguage)
	message, err := s.messageRepo.GetByID(ctx, messageID)

	if err != nil {
		logger.Errorw("failed to get message", "error", err)
		return nil, errors.MapRepoError(err)
	}

	model, appErr := s.AnalyzeText(ctx, message.Transcript, explanationLanguage)
	if appErr != nil {
		return nil, appErr
	}

	model.SessionID = message.SessionID
	model.MessageID = messageID

	if err := s.grammarRepo.Upsert(ctx, model); err != nil {
		logger.Errorw("failed to save grammar analysis", "error", err)
		return nil, errors.MapRepoError(err)
	}

	return model, nil
}
