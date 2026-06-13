package repository_interfaces

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/database"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"gorm.io/datatypes"
)

type ISessionRepository interface {
	Create(ctx context.Context, model *models.Session) error
	Get(ctx context.Context, sessionId string) (*models.Session, error)
	GetForUpdate(ctx context.Context, sessionID string) (*models.Session, error)
	GetBySpeechSessionID(ctx context.Context, speechSessionId string) (*models.Session, error)
	GetActiveOrPendingByUserIDForUpdate(ctx context.Context, userID string) (*models.Session, error)
	UpdateSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) error
	UpdateStatus(ctx context.Context, sessionID string, status enums.SessionStatus) error
	SetSessionActive(ctx context.Context, sessionID string, startedAt time.Time) error
	SetQuotaDate(ctx context.Context, sessionID string, quotaDate time.Time) error
	SetMaxDuration(ctx context.Context, sessionID string, maxDuration int64) error
	SetActualUsage(ctx context.Context, sessionID string, actualUsage int64) error
	SetSessionFailed(ctx context.Context, sessionID string) error
	SetSessionInactive(ctx context.Context, sessionID string, endedAt time.Time) error
	UpdateSpeechStartResponse(ctx context.Context, sessionID string, speechStartResponse datatypes.JSON) error
	List(ctx context.Context, q *database.Query) (*response.PaginatedResult[*models.Session], error)
}
