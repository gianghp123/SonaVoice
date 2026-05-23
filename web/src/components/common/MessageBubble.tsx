import { Badge } from "@/components/ui/badge"
import { ChatBubble, ChatBubbleAvatar, ChatBubbleMessage } from "@/components/ui/chat-bubble"
import { MessageRole } from "@/lib/enums/message-role.enum"
import { AudioLines } from "lucide-react"

interface MessageBubbleProps {
  role: MessageRole
  children: React.ReactNode
  className?: string
  timestamp?: Date
  wasInterrupted?: boolean
}

function RoleBadge({ role }: { role: MessageRole }) {
  if (role === MessageRole.User) return null
  return (
    <Badge
      variant="secondary"
      className="text-[10px] font-bold uppercase tracking-widest bg-transparent px-0 hover:bg-transparent"
    >
      {role === MessageRole.Analysis ? "ANALYSIS" : "SONA"}
    </Badge>
  )
}

function MessageFooter({ timestamp, wasInterrupted }: { timestamp?: Date, wasInterrupted?: boolean }) {
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
          Interrupted
        </Badge>
      )}
    </div>
  )
}

export function MessageBubble({ role, children, className, timestamp, wasInterrupted }: MessageBubbleProps) {
  if (role === MessageRole.User) {
    return (
      <ChatBubble variant="sent" className={className}>
        <ChatBubbleMessage variant="sent" className="font-medium max-w-[85%]">
          {children}
          <MessageFooter timestamp={timestamp} wasInterrupted={wasInterrupted} />
        </ChatBubbleMessage>
      </ChatBubble>
    )
  }

  return (
    <ChatBubble variant="received" layout="ai" className={className}>
      <ChatBubbleAvatar className="bg-primary text-primary-foreground" fallback={<AudioLines />} />
      <div className="flex flex-col gap-1 flex-1">
        <RoleBadge role={role} />
        <ChatBubbleMessage variant="received" className="bg-card border-[0.5px]">
          {children}
          <MessageFooter timestamp={timestamp} wasInterrupted={wasInterrupted} />
        </ChatBubbleMessage>
      </div>
    </ChatBubble>
  )
}
