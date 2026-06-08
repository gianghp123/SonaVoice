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
import { usePipecatConversation } from "@pipecat-ai/client-react"
import { Loader2, Sparkle } from "lucide-react"
import { useActionState, useEffect } from "react"
import { toast } from "sonner"
import { AnalysisCard } from "../../../components/common/AnalysisCard"
import { MessageBubble } from "../../../components/common/MessageBubble"
import { analyzeGrammarAction } from "../services/grammar.actions"
import { HistoryHeader } from "./HistoryHeader"
import { useT } from "next-i18next/client"

export function HistoryPanelContent() {
  const { messages } = usePipecatConversation()
  const { t } = useT('chat')
  const [grammarState, dispatchGrammar, isPending] = useActionState(
    analyzeGrammarAction,
    { result: null, error: null, index: -1 }
  )

  useEffect(() => {
    if (grammarState.error) {
      toast.error(grammarState.error)
    }
  }, [grammarState.error])

  return (
    <div className="flex flex-col h-full px-5">
      <HistoryHeader />
      <ChatContainerRoot className="flex-1">
        <ChatContainerContent className="flex flex-col gap-6">
          {messages.map((message, i) => {
            // function call
            if (message.role === "function_call" && message.functionCall) {
              return (
                <MessageBubble key={i} role={MessageRole.Analysis}>
                  <AnalysisCard
                    suggestions={{
                      hint: t('correction'),
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
                return (p.text as { spoken: string; unspoken: string }).spoken +
                  (p.text as { spoken: string; unspoken: string }).unspoken
              }
              return ""
            }).join("")

            return (
              <div key={i} className="flex flex-col">
                <MessageBubble role={role} action={role === MessageRole.User && (
                  <MessageActions className="self-end!">
                    <form action={dispatchGrammar}>
                      <input type="hidden" name="transcript" value={text} />
                      <input type="hidden" name="index" value={String(i)} />
                      <MessageAction tooltip={t('analyze_grammar')}>
                        <Button
                          variant="ghost"
                          size="icon"
                          type="submit"
                          disabled={isPending}
                          className="h-6 w-6 text-muted-foreground hover:text-foreground"
                        >
                          {isPending && grammarState.index === i ? (
                            <Loader2 className="h-3 w-3 animate-spin" />
                          ) : (
                            <Sparkle className="h-3 w-3 text-purple-500"/>
                          )}
                        </Button>
                      </MessageAction>
                    </form>
                  </MessageActions>
                )}>
                  {text}

                </MessageBubble>
                {grammarState.result && grammarState.index === i && (
                  <MessageBubble role={MessageRole.Analysis}>
                    <AnalysisCard grammar={grammarState.result} />
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
      {/* <div className="flex justify-center py-2">
        <ScrollButton />
      </div> */}
    </div>
  )
}
