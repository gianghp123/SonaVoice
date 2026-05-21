package domain

import (
	"testing"

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
		{"cannot close failed", enums.SessionStatusFailed, true, 400},
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


func TestSession_CanBeCancelled(t *testing.T) {
	tests := []struct {
		name    string
		status  enums.SessionStatus
		wantErr bool
	}{
		{"pending", enums.SessionStatusPending, false},
		{"active", enums.SessionStatusActive, false},
		{"inactive", enums.SessionStatusInactive, true},
		{"failed", enums.SessionStatusFailed, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{Status: tt.status}
			err := s.CanBeCancelled()
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
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
