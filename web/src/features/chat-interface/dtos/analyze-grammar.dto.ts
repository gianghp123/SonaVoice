import { SupportedLanguage } from "@/lib/i18n/i18n"

export interface AnalyzeGrammarDto {
  transcript: string
  explainationLanguage?: string
}