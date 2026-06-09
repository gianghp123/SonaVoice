package repositories

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ repository_interfaces.IGrammarAnalysisRepository = (*grammarAnalysisRepository)(nil)

type grammarAnalysisRepository struct {
	db *gorm.DB
}

func NewGrammarAnalysisRepository(db *gorm.DB) repository_interfaces.IGrammarAnalysisRepository {
	return &grammarAnalysisRepository{db: db}
}

func (r *grammarAnalysisRepository) Upsert(ctx context.Context, m *models.GrammarAnalysis) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"corrected_text", "explanation", "has_correction", "severity", "practice_sentence", "practice_focus", "practice_reason", "metadata", "updated_at"}),
		}).
		Create(m).Error
}

func (r *grammarAnalysisRepository) GetByID(ctx context.Context, id string) (*models.GrammarAnalysis, error) {
	m := new(models.GrammarAnalysis)
	if err := r.db.WithContext(ctx).First(m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (r *grammarAnalysisRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*models.GrammarAnalysis, error) {
	var analyses []*models.GrammarAnalysis
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&analyses).Error; err != nil {
		return nil, err
	}
	return analyses, nil
}

func (r *grammarAnalysisRepository) GetByMessageID(ctx context.Context, messageID string) (*models.GrammarAnalysis, error) {
	m := new(models.GrammarAnalysis)
	if err := r.db.WithContext(ctx).First(m, "message_id = ?", messageID).Error; err != nil {
		return nil, err
	}
	return m, nil
}
