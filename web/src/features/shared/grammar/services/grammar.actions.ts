"use server"

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import type { IGrammarAIResult, IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"
import { refresh, updateTag } from "next/cache"
import { tags } from "@/lib/tags"
import { AnalyzeGrammarDto } from "../dtos/analyze-grammar.req"


export async function analyzeGrammar(
  payload: AnalyzeGrammarDto,
) {
  return apiFetch<IGrammarAIResult>(
    API_ROUTES.LEARNING.GRAMMAR.ANALYZE,
    {
      method: "POST",
      withCredentials: true,
      body: payload
    }
  )
}


export async function analyzeGrammarByMessage(
  messageId: string,
  explanationLanguage?: string
) {
  const result = await apiFetch<IGrammarAnalysis>(
    API_ROUTES.LEARNING.GRAMMAR.BY_MESSAGE(messageId),
    {
      method: "POST",
      withCredentials: true,
      query: { explanationLanguage },
    }
  )

  if (!result.error) {
    updateTag(tags.grammarAnalyses)
    refresh()
  }

  return result
}