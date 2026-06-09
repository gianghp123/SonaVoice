package repository_interfaces

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
)

type IGrammarAnalysisRepository interface {
	Upsert(ctx context.Context, m *models.GrammarAnalysis) error
	GetByID(ctx context.Context, id string) (*models.GrammarAnalysis, error)
	GetByMessageID(ctx context.Context, messageID string) (*models.GrammarAnalysis, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*models.GrammarAnalysis, error)
}
