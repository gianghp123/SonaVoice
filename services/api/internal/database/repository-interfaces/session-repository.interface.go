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
	GetForUpdate(ctx context.Context, sessionID string) (*models.Session, error)
	GetBySpeechSessionID(ctx context.Context, speechSessionId string) (*models.Session, error)
	GetPendingByUserID(ctx context.Context, userID string) (*models.Session, error)
	GetPendingByUserIDForUpdate(ctx context.Context, userID string) (*models.Session, error)
	UpdateSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) error
	UpdateStatus(ctx context.Context, sessionID string, status enums.SessionStatus) error
	SetSessionActive(ctx context.Context, sessionID string, startedAt time.Time) error
	SetQuotaDate(ctx context.Context, sessionID string, quotaDate time.Time) error
	SetReservedAmount(ctx context.Context, sessionID string, amount int64) error
	SetSessionFailed(ctx context.Context, sessionID string) error
	SetSessionInactive(ctx context.Context, sessionID string, endedAt time.Time) error
	SetQuotaDateToNil(ctx context.Context, sessionID string) error
}
