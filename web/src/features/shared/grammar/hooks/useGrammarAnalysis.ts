import { useT } from "next-i18next/client"
import { useState } from "react"
import { toast } from "sonner"
import { analyzeGrammarByMessage } from "../services/grammar.actions"
import {
  FALLBACK_LANGUAGE,
  isSupportedLanguage,
  LANGUAGE_FULL_NAMES,
  SupportedLanguage,
} from "@/lib/i18n/i18n"

export function useGrammarAnalysis() {
  const { i18n } = useT()
  const [pendingId, setPendingId] = useState<string | null>(null)

  const explanationLanguage =
    LANGUAGE_FULL_NAMES[
      isSupportedLanguage(i18n.language)
        ? (i18n.language as SupportedLanguage)
        : FALLBACK_LANGUAGE
    ]

  const triggerAnalysis = async (messageId: string) => {
    setPendingId(messageId)

    try {
      const response = await analyzeGrammarByMessage(
        messageId,
        explanationLanguage
      )

      if (response.error) {
        toast.error(response.error.message)
      }
    } catch {
      toast.error("Failed to analyze grammar")
    } finally {
      setPendingId(null)
    }
  }

  return {
    pendingId,
    triggerAnalysis,
  }
}
