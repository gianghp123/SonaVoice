"use client"

import { Badge } from "@/components/ui/badge"
import { ChatBubble, ChatBubbleAvatar, ChatBubbleMessage } from "@/components/ui/chat-bubble"
import { MessageRole } from "@/lib/enums/message-role.enum"
import { useT } from "next-i18next/client"
import { AudioLines } from "lucide-react"

interface MessageBubbleProps {
  role: MessageRole
  children: React.ReactNode
  className?: string
  timestamp?: Date
  wasInterrupted?: boolean
}

function RoleBadge({ role, t }: { role: MessageRole; t: (key: string) => string }) {
  if (role === MessageRole.User) return null
  return (
    <Badge
      variant="secondary"
      className="text-[10px] font-bold uppercase tracking-widest bg-transparent px-0 hover:bg-transparent"
    >
      {role === MessageRole.Analysis ? t('analysis') : t('sona')}
    </Badge>
  )
}

function MessageFooter({ timestamp, t, wasInterrupted }: { timestamp?: Date; t: (key: string) => string; wasInterrupted?: boolean }) {
  if (!timestamp && !wasInterrupted) return null

  return (
    <div className="flex items-center gap-2 mt-1">
      {timestamp && (
        <span className="text-[10px] text-muted-foreground">
          {timestamp.toLocaleString()}
        </span>
      )}
      {wasInterrupted && (
        <Badge variant="destructive" className="text-[10px] px-1 py-0">
          {t('interrupted')}
        </Badge>
      )}
    </div>
  )
}

export function MessageBubble({ role, children, className, timestamp, wasInterrupted }: MessageBubbleProps) {
  const { t } = useT('chat')

  if (role === MessageRole.User) {
    return (
      <ChatBubble variant="sent" className={className}>
        <ChatBubbleMessage variant="sent" className="font-medium max-w-[85%]">
          {children}
          <MessageFooter timestamp={timestamp} wasInterrupted={wasInterrupted} t={t} />
        </ChatBubbleMessage>
      </ChatBubble>
    )
  }

  return (
    <ChatBubble variant="received" className={className}>
      <ChatBubbleAvatar className="bg-primary text-primary-foreground" fallback={<AudioLines />} />
      <div className="flex flex-col gap-1 flex-1">
        <RoleBadge role={role} t={t} />
        <ChatBubbleMessage variant="received" className="bg-card border-[0.5px]">
          {children}
          <MessageFooter timestamp={timestamp} wasInterrupted={wasInterrupted} t={t} />
        </ChatBubbleMessage>
      </div>
    </ChatBubble>
  )
}
