package services

import (
	"context"

	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type IQuotaService interface {
	CheckRemaining(ctx context.Context, userID string, dailyLimit int64) (int64, error)
}

type quotaService struct {
	quotaRepo repository_interfaces.IUserQuotaRepository
}

func NewQuotaService(quotaRepo repository_interfaces.IUserQuotaRepository) IQuotaService {
	return &quotaService{quotaRepo: quotaRepo}
}

func (s *quotaService) CheckRemaining(ctx context.Context, userID string, dailyLimit int64) (int64, error) {
	quotaDate := utils.QuotaDate()
	return s.quotaRepo.GetOrCreate(ctx, userID, "voice", quotaDate, dailyLimit)
}
