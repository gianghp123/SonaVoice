package repositories

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/database"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"gorm.io/gorm"
)

var _ repository_interfaces.IMessageRepository = (*messageRepository)(nil)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) repository_interfaces.IMessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, m *models.Message) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *messageRepository) CreateBatch(ctx context.Context, msgs []*models.Message) error {
	return r.db.WithContext(ctx).Create(msgs).Error
}

func (r *messageRepository) ListBySessionID(ctx context.Context, sessionID string, q *database.Query) (*response.PaginatedResult[*models.Message], error) {
	tx := r.db.WithContext(ctx).Model(&models.Message{}).Where("session_id = ?", sessionID)

	total, err := q.Count(tx)
	if err != nil {
		return nil, err
	}

	tx = q.Apply(tx)

	var messages []*models.Message
	if err := tx.Find(&messages).Error; err != nil {
		return nil, err
	}

	meta := response.NewMeta(q.Page, q.Limit, total)
	return &response.PaginatedResult[*models.Message]{Data: messages, Meta: meta}, nil
}

func (r *messageRepository) GetByID(ctx context.Context, id string) (*models.Message, error) {
	message := new(models.Message)

	if err := r.db.WithContext(ctx).First(message, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return message, nil
}
