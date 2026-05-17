package res

import (
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
)

type SessionRes struct {
	ID             string              `json:"id"`
	UserID         string              `json:"user_id"`
	Status         enums.SessionStatus `json:"status"`
	ReservedAmount int64               `json:"reserved_amount"`
	DailyQuota     int64               `json:"daily_quota"`
	QuotaReleased  bool                `json:"quota_released"`
	CreatedAt      time.Time           `json:"created_at"`
}
