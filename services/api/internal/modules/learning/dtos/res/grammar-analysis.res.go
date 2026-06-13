package res

type GrammarAIResult struct {
	CleanedTranscript    string `json:"cleaned_transcript,omitempty"`
	HasTranscriptCleanup bool   `json:"has_transcript_cleanup"`
	HasCorrection        bool   `json:"has_correction"`
	IssueType            string `json:"issue_type,omitempty"`
	CorrectedText        string `json:"corrected_text,omitempty"`
	PracticeSentence     string `json:"practice_sentence,omitempty"`
	Severity             string `json:"severity,omitempty"`
	PracticeFocus        string `json:"practice_focus,omitempty"`
	Explanation          string `json:"explanation,omitempty"`
	PracticeReason       string `json:"practice_reason,omitempty"`
}

type GrammarAnalyzeRes struct {
	ID           string `json:"id,omitempty"`
	MessageID    string `json:"message_id,omitempty"`
	OriginalText string `json:"original_text"`
	GrammarAIResult
}
