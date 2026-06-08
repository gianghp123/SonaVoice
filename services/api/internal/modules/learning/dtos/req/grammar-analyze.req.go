package req

type GrammarAnalyzeQuery struct {
	ExplanationLanguage string `form:"explanationLanguage" json:"explanationLanguage"`
}

type GrammarAnalyzeBody struct {
	Transcript          string `json:"transcript" binding:"required"`
	ExplanationLanguage string `json:"explanationLanguage"`
}
