"use client"

import { PanelRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Sidebar,
  SidebarContent,
  SidebarInset,
  SidebarProvider,
  useSidebar,
} from "@/components/ui/sidebar"
import { VoicePanel } from "@/features/chat-interface/components/VoicePanel"
import { HistoryPanelContent } from "@/features/chat-interface/components/HistoryPanelContent"

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

export default function ChatPage() {
  return (
    <SidebarProvider
      defaultOpen={false}
      style={
        {
          "--sidebar-width": "40vw",
        } as React.CSSProperties
      }
    >
      {/* 1. Main content goes FIRST */}
      <SidebarInset className="bg-card">
        <VoicePanel>
          <HistoryTrigger />
        </VoicePanel>
      </SidebarInset>
      
      {/* 2. Right sidebar goes SECOND */}
      <Sidebar side="right">
        <SidebarContent>
          <HistoryPanelContent />
        </SidebarContent>
      </Sidebar>
    </SidebarProvider>
  )
}