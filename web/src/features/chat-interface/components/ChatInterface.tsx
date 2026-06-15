'use client'

import { Button } from "@/components/ui/button"
import {
  Sidebar,
  SidebarContent,
  SidebarInset,
  SidebarProvider,
  useSidebar
} from "@/components/ui/sidebar"
import { HistoryPanelContent } from "@/features/chat-interface/components/HistoryPanelContent"
import { VoicePanel } from "@/features/chat-interface/components/VoicePanel"
import { useMediaQuery } from "@/hooks/use-media-query"
import { PanelRight } from "lucide-react"

function HistoryTrigger() {
  const { toggleSidebar, open } = useSidebar()

  if (open) return null

  return (
    <div className="absolute top-4 right-4 z-10">
      <Button variant="ghost" size="icon-sm" onClick={toggleSidebar}>
        <PanelRight />
      </Button>
    </div>
  )
}

export function ChatInterface({
  maxDuration,
  handleDisconnect,
}: {
  maxDuration: number
  handleDisconnect: () => void | Promise<void>
}) {
  const isMobile = useMediaQuery("(max-width: 767px)")
  const isDesktop = useMediaQuery("(min-width: 1024px)")

  return (
    <SidebarProvider
      defaultOpen={false}
      style={{ "--sidebar-width": isMobile ? "100%" : isDesktop ? "60vh" : "280px" } as React.CSSProperties}
    >
      <SidebarInset>
        <VoicePanel maxDuration={maxDuration} handleDisconnect={handleDisconnect}>
          <HistoryTrigger />
        </VoicePanel>
      </SidebarInset>

      <Sidebar side="right">
        <SidebarContent>
          <HistoryPanelContent />
        </SidebarContent>
      </Sidebar>
    </SidebarProvider>
  )
}
