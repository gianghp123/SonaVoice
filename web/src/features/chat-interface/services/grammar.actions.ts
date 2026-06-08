'use server'

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"

export interface GrammarAnalysisState {
  result: IGrammarAnalysis | null
  error: string | null
  index: number
}

export async function analyzeGrammarAction(
  prevState: GrammarAnalysisState,
  formData: FormData
): Promise<GrammarAnalysisState> {
  const transcript = formData.get("transcript") as string
  const index = Number(formData.get("index"))
  const explanationLanguage = formData.get("explanationLanguage") as string | null

  if (!transcript) {
    return { result: null, error: "Transcript is required", index }
  }

  const response = await apiFetch<IGrammarAnalysis>(API_ROUTES.LEARNING.GRAMMAR.ANALYZE, {
    method: "POST",
    withCredentials: true,
    body: {
      transcript,
      explanationLanguage: explanationLanguage || undefined,
    },
  })

  if (response.error) {
    return { result: null, error: response.error.message, index }
  }

  return { result: response.data ?? null, error: null, index }
}
