package res

import (
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
)

type SessionRes struct {
	ID        string              `json:"id"`
	UserID    string              `json:"user_id"`
	Status    enums.SessionStatus `json:"status"`
	CreatedAt time.Time           `json:"created_at"`
}
