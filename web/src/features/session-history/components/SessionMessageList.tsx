"use client"

import { GrammarAnalysisCard } from "@/features/shared/grammar/components/GrammarAnalysisCard"
import { BotAvatarIcon } from "@/components/common/BotAvatarIcon"
import { GrammarAnalysisButton } from "@/features/shared/grammar/components/GrammarAnalysisButton"
import { MessageBubble } from "@/components/common/MessageBubble"
import type { SessionItem } from "@/features/session-history/types"
import { useGrammarAnalysis } from "@/features/shared/grammar/hooks/useGrammarAnalysis"
import { MessageRole } from "@/lib/enums/message-role.enum"
import { useT } from "next-i18next/client"

interface SessionMessageListProps {
  items: SessionItem[]
}

export function SessionMessageList({ items }: SessionMessageListProps) {
  const { t } = useT("session")
  const { t: tChat } = useT("chat")
  const { pendingId, triggerAnalysis } = useGrammarAnalysis()

  if (items.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <p className="text-sm text-muted-foreground">{t("no_messages")}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6 px-5 w-full max-w-7/12">
      {items.map((item, index) => {
        if (item.type === "analysis") {
          return (
            <MessageBubble
              key={`analysis-${item.data.messageId ?? index}`}
              role={MessageRole.Analysis}
              avatar={<BotAvatarIcon />}
              asChild
              contentClassName="w-full"
            >
              <GrammarAnalysisCard grammar={item.data} />
            </MessageBubble>
          )
        }

        const msg = item.data
        const isUser = msg.role === MessageRole.User

        return (
          <MessageBubble
            key={msg.id}
            role={msg.role}
            avatar={!isUser ? <BotAvatarIcon /> : undefined}
            timestamp={new Date(msg.createdAt)}
            wasInterrupted={msg.wasInterrupted}
            className={isUser ? "justify-end" : "justify-start"}
            actions={
              isUser ? [
                {
                  tooltip: tChat("analyze_grammar"),
                  element: (
                    <GrammarAnalysisButton
                      tooltip={tChat("analyze_grammar")}
                      disabled={pendingId !== null}
                      isLoading={pendingId === msg.id}
                      onClick={() => triggerAnalysis(msg.id)}
                    />
                  )
                }
              ] : undefined
            }
          >
            {msg.transcript}
          </MessageBubble>
        )
      })}
    </div>
  )
}
