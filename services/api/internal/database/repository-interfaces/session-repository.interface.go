package repository_interfaces

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
)

type ISessionRepository interface {
	Create(ctx context.Context, model *models.Session) error
	Get(ctx context.Context, sessionId string) (*models.Session, error)
	Update(ctx context.Context, model *models.Session) error
	GetBySpeechSessionID(ctx context.Context, speechSessionId string) (*models.Session, error)
	FindStaleByUserID(ctx context.Context, userID string, pendingTimeoutSeconds int64) ([]*models.Session, error)
	FindActiveByUserID(ctx context.Context, userID string) (*models.Session, error)
	FindResumableByUserID(ctx context.Context, userID string) ([]*models.Session, error)
	UpdateSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) error
	UpdateReservation(ctx context.Context, sessionID string, reservedAmount, dailyQuota int64) error
	UpdateStatus(ctx context.Context, sessionID string, status enums.SessionStatus) error
	UpdateActiveSession(ctx context.Context, sessionID string, startedAt time.Time) error
	UpdateQuotaReleased(ctx context.Context, sessionID string) error
}