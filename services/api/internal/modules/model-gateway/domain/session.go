package domain

import (
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
)

type Session struct {
	ID             string
	UserID         string
	Status         enums.SessionStatus
	ReservedAmount int64
	DailyQuota     int64
	QuotaReleased  bool
}

func (s *Session) IsOwnedBy(userID string) bool {
	return s.UserID == userID
}

func (s *Session) CanBeResumedBy(userID string) *errors.AppError {
	if !s.IsOwnedBy(userID) {
		return errors.Forbidden()
	}
	if s.Status != enums.SessionStatusInactive {
		return errors.BadRequest("session is not resumable")
	}
	return nil
}

func (s *Session) CanBeClosed() *errors.AppError {
	if s.Status == enums.SessionStatusInactive {
		return errors.BadRequest("session is already inactive")
	}
	return nil
}

func (s *Session) ShouldReleaseQuota() bool {
	return !s.QuotaReleased
}

func (s *Session) ClampActualUsage(actual int64) int64 {
	if actual < 0 {
		return 0
	}
	if actual > s.ReservedAmount {
		return s.ReservedAmount
	}
	return actual
}

func NewSessionFromModel(m *models.Session) *Session {
	if m == nil {
		return nil
	}
	return &Session{
		ID:             m.ID,
		UserID:         m.UserID,
		Status:         m.Status,
		ReservedAmount: m.ReservedAmount,
		DailyQuota:     m.DailyQuota,
		QuotaReleased:  m.QuotaReleased,
	}
}
