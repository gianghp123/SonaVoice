"use client"

import {
  ChatContainerContent,
  ChatContainerRoot,
  ChatContainerScrollAnchor,
} from "@/components/prompt-kit/chat-container"
import {
  MessageAction,
  MessageActions,
} from "@/components/prompt-kit/message"
import { ScrollButton } from "@/components/prompt-kit/scroll-button"
import { Button } from "@/components/ui/button"
import { MessageRole } from "@/lib/enums/message-role.enum"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"
import { usePipecatConversation } from "@pipecat-ai/client-react"
import { Loader2, Sparkle } from "lucide-react"
import { useT } from "next-i18next/client"
import { useState } from "react"
import { toast } from "sonner"
import { AnalysisCard } from "../../../components/common/AnalysisCard"
import { MessageBubble } from "../../../components/common/MessageBubble"
import { analyzeGrammar } from "../services/grammar.actions"
import { HistoryHeader } from "./HistoryHeader"
import { FALLBACK_LANGUAGE, isSupportedLanguage, LANGUAGE_FULL_NAMES, SupportedLanguage } from "@/lib/i18n/i18n"

export function HistoryPanelContent() {
  const { messages } = usePipecatConversation()
  const { t, i18n } = useT("chat")
  const [analyses, setAnalyses] = useState<Record<number, IGrammarAnalysis>>({})
  const [pendingIndex, setPendingIndex] = useState<number | null>(null)

  const explainationLanguage = LANGUAGE_FULL_NAMES[isSupportedLanguage(i18n.language) ? i18n.language as SupportedLanguage : FALLBACK_LANGUAGE]

  const handleAnalyzeGrammar = async (index: number, transcript: string) => {
    if (!transcript.trim()) {
      toast.error("Transcript is required")
      return
    }

    setPendingIndex(index)

    try {
      const response = await analyzeGrammar({
        transcript,
        explainationLanguage,
      })

      if (response.error) {
        toast.error(response.error.message)
        return
      }

      if (response.data) {
        setAnalyses((prev) => ({
          ...prev,
          [index]: response.data!,
        }))
      }
    } catch {
      toast.error("Failed to analyze grammar")
    } finally {
      setPendingIndex(null)
    }
  }

  return (
    <div className="flex flex-col h-full md:pb-10">
      <HistoryHeader />

      <ChatContainerRoot className="flex-1">
        <ChatContainerContent className="flex flex-col gap-6">
          {messages.map((message, i) => {
            if (message.role === "function_call" && message.functionCall) {
              return (
                <MessageBubble key={i} role={MessageRole.Analysis}>
                  <AnalysisCard
                    suggestions={{
                      hint: t("correction"),
                      original: "",
                      corrected: "",
                    }}
                    pronunciation={{
                      word: "",
                      phonetic: "",
                    }}
                  />
                </MessageBubble>
              )
            }

            const role =
              message.role === "user"
                ? MessageRole.User
                : MessageRole.Assistant

            const text =
              message.parts
                ?.map((p) => {
                  if (typeof p.text === "string") return p.text

                  if (
                    p.text !== null &&
                    typeof p.text === "object" &&
                    "spoken" in p.text &&
                    "unspoken" in p.text
                  ) {
                    const value = p.text as {
                      spoken: string
                      unspoken: string
                    }

                    return value.spoken + value.unspoken
                  }

                  return ""
                })
                .join("") ?? ""

            const isCurrentLoading = pendingIndex === i
            const analysis = analyses[i]

            return (
              <div key={i} className="flex flex-col">
                <MessageBubble
                  role={role}
                  action={
                    role === MessageRole.User && (
                      <MessageActions className="self-end!">
                        <MessageAction tooltip={t("analyze_grammar")}>
                          <Button
                            variant="ghost"
                            size="icon"
                            type="button"
                            disabled={pendingIndex !== null}
                            onClick={() => handleAnalyzeGrammar(i, text)}
                            className="h-6 w-6 text-muted-foreground hover:text-foreground"
                          >
                            {isCurrentLoading ? (
                              <Loader2 className="h-3 w-3 animate-spin" />
                            ) : (
                              <Sparkle className="h-3 w-3 text-purple-500" />
                            )}
                          </Button>
                        </MessageAction>
                      </MessageActions>
                    )
                  }
                >
                  {text}
                </MessageBubble>

                {analysis && (
                  <MessageBubble role={MessageRole.Analysis}>
                    <AnalysisCard grammar={analysis} />
                  </MessageBubble>
                )}
              </div>
            )
          })}

          <ChatContainerScrollAnchor />
        </ChatContainerContent>

        <div className="absolute right-12 bottom-4">
          <ScrollButton />
        </div>
      </ChatContainerRoot>
    </div>
  )
}