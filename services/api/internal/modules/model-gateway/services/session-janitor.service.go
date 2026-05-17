package services

import (
	"context"

	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/domain"
)

type ISessionJanitorService interface {
	CleanupStaleSessions(ctx context.Context, p transaction.IProvider, userID string, maxSessionLockTTL int64, dailyVoiceSeconds int64) error
}

type sessionJanitorService struct {
	quotaSvc ISessionQuotaService
}

func NewSessionJanitorService(quotaSvc ISessionQuotaService) ISessionJanitorService {
	return &sessionJanitorService{quotaSvc: quotaSvc}
}

func (s *sessionJanitorService) CleanupStaleSessions(ctx context.Context, p transaction.IProvider, userID string, maxSessionLockTTL int64, dailyVoiceSeconds int64) error {
	logger := zapLogger.S()
	sessionRepo := p.Session()

	staleSessions, err := sessionRepo.FindStaleByUserID(ctx, userID, maxSessionLockTTL)
	if err != nil {
		logger.Errorw("Failed to find stale sessions", "userId", userID, "error", err)
		return err
	}

	if len(staleSessions) == 0 {
		return nil
	}

	var totalUnused int64
	var staleIDs []string

	for _, ss := range staleSessions {
		ds := domain.NewSessionFromModel(ss)

		effectiveQuota := ds.DailyQuota
		if effectiveQuota <= 0 {
			effectiveQuota = dailyVoiceSeconds
		}
		reservedAmount := ds.ReservedAmount
		if reservedAmount <= 0 {
			reservedAmount = effectiveQuota
		}

		if ds.ShouldReleaseQuota() {
			totalUnused += reservedAmount
		}

		staleIDs = append(staleIDs, ss.ID)
	}

	if totalUnused > 0 {
		if appErr := s.quotaSvc.ReleaseAll(ctx, userID, totalUnused); appErr != nil {
			logger.Errorw("Failed to release stale session quota", "userId", userID, "totalUnused", totalUnused, "error", appErr)
			return appErr
		}
	}

	if err := sessionRepo.MarkStaleInactive(ctx, staleIDs); err != nil {
		logger.Errorw("Failed to mark stale sessions inactive", "userId", userID, "error", err)
		return err
	}

	return nil
}
