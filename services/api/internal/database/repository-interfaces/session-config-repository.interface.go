package repository_interfaces

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
)

type ISessionConfigRepository interface {
	Get(ctx context.Context) (*models.SessionConfig, error)
	Save(ctx context.Context, model *models.SessionConfig) error
}
