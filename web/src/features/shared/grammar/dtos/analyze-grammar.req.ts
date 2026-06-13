export interface AnalyzeGrammarDto {
  transcript: string
  explanationLanguage?: string
  conversationContext: ConversationTurnDto[] 
}

export interface ConversationTurnDto {
  role: string
  text: string
}