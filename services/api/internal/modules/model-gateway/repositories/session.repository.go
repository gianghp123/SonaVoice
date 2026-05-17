package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"gorm.io/gorm"
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

func (s *sessionRepository) Update(ctx context.Context, model *models.Session) error {
	return s.db.Save(model).Error
}

func (s *sessionRepository) GetBySpeechSessionID(ctx context.Context, speechSessionId string) (*models.Session, error) {
	var model models.Session

	if err := s.db.First(&model, "speech_session_id = ?", speechSessionId).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func (s *sessionRepository) FindStaleByUserID(ctx context.Context, userID string, pendingTimeoutSeconds int64) ([]*models.Session, error) {
	var sessions []*models.Session
	if err := s.db.Where(
		"user_id = ? AND quota_released = ? AND ((status = ? AND EXTRACT(EPOCH FROM (now() - created_at)) > ?) OR (status = ? AND started_at IS NOT NULL AND EXTRACT(EPOCH FROM (now() - started_at)) > reserved_amount))",
		userID, false,
		enums.SessionStatusPending, pendingTimeoutSeconds,
		enums.SessionStatusActive,
	).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *sessionRepository) UpdateSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("speech_session_id", speechSessionID).Error
}

func (s *sessionRepository) UpdateReservation(ctx context.Context, sessionID string, reservedAmount, dailyQuota int64) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"reserved_amount": reservedAmount,
		"daily_quota":     dailyQuota,
	}).Error
}

func (s *sessionRepository) UpdateStatus(ctx context.Context, sessionID string, status enums.SessionStatus) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("status", status).Error
}

func (s *sessionRepository) UpdateActiveSession(ctx context.Context, sessionID string, startedAt time.Time) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"status":     enums.SessionStatusActive,
		"started_at": startedAt,
	}).Error
}

func (s *sessionRepository) FindActiveByUserID(ctx context.Context, userID string) (*models.Session, error) {
	var model models.Session
	if err := s.db.Where("user_id = ? AND status IN ?", userID, []enums.SessionStatus{enums.SessionStatusActive, enums.SessionStatusPending}).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model, nil
}

func (s *sessionRepository) FindResumableByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	var sessions []*models.Session
	if err := s.db.Where("user_id = ? AND status = ?", userID, enums.SessionStatusInactive).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *sessionRepository) UpdateQuotaReleased(ctx context.Context, sessionID string) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("quota_released", true).Error
}
