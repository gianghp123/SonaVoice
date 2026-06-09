import "server-only"

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import { tags } from "@/lib/tags"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"

export async function getGrammarAnalyses(sessionId: string) {
  return apiFetch<IGrammarAnalysis[]>(
    API_ROUTES.LEARNING.GRAMMAR.BY_SESSION(sessionId),
    {
      withCredentials: true,
      next: { tags: [tags.grammarAnalyses] },
    }
  )
}
