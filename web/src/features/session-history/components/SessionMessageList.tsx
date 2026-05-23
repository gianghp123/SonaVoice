"use client"

import { MessageBubble } from "@/components/common/MessageBubble"
import type { IMessage } from "@/lib/types/message.interface"

interface SessionMessageListProps {
  messages: IMessage[]
}

export function SessionMessageList({ messages }: SessionMessageListProps) {
  if (messages.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <p className="text-sm text-muted-foreground">No messages in this session</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6 px-5 w-full max-w-7/12">
      {[...messages]
        .sort(
          (a, b) =>
            new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        ).map((message) => (
          <MessageBubble
            key={message.id}
            role={message.role}
            timestamp={new Date(message.createdAt)}
            wasInterrupted={message.wasInterrupted}
          >
            {message.transcript}
          </MessageBubble>
        ))}
    </div>
  )
}
