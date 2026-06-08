package schemas

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type GrammarAnalysisOutput struct {
	OriginalText     string                  `json:"original_text" jsonschema_description:"The original learner sentence from the speech-to-text transcript"`
	CorrectedText    string                  `json:"corrected_text" jsonschema_description:"A corrected version of the sentence in natural spoken English. If no correction is needed, repeat the original sentence"`
	Explanation      string                  `json:"explanation" jsonschema_description:"A short, simple explanation of the correction. If no correction is needed, explain that the sentence is already natural"`
	HasCorrection    bool                    `json:"has_correction" jsonschema_description:"Whether the original sentence contains a grammar or naturalness issue that should be corrected"`
	Severity         string                  `json:"severity" jsonschema:"enum=low,enum=medium,enum=high" jsonschema_description:"How important the correction is for communication or learning"`
	PracticeSentence string                  `json:"practice_sentence" jsonschema_description:"A sentence the learner should repeat for practice. Usually this should be the corrected sentence"`
	PracticeFocus    string                  `json:"practice_focus" jsonschema_description:"The main grammar or speaking skill to practice, such as past tense, articles, prepositions, word order, or natural phrasing"`
	PracticeReason   string                  `json:"practice_reason" jsonschema_description:"A short explanation of why this practice sentence is useful"`
	Metadata         grammarAnalysisMetadata `json:"metadata,nullable" jsonschema_description:"Additional metadata about the analysis, this is optional"`
}

type grammarAnalysisMetadata struct {
	Notes []string `json:"notes,omitempty" jsonschema_description:"Additional notes about the analysis, this is optional"`
}

var GrammarAnalysisOutputSchema = utils.GenerateSchema[GrammarAnalysisOutput]()
var GrammarAnalysisOutputSchemaName = "GrammarAnalysisOutput"
