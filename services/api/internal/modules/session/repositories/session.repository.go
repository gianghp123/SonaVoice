package repositories

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/database"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) repository_interfaces.ISessionRepository {
	return &sessionRepository{
		db: db,
	}
}

func (s *sessionRepository) Create(ctx context.Context, model *models.Session) error {
	return s.db.Create(model).Error
}

func (s *sessionRepository) Get(ctx context.Context, sessionId string) (*models.Session, error) {
	var model models.Session
	if err := s.db.First(&model, "id = ?", sessionId).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *sessionRepository) GetBySpeechSessionID(ctx context.Context, speechSessionId string) (*models.Session, error) {
	var model models.Session
	if err := s.db.First(&model, "speech_session_id = ?", speechSessionId).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *sessionRepository) GetForUpdate(ctx context.Context, sessionID string) (*models.Session, error) {
	var model models.Session
	if err := s.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&model, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *sessionRepository) UpdateSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("speech_session_id", speechSessionID).Error
}

func (s *sessionRepository) SetMaxDuration(ctx context.Context, sessionID string, maxDuration int64) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("max_duration", maxDuration).Error
}

func (s *sessionRepository) SetActualUsage(ctx context.Context, sessionID string, actualUsage int64) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("actual_usage", actualUsage).Error
}

func (s *sessionRepository) UpdateStatus(ctx context.Context, sessionID string, status enums.SessionStatus) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("status", status).Error
}

func (s *sessionRepository) SetSessionActive(ctx context.Context, sessionID string, startedAt time.Time) error {
	result := s.db.Model(&models.Session{}).Where("id = ? AND status = ?", sessionID, enums.SessionStatusPending).Updates(map[string]interface{}{
		"status":     enums.SessionStatusActive,
		"started_at": startedAt,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *sessionRepository) SetQuotaDate(ctx context.Context, sessionID string, quotaDate time.Time) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("quota_date", quotaDate).Error
}

func (s *sessionRepository) SetSessionFailed(ctx context.Context, sessionID string) error {
	result := s.db.Model(&models.Session{}).Where("id = ? AND status IN ?", sessionID, []enums.SessionStatus{enums.SessionStatusPending, enums.SessionStatusActive}).	Updates(map[string]interface{}{
		"status":   enums.SessionStatusFailed,
		"ended_at": utils.NowUTC(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *sessionRepository) GetPendingByUserID(ctx context.Context, userID string) (*models.Session, error) {
	var model models.Session
	if err := s.db.First(&model, "user_id = ? AND status = ?", userID, enums.SessionStatusPending).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *sessionRepository) GetPendingByUserIDForUpdate(ctx context.Context, userID string) (*models.Session, error) {
	var model models.Session
	if err := s.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&model, "user_id = ? AND status = ?", userID, enums.SessionStatusPending).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *sessionRepository) List(ctx context.Context, q *database.Query) (*response.PaginatedResult[*models.Session], error) {
	tx := s.db.WithContext(ctx).Model(&models.Session{})
	total, err := q.Count(tx)
	if err != nil {
		return nil, err
	}
	tx = q.Apply(tx)
	var sessions []*models.Session
	if err := tx.Find(&sessions).Error; err != nil {
		return nil, err
	}
	meta := response.NewMeta(q.Page, q.Limit, total)
	return &response.PaginatedResult[*models.Session]{Data: sessions, Meta: meta}, nil
}

func (s *sessionRepository) SetSessionInactive(ctx context.Context, sessionID string, endedAt time.Time) error {
	result := s.db.Model(&models.Session{}).Where("id = ? AND status != ?", sessionID, enums.SessionStatusInactive).	Updates(map[string]interface{}{
		"status":   enums.SessionStatusInactive,
		"ended_at": endedAt,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
