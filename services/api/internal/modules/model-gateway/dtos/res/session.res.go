package res

import (
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
)

type SessionListItemRes struct {
	ID             string              `json:"id"`
	Status         enums.SessionStatus `json:"status"`
	ReservedAmount int64               `json:"reserved_amount"`
	CreatedAt      time.Time           `json:"created_at"`
}

type SessionRes struct {
	ID             string              `json:"id"`
	UserID         string              `json:"user_id"`
	Status         enums.SessionStatus `json:"status"`
	ReservedAmount int64               `json:"reserved_amount"`
	CreatedAt      time.Time           `json:"created_at"`
}
