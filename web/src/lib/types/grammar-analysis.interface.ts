export interface IGrammarAnalysis {
  id?: string
  messageId?: string
  originalText: string
  correctedText: string
  explanation: string
  hasCorrection: boolean
  severity: 'low' | 'medium' | 'high'
  practiceSentence?: string
  practiceFocus?: string
  practiceReason?: string
}
