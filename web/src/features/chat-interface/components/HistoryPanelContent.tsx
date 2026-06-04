"use client"

import { usePipecatConversation } from "@pipecat-ai/client-react"
import { MessageBubble } from "../../../components/common/MessageBubble"
import { MessageRole } from "@/lib/enums/message-role.enum"
import { AnalysisCard } from "../../../components/common/AnalysisCard"
import { HistoryHeader } from "./HistoryHeader"
import { useEffect, useRef } from "react"

export function HistoryPanelContent() {
  const { messages } = usePipecatConversation()
  const bottomRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    bottomRef.current?.scrollIntoView({
      behavior: "smooth",
    })
  }, [messages])

  return (
    <div className="px-5">
      <HistoryHeader />
      <div className="flex flex-col gap-6">
        {messages.map((message, i) => {
          // function call
          if (message.role === "function_call" && message.functionCall) {
            const args = (message.functionCall.args ?? {})
            return (
              <MessageBubble key={i} role={MessageRole.Analysis}>
                <AnalysisCard
                  suggestions={{
                    hint: "Correction:",
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

          const role = message.role === "user" ? MessageRole.User : MessageRole.Assistant
          const text = message.parts?.map(p => {
            if (typeof p.text === "string") return p.text
            if (
              p.text !== null &&
              typeof p.text === "object" &&
              "spoken" in p.text &&
              "unspoken" in p.text
            ) {
              // BotOutputText
              return (p.text as { spoken: string; unspoken: string }).spoken + 
                    (p.text as { spoken: string; unspoken: string }).unspoken
            }
            return ""
          }).join("")

          return (
            <MessageBubble key={i} role={role}>
              {text}
            </MessageBubble>
          )
        })}
        <div ref={bottomRef} />
      </div>
    </div>
  )
}
