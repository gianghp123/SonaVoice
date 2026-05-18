package domain

import (
	"testing"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/stretchr/testify/assert"
)

func TestSession_CanBeStarted(t *testing.T) {
	tests := []struct {
		name    string
		status  enums.SessionStatus
		wantErr bool
		errCode int
	}{
		{"can start pending", enums.SessionStatusPending, false, 0},
		{"cannot start active", enums.SessionStatusActive, true, 400},
		{"cannot start inactive", enums.SessionStatusInactive, true, 400},
		{"cannot start failed", enums.SessionStatusFailed, true, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{Status: tt.status}
			err := s.CanBeStarted()
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

func TestSession_WantsQuotaRelease(t *testing.T) {
	s := &Session{QuotaDate: nil}
	assert.False(t, s.WantsQuotaRelease())
	now := time.Now()
	s.QuotaDate = &now
	assert.True(t, s.WantsQuotaRelease())
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
