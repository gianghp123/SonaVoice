
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
import { PanelRight } from "lucide-react"
import { useRouter } from "next/navigation"

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