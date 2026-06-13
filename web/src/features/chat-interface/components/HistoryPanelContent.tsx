"use client"

import { BotAvatarIcon } from "@/components/common/BotAvatarIcon"
import { MessageBubble } from "@/components/common/MessageBubble"
import {
  ChatContainerContent,
  ChatContainerRoot,
  ChatContainerScrollAnchor,
} from "@/components/prompt-kit/chat-container"
import { ScrollButton } from "@/components/prompt-kit/scroll-button"
import { GrammarAnalysisButton } from "@/features/shared/grammar/components/GrammarAnalysisButton"
import { GrammarAnalysisCard } from "@/features/shared/grammar/components/GrammarAnalysisCard"
import { ConversationTurnDto } from "@/features/shared/grammar/dtos/analyze-grammar.req"
import { analyzeGrammar } from "@/features/shared/grammar/services/grammar.actions"
import { MessageRole } from "@/lib/enums/message-role.enum"
import { FALLBACK_LANGUAGE, isSupportedLanguage, LANGUAGE_FULL_NAMES, SupportedLanguage } from "@/lib/i18n/i18n"
import type { IGrammarAIResult } from "@/lib/types/grammar-analysis.interface"
import { ConversationMessage, ConversationMessagePart, usePipecatConversation } from "@pipecat-ai/client-react"
import { useT } from "next-i18next/client"
import { useState } from "react"
import { toast } from "sonner"
import { HistoryHeader } from "./HistoryHeader"

function getMessageText(message: ConversationMessage) {
  return (
    message.parts
      ?.map((p: ConversationMessagePart) => {
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

          return [value.spoken, value.unspoken]
            .filter(Boolean)
            .join(" ")
        }

        return ""
      })
      .filter(Boolean)
      .join(" ")
      .replace(/\s+/g, " ")
      .trim() ?? ""
  )
}

function getConversationContext(messages: ConversationMessage[], currentIndex: number): ConversationTurnDto[] {
  return messages
    .slice(Math.max(0, currentIndex - 5), currentIndex)
    .map((message) => ({
      role: message.role === "user" ? MessageRole.User : MessageRole.Assistant,
      text: getMessageText(message),
    }))
    .filter((message) => message.text.trim().length > 0)
}


export function HistoryPanelContent() {
  const { messages } = usePipecatConversation()
  const { t, i18n } = useT("chat")
  const [analyses, setAnalyses] = useState<Record<number, IGrammarAIResult>>({})
  const [pendingIndex, setPendingIndex] = useState<number | null>(null)

  const explanationLanguage = LANGUAGE_FULL_NAMES[isSupportedLanguage(i18n.language) ? i18n.language as SupportedLanguage : FALLBACK_LANGUAGE]

  const handleAnalyzeGrammar = async (index: number, transcript: string) => {
    if (!transcript.trim()) {
      toast.error("Transcript is required")
      return
    }

    const conversationContext = getConversationContext(messages, index)

    setPendingIndex(index)

    try {
      const response = await analyzeGrammar({
        transcript,
        explanationLanguage,
        conversationContext
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
        <ChatContainerContent className="flex flex-col gap-6 px-3">
          {messages.map((message, i) => {
            const role =
              message.role === "user"
                ? MessageRole.User
                : MessageRole.Assistant

            const text = getMessageText(message)
            const isCurrentLoading = pendingIndex === i
            const analysis = analyses[i]
            const isUser = role === MessageRole.User
            return (
              <div key={i} className="flex flex-col">
                <MessageBubble
                  role={role}
                  avatar={!isUser ? <BotAvatarIcon /> : undefined}
                  actions={
                    isUser ? [
                      {
                        tooltip: t("analyze_grammar"),
                        element: (
                          <GrammarAnalysisButton
                            tooltip={t("analyze_grammar")}
                            disabled={pendingIndex !== null}
                            isLoading={isCurrentLoading}
                            onClick={() => handleAnalyzeGrammar(i, text)}
                          />
                        )
                      }] : undefined
                  }
                >
                  {text}
                </MessageBubble>

                {analysis && (
                  <MessageBubble
                    role={MessageRole.Analysis}
                    asChild
                    contentClassName="w-full"
                    avatar={<BotAvatarIcon />}
                  >
                    <GrammarAnalysisCard grammar={analysis} originalText={text} />
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