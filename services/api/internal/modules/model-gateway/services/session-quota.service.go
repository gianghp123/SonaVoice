package services

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/domain"
)

func today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

type ISessionQuotaService interface {
	Reserve(ctx context.Context, userID string, dailyQuota int) (int64, *errors.AppError)
	ReleaseAll(ctx context.Context, userID string, amount int64) *errors.AppError
	ReleaseWithActualUsage(ctx context.Context, session *models.Session, actualUsage int64) *errors.AppError
}

type sessionQuotaService struct {
	quotaRepo repository_interfaces.IUserQuotaRepository
}

func NewSessionQuotaService(quotaRepo repository_interfaces.IUserQuotaRepository) ISessionQuotaService {
	return &sessionQuotaService{quotaRepo: quotaRepo}
}

func (s *sessionQuotaService) Reserve(ctx context.Context, userID string, dailyQuota int) (int64, *errors.AppError) {
	logger := zapLogger.S()
	quotaDate := today()

	reservedAmount, err := s.quotaRepo.ReserveAll(ctx, userID, "voice", quotaDate, int64(dailyQuota))
	if err != nil {
		logger.Errorw("Failed to reserve quota", "userId", userID, "error", err)
		return 0, errors.Internal()
	}
	if reservedAmount <= 0 {
		return 0, errors.Forbidden("quota exceeded")
	}

	return reservedAmount, nil
}

func (s *sessionQuotaService) ReleaseAll(ctx context.Context, userID string, amount int64) *errors.AppError {
	if amount <= 0 {
		return nil
	}
	quotaDate := today()
	if err := s.quotaRepo.Release(ctx, userID, "voice", quotaDate, amount); err != nil {
		logger := zapLogger.S()
		logger.Errorw("Failed to release all quota", "userId", userID, "amount", amount, "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionQuotaService) ReleaseWithActualUsage(ctx context.Context, session *models.Session, actualUsage int64) *errors.AppError {
	domainSession := domain.NewSessionFromModel(session)
	if domainSession == nil {
		return errors.Internal()
	}

	if !domainSession.ShouldReleaseQuota() {
		return nil
	}

	reservedAmount := domainSession.ReservedAmount
	dailyQuota := domainSession.DailyQuota

	if reservedAmount <= 0 || dailyQuota <= 0 {
		return nil
	}

	clampedActual := domainSession.ClampActualUsage(actualUsage)
	unused := reservedAmount - clampedActual
	if unused < 0 {
		unused = 0
	}

	quotaDate := today()
	if err := s.quotaRepo.Release(ctx, session.UserID, "voice", quotaDate, unused); err != nil {
		logger := zapLogger.S()
		logger.Errorw("Failed to release quota with actual usage", "sessionId", session.ID, "error", err)
		return errors.Internal()
	}

	return nil
}
