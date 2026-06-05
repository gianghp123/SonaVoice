"use client"

import { Badge } from "@/components/ui/badge"
import { useT } from "next-i18next/client"

interface SessionIndicatorProps {
  transportType?: string | null
}

export function SessionIndicator({ transportType }: SessionIndicatorProps) {
  const { t } = useT('chat')
  return (
    <Badge variant="outline" className="font-semibold tracking-wider">
      {transportType ? t('session_via', { transportType }) : t('session')}
    </Badge>
  )
}
