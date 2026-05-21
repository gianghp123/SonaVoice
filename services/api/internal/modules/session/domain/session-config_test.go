package domain

import (
	"testing"

	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestParseSessionConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  datatypes.JSON
		want    dtos.ConfigPayload
		wantErr bool
	}{
		{
			name:   "valid config",
			config: datatypes.JSON(`{"enabled":true,"reset_timezone":"UTC","limits":{"user":{"daily_voice_seconds":300,"daily_request_count":50},"session":{"max_session_lockTTL":5}}}`),
			want: dtos.ConfigPayload{
				Enabled:       true,
				ResetTimezone: "UTC",
				Limits: dtos.LimitsConfig{
					User:    dtos.UserLimitConfig{DailyVoiceSeconds: 300, DailyRequestCount: 50},
					Session: dtos.SessionLimitConfig{MaxSessionLockTTL: 5},
				},
			},
		},
		{
			name:    "empty config",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "invalid json",
			config:  datatypes.JSON(`{invalid}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSessionConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.config != nil {
				assert.Equal(t, tt.want.Enabled, result.Enabled)
				assert.Equal(t, tt.want.ResetTimezone, result.ResetTimezone)
				assert.Equal(t, tt.want.Limits.User.DailyVoiceSeconds, result.Limits.User.DailyVoiceSeconds)
				assert.Equal(t, tt.want.Limits.Session.MaxSessionLockTTL, result.Limits.Session.MaxSessionLockTTL)
			}
		})
	}
}
