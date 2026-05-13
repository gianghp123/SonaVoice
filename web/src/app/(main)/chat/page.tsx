"use client"

import { Button } from "@/components/ui/button"
import {
  Sidebar,
  SidebarContent,
  SidebarInset,
  SidebarProvider,
  useSidebar,
} from "@/components/ui/sidebar"
import { HistoryPanelContent } from "@/features/chat-interface/components/HistoryPanelContent"
import { VoicePanel } from "@/features/chat-interface/components/VoicePanel"
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit"
import { PanelRight } from "lucide-react"

const PIPECAT_ENDPOINT =
  process.env.NEXT_PUBLIC_PIPECAT_ENDPOINT ?? "http://localhost:7860/api/offer"

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

function ChatLayout({ handleDisconnect }: { handleDisconnect: () => void | Promise<void> }) {
  return (
    <SidebarProvider
      defaultOpen={true}
      style={{ "--sidebar-width": "40vh" } as React.CSSProperties}
    >
      <SidebarInset className="bg-card">
        <VoicePanel handleDisconnect={handleDisconnect}>
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

export default function ChatPage() {
  return (
    <PipecatAppBase
      transportType="smallwebrtc"
      connectParams={{ webrtcUrl: PIPECAT_ENDPOINT }}
      noThemeProvider
      clientOptions={{ enableMic: true }}
      initDevicesOnMount={true}
      connectOnMount={true}
    >
      {({ client, error, handleConnect, handleDisconnect }) => {
        if (error) return <div>Error: {error}</div>
        if (!client) return <div>Initializing...</div>
        return <ChatLayout
          handleDisconnect={handleDisconnect ?? (() => {})}
        />
      }}
    </PipecatAppBase>
  )
}