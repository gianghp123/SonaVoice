package domain

import (
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
)

type Session struct {
	ID          string
	UserID      string
	Status      enums.SessionStatus
	MaxDuration int64
	ActualUsage int64
	QuotaDate   *time.Time
}

func (s *Session) IsOwnedBy(userID string) bool {
	return s.UserID == userID
}

func (s *Session) CanBeStarted() *errors.AppError {
	if s.Status != enums.SessionStatusPending {
		return errors.BadRequest("session is not startable")
	}
	return nil
}

func (s *Session) CanBeClosed() *errors.AppError {
	if s.Status == enums.SessionStatusInactive {
		return errors.BadRequest("session is already closed")
	}
	if s.Status == enums.SessionStatusFailed {
		return errors.BadRequest("session has already failed")
	}
	return nil
}

func (s *Session) CanBeCancelled() *errors.AppError {
	if s.Status == enums.SessionStatusInactive || s.Status == enums.SessionStatusFailed {
		return errors.BadRequest("session is already closed")
	}
	return nil
}

func NewSessionFromModel(m *models.Session) *Session {
	if m == nil {
		return nil
	}
	return &Session{
		ID:          m.ID,
		UserID:      m.UserID,
		Status:      m.Status,
		MaxDuration: m.MaxDuration,
		ActualUsage: m.ActualUsage,
		QuotaDate:   m.QuotaDate,
	}
}
