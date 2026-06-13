package req

import "github.com/gianghp123/SonaVoice/api/internal/core/enums"

type GrammarAnalyzeQuery struct {
	ExplanationLanguage string `form:"explanation_language" json:"explanation_language"`
}

type ConversationTurn struct {
	Role enums.MessageRole `json:"role"`
	Text string            `json:"text"`
}

type GrammarAnalyzeTextReq struct {
	Transcript          string             `json:"transcript" binding:"required"`
	ExplanationLanguage string             `json:"explanation_language" binding:"required"`
	ConversationContext []ConversationTurn `json:"conversation_context" binding:"required"`
}

type GrammarAnalyzeReq struct {
	MessageID           string `json:"message_id" binding:"required"`
	ExplanationLanguage string `json:"explanation_language"`
}
