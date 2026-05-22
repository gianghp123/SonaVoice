package res

import (
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
)

type MessageRes struct {
	ID             string            `json:"id"`
	SessionID      string            `json:"session_id"`
	Role           enums.MessageRole `json:"role"`
	Transcript     string            `json:"transcript"`
	WasInterrupted bool              `json:"was_interrupted"`
	CreatedAt      time.Time         `json:"created_at"`
}
