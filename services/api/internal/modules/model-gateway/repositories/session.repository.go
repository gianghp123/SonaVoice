package repositories

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
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

func (s *sessionRepository) UpdateStatus(ctx context.Context, sessionID string, status enums.SessionStatus) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("status", status).Error
}

func (s *sessionRepository) SetSessionActive(ctx context.Context, sessionID string, startedAt time.Time) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"status":     enums.SessionStatusActive,
		"started_at": startedAt,
	}).Error
}

func (s *sessionRepository) SetQuotaDate(ctx context.Context, sessionID string, quotaDate time.Time) error {
	return s.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("quota_date", quotaDate).Error
}
