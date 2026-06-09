"use client"

import {
  Message,
  MessageAction,
  MessageActions,
  MessageAvatar,
  MessageContent,
} from "@/components/prompt-kit/message"
import { MessageRole } from "@/lib/enums/message-role.enum"
import { cn } from "@/lib/utils"
import { Slot } from "@radix-ui/react-slot"
import { useT } from "next-i18next/client"
import { is } from "zod/v4/locales"

interface MessageBubbleProps {
  role: MessageRole
  children: React.ReactNode
  actions?: { element: React.ReactNode; tooltip: string }[]
  avatar?: React.ReactNode,
  className?: string
  contentClassName?: string
  timestamp?: Date
  wasInterrupted?: boolean
  asChild?: boolean
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
  actions,
  className,
  contentClassName,
  avatar,
  timestamp,
  wasInterrupted,
  asChild = false,
}: MessageBubbleProps) {
  const { t } = useT('chat')
  const Content = asChild ? Slot : MessageContent
  const isUser = role === MessageRole.User

  return (
    <Message className={cn(className)}>
      {avatar && (
        <MessageAvatar
          fallback={avatar}
          className="h-8 w-8"
        />
      )}

      <div className={cn("flex flex-col gap-1", contentClassName)}>
        <RoleBadge role={role} />

        <Content
          className={cn(isUser ? "bg-primary text-primary-foreground" : "bg-muted")}
        >
          {children}
        </Content>


        {timestamp && (
          <span className={cn("text-[10px] text-muted-foreground", isUser ? "self-end" : "self-start")}>
            {timestamp.toLocaleTimeString()}
          </span>
        )}

        <MessageActions className={cn(isUser ? "self-end" : "self-start")}>
          {
            actions?.map((action) =>
            (
              <MessageAction tooltip={action.tooltip}>
                {action.element}
              </MessageAction>
            ))
          }
        </MessageActions>


        {wasInterrupted && (
          <span className="inline-flex w-fit items-center rounded-md border border-destructive/50 bg-destructive/10 px-1.5 py-0.5 text-[10px] font-medium text-destructive">
            {t('interrupted')}
          </span>
        )}
      </div>
    </Message>
  )
}
