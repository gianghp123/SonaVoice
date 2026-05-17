package domain

import (
	"testing"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/stretchr/testify/assert"
)

func TestSession_CanBeResumedBy(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		status  enums.SessionStatus
		wantErr bool
		errCode int
	}{
		{"owner can resume inactive", "user-1", enums.SessionStatusInactive, false, 0},
		{"different user cannot resume", "user-2", enums.SessionStatusInactive, true, 403},
		{"cannot resume active session", "user-1", enums.SessionStatusActive, true, 400},
		{"cannot resume pending session", "user-1", enums.SessionStatusPending, true, 400},
		{"cannot resume failed session", "user-1", enums.SessionStatusFailed, true, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{UserID: "user-1", Status: tt.status}
			err := s.CanBeResumedBy(tt.userID)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.errCode, err.Code)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestSession_CanBeClosed(t *testing.T) {
	tests := []struct {
		name    string
		status  enums.SessionStatus
		wantErr bool
		errCode int
	}{
		{"can close active", enums.SessionStatusActive, false, 0},
		{"can close pending", enums.SessionStatusPending, false, 0},
		{"can close failed", enums.SessionStatusFailed, false, 0},
		{"cannot close inactive", enums.SessionStatusInactive, true, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{Status: tt.status}
			err := s.CanBeClosed()
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.errCode, err.Code)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestSession_ClampActualUsage(t *testing.T) {
	s := &Session{ReservedAmount: 300}
	assert.Equal(t, int64(0), s.ClampActualUsage(-10))
	assert.Equal(t, int64(0), s.ClampActualUsage(0))
	assert.Equal(t, int64(60), s.ClampActualUsage(60))
	assert.Equal(t, int64(300), s.ClampActualUsage(300))
	assert.Equal(t, int64(300), s.ClampActualUsage(500))
}

func TestSession_ShouldReleaseQuota(t *testing.T) {
	s := &Session{QuotaReleased: false}
	assert.True(t, s.ShouldReleaseQuota())
	s.QuotaReleased = true
	assert.False(t, s.ShouldReleaseQuota())
}

func TestSession_IsOwnedBy(t *testing.T) {
	s := &Session{UserID: "user-1"}
	assert.True(t, s.IsOwnedBy("user-1"))
	assert.False(t, s.IsOwnedBy("user-2"))
}

func TestNewSessionFromModel_Nil(t *testing.T) {
	s := NewSessionFromModel(nil)
	assert.Nil(t, s)
}
