package req

import "github.com/gianghp123/SonaVoice/api/internal/core/enums"

type MessageListQuery struct {
	Page  int    `form:"page" json:"page"`
	Limit int    `form:"limit" json:"limit"`
	Order string `form:"order" json:"order"`
}

type MessageItem struct {
	Role           enums.MessageRole `json:"role" binding:"required"`
	Transcript     string            `json:"transcript"`
	WasInterrupted bool              `json:"was_interrupted"`
}

type CreateMessagesReq struct {
	Messages []MessageItem `json:"messages" binding:"required"`
}
