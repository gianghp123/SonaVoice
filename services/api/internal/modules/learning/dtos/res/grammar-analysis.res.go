package res

type GrammarAIResult struct {
	ID            string `json:"id,omitempty"`
	MessageID     string `json:"messageId,omitempty"`
	OriginalText  string `json:"originalText"`
	CorrectedText string `json:"correctedText,omitempty"`
	Explanation   string `json:"explanation,omitempty"`

	HasCorrection bool   `json:"hasCorrection"`
	Severity      string `json:"severity" validate:"oneof=low medium high"`

	PracticeSentence string `json:"practiceSentence,omitempty"`
	PracticeFocus    string `json:"practiceFocus,omitempty"`
	PracticeReason   string `json:"practiceReason,omitempty"`
	Metadata         any    `json:"metadata,omitempty"`
}
