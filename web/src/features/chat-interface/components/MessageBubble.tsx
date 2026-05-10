import { AudioLines } from "lucide-react"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { cn } from "@/lib/utils"

type MessageRole = "sona" | "user" | "analysis"

interface MessageBubbleProps {
  role: MessageRole
  children: React.ReactNode
  className?: string
}

function SonaAvatar() {
  return (
    <Avatar className="size-7 bg-primary flex-shrink-0">
      <AvatarFallback className="bg-primary text-primary-foreground">
        <AudioLines />
      </AvatarFallback>
    </Avatar>
  )
}

function RoleBadge({ role }: { role: MessageRole }) {
  if (role === "user") return null
  return (
    <Badge
      variant="secondary"
      className="text-[10px] font-bold uppercase tracking-widest bg-transparent px-0 hover:bg-transparent"
    >
      {role === "analysis" ? "ANALYSIS" : "SONA"}
    </Badge>
  )
}

export function MessageBubble({ role, children, className }: MessageBubbleProps) {
  if (role === "user") {
    return (
      <div className={cn("flex flex-col items-end", className)}>
        <Card className="max-w-[85%] border-secondary bg-secondary text-primary shadow-sm">
          <CardContent className="font-medium">{children}</CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className={cn("flex gap-4", className)}>
      <SonaAvatar />
      <div className="flex flex-col gap-1 flex-1">
        <RoleBadge role={role} />
        <Card className="border-secondary/40 shadow-sm">
          <CardContent>{children}</CardContent>
        </Card>
      </div>
    </div>
  )
}
