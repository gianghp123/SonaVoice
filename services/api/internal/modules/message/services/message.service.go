package services

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/database"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/modules/message/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type IMessageService interface {
	List(ctx context.Context, sessionID string, q req.MessageListQuery) (*response.PaginatedResult[*models.Message], *errors.AppError)
	Create(ctx context.Context, sessionID string, body *req.CreateMessagesReq) ([]*models.Message, *errors.AppError)
}

type messageService struct {
	repo        repository_interfaces.IMessageRepository
	sessionRepo repository_interfaces.ISessionRepository
}

func NewMessageService(repo repository_interfaces.IMessageRepository, sessionRepo repository_interfaces.ISessionRepository) IMessageService {
	return &messageService{repo: repo, sessionRepo: sessionRepo}
}

func (s *messageService) List(ctx context.Context, sessionID string, q req.MessageListQuery) (*response.PaginatedResult[*models.Message], *errors.AppError) {
	userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, errors.MapRepoError(err)
	}

	if appErr := utils.EnforceOwnership(session.UserID, userID); appErr != nil {
		return nil, appErr
	}

	dbQuery := database.NewQuery().
		SetPage(q.Page).
		SetLimit(q.Limit)

	if q.Order != "" {
		dbQuery.SetOrderBy(q.Order)
	} else {
		dbQuery.SetOrderBy("created_at ASC")
	}

	result, err := s.repo.ListBySessionID(ctx, sessionID, dbQuery)
	if err != nil {
		return nil, errors.MapRepoError(err)
	}

	return result, nil
}

func (s *messageService) Create(ctx context.Context, sessionID string, body *req.CreateMessagesReq) ([]*models.Message, *errors.AppError) {
	userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	if len(body.Messages) == 0 {
		return nil, errors.BadRequest("messages must not be empty")
	}

	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, errors.MapRepoError(err)
	}

	if appErr := utils.EnforceOwnership(session.UserID, userID); appErr != nil {
		return nil, appErr
	}

	var messages []*models.Message
	for _, item := range body.Messages {
		msg := &models.Message{
			SessionID:      sessionID,
			Role:           item.Role,
			Transcript:     item.Transcript,
			WasInterrupted: item.WasInterrupted,
		}
		messages = append(messages, msg)
	}

	if err := s.repo.CreateBatch(ctx, messages); err != nil {
		return nil, errors.MapRepoError(err)
	}

	return messages, nil
}
