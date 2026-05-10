import { AudioLines } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { MessageRole } from "@/lib/enums/message-role.enum"
import { ChatBubble, ChatBubbleMessage, ChatBubbleAvatar } from "@/components/ui/chat-bubble"

interface MessageBubbleProps {
  role: MessageRole
  children: React.ReactNode
  className?: string
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

export function MessageBubble({ role, children, className }: MessageBubbleProps) {
  if (role === MessageRole.User) {
    return (
      <ChatBubble variant="sent" className={className}>
        <ChatBubbleMessage variant="sent" className="font-medium max-w-[85%]">
          {children}
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
        </ChatBubbleMessage>
      </div>
    </ChatBubble>
  )
}
