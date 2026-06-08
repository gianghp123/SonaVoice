"use client"

import {
  Message,
  MessageAvatar,
  MessageContent,
} from "@/components/prompt-kit/message"
import { MessageRole } from "@/lib/enums/message-role.enum"
import { AudioLines } from "lucide-react"
import { useT } from "next-i18next/client"

interface MessageBubbleProps {
  role: MessageRole
  children: React.ReactNode
  action?: React.ReactNode
  className?: string
  timestamp?: Date
  wasInterrupted?: boolean
}

function RoleBadge({ role }: { role: MessageRole }) {
  const { t } = useT('chat')
  if (role === MessageRole.User) return null
  const label = role === MessageRole.Analysis ? t('analysis') : t('sona')
  return (
    <span className="font-mono text-[9px] font-bold uppercase tracking-wider text-muted-foreground">
      {label}
    </span>
  )
}

export function MessageBubble({
  role,
  children,
  action,
  timestamp,
  wasInterrupted,
}: MessageBubbleProps) {
  const isUser = role === MessageRole.User

  return (
    <Message className={isUser ? "justify-end" : "justify-start"}>
      {!isUser && (
        <MessageAvatar
          fallback={<AudioLines className="h-4 w-4" />}
          className="h-8 w-8"
        />
      )}

      <div className="flex flex-col gap-1">
        <RoleBadge role={role} />

        <MessageContent
          className={
            isUser
              ? "bg-primary text-primary-foreground"
              : "bg-muted"
          }
        >
          {children}
        </MessageContent>

        
        {action}

        {timestamp && (
          <span className="text-[10px] text-muted-foreground">
            {timestamp.toLocaleTimeString()}
          </span>
        )}

        {wasInterrupted && (
          <span className="inline-flex w-fit items-center rounded-md border border-destructive/50 bg-destructive/10 px-1.5 py-0.5 text-[10px] font-medium text-destructive">
            {t('interrupted')}
          </span>
        )}
      </div>
    </Message>
  )
}
