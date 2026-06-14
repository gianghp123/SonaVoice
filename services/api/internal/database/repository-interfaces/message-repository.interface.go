package repository_interfaces

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/database"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
)

type IMessageRepository interface {
	Create(ctx context.Context, m *models.Message) error
	CreateBatch(ctx context.Context, msgs []*models.Message) error
	ListBySessionID(ctx context.Context, sessionID string, q *database.Query) (*response.PaginatedResult[*models.Message], error)
	GetByID(ctx context.Context, id string) (*models.Message, error)
}
