"use client"

import { Badge } from "@/components/ui/badge"

interface SessionIndicatorProps {
  transportType?: string | null
}

export function SessionIndicator({ transportType }: SessionIndicatorProps) {
  return (
    <Badge variant="outline" className="font-semibold tracking-wider">
      {transportType ? `Session via ${transportType}` : "Session"}
    </Badge>
  )
}
