"use server"

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"
import { AnalyzeGrammarDto } from "../dtos/analyze-grammar.dto"

export interface GrammarAnalysisState {
  analyses: Record<number, IGrammarAnalysis>
  error: string | null
}

export async function analyzeGrammar(
  payload: AnalyzeGrammarDto,
) {
  return apiFetch<IGrammarAnalysis>(
    API_ROUTES.LEARNING.GRAMMAR.ANALYZE,
    {
      method: "POST",
      withCredentials: true,
      body: payload
    }
  )
}