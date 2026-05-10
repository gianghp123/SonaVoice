import { Badge } from "@/components/ui/badge"

export function SessionIndicator({
  sessionNumber = 4,
}: {
  sessionNumber?: number
}) {
  return (
    <Badge variant="outline" className="font-semibold tracking-wider">
      Session {String(sessionNumber).padStart(2, "0")}
    </Badge>
  )
}
