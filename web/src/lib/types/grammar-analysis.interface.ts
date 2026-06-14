export interface IGrammarAIResult {
  cleanedTranscript?: string
  hasTranscriptCleanup: boolean
  hasCorrection: boolean
  issueType?: string
  correctedText?: string
  practiceSentence?: string
  severity?: 'low' | 'medium' | 'high'
  practiceFocus?: string
  explanation?: string
  practiceReason?: string
}

export interface IGrammarAnalysis extends IGrammarAIResult {
  id?: string
  messageId?: string
  originalText: string
}
