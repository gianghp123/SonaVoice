package schemas

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type GrammarAnalysisOutput struct {
	CleanedTranscript    string `json:"cleaned_transcript" jsonschema_description:"A cleaned version of the transcript with STT noise removed"`
	HasTranscriptCleanup bool   `json:"has_transcript_cleanup" jsonschema_description:"Whether the transcript was cleaned of STT noise"`
	HasCorrection        bool   `json:"has_correction" jsonschema_description:"Whether the normalized sentence contains a grammar or naturalness issue"`
	IssueType            string `json:"issue_type" jsonschema:"enum=none,enum=grammar,enum=tense,enum=article,enum=word_order,enum=word_choice,enum=incomplete_sentence,enum=pronunciation,enum=mixed" jsonschema_description:"The type of issue found"`
	CorrectedText        string `json:"corrected_text" jsonschema_description:"A corrected version of the sentence in natural spoken English"`
	PracticeSentence     string `json:"practice_sentence" jsonschema_description:"A sentence the learner should repeat"`
	Severity             string `json:"severity" jsonschema:"enum=low,enum=medium,enum=high" jsonschema_description:"How important the correction is"`
	PracticeFocus        string `json:"practice_focus" jsonschema_description:"The main grammar or speaking skill to practice"`
	Explanation          string `json:"explanation" jsonschema_description:"A short, simple explanation of the correction"`
	PracticeReason       string `json:"practice_reason" jsonschema_description:"A short explanation of why this practice sentence is useful"`
}

var GrammarAnalysisOutputSchema = utils.GenerateSchema[GrammarAnalysisOutput]()
var GrammarAnalysisOutputSchemaName = "GrammarAnalysisOutput"
