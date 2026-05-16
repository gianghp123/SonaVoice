package repository_interfaces

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
)

type IGlobalConfigRepository interface {
	Get(ctx context.Context) (*models.GlobalConfig, error)
	Save(ctx context.Context, model *models.GlobalConfig) error
}
