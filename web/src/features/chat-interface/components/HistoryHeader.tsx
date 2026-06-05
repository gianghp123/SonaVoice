"use client"

import { useT } from "next-i18next/client"
import { PanelRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useSidebar } from "@/components/ui/sidebar"

export function HistoryHeader() {
  const { t } = useT('chat')
  const { toggleSidebar } = useSidebar()

  return (
    <div className="flex items-center justify-between p-4 border-border sticky top-0 z-10 bg-background">
      <div className="flex items-center gap-2">
        <h2 className="text-md font-bold leading-relaxed tracking-tight text-primary">
          {t('history')}
        </h2>
      </div>
      <Button variant="ghost" size="icon-sm" onClick={toggleSidebar}>
        <PanelRight />
      </Button>
    </div>
  )
}
