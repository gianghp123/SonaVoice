package repositories

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"gorm.io/gorm"
)

type sessionConfigRepository struct {
	db *gorm.DB
}

func NewSessionConfigRepository(db *gorm.DB) repository_interfaces.ISessionConfigRepository {
	return &sessionConfigRepository{
		db: db,
	}
}

func (r *sessionConfigRepository) Get(ctx context.Context) (*models.SessionConfig, error) {
	var model models.SessionConfig

	if err := r.db.FirstOrCreate(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func (r *sessionConfigRepository) Save(ctx context.Context, model *models.SessionConfig) error {
	return r.db.Save(model).Error
}
