package repositories

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"gorm.io/gorm"
)

type IGlobalConfigRepository interface {
	Get(ctx context.Context) (*models.GlobalConfig, error)
	Save(ctx context.Context, model *models.GlobalConfig) error
}

type globalConfigRepository struct {
	db *gorm.DB
}

func NewGlobalConfigRepository(db *gorm.DB) IGlobalConfigRepository {
	return &globalConfigRepository{
		db: db,
	}
}

func (r *globalConfigRepository) Get(ctx context.Context) (*models.GlobalConfig, error) {
	var model models.GlobalConfig

	if err := r.db.FirstOrCreate(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func (r *globalConfigRepository) Save(ctx context.Context, model *models.GlobalConfig) error {
	return r.db.Save(model).Error
}
