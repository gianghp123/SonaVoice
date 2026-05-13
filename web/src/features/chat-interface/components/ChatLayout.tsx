
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
import { RTVIEvent } from "@pipecat-ai/client-js"
import { useRTVIClientEvent } from "@pipecat-ai/client-react"
import { PanelRight } from "lucide-react"
import { useRouter } from "next/navigation"
import { useCallback } from "react"
import { toast } from "sonner"

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

export function ChatLayout({ handleDisconnect }: { handleDisconnect: () => void | Promise<void> }) {
  const router = useRouter()
  useRTVIClientEvent(
    RTVIEvent.Error,
    useCallback((message) => {
      const { message: text, fatal } = message.data as any;
      console.error("Bot runtime error:", text);
      toast.error("An error occurred in the bot runtime. Please try again.");
      if (fatal) {
        // Bot has disconnected — show reconnect UI
      }
    }, [])
  );


  const redirectOnDisconnect = () => {
    handleDisconnect()
    router.push("/")
  }

  return (
    <SidebarProvider
      defaultOpen={true}
      style={{ "--sidebar-width": "60vh" } as React.CSSProperties}
    >
      <SidebarInset className="bg-card">
        <VoicePanel handleDisconnect={redirectOnDisconnect}>
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