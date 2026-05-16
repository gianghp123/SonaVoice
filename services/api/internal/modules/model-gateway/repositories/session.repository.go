package repositories

import (
	"context"

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
