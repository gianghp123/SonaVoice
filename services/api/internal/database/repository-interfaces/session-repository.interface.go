package repository_interfaces

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
)

type ISessionRepository interface {
	Create(ctx context.Context, model *models.Session) error
	Get(ctx context.Context, sessionId string) (*models.Session, error)
	Update(ctx context.Context, model *models.Session) error
	GetBySpeechSessionID(ctx context.Context, speechSessionId string) (*models.Session, error)
	FindStaleByUserID(ctx context.Context, userID string) ([]*models.Session, error)
}