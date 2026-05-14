import { SidebarFooterUI } from "@/components/common/SidebarFooter"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarInset,
  SidebarProvider,
  useSidebar,
} from "@/components/ui/sidebar"
import { HistoryPanelContent } from "@/features/chat-interface/components/HistoryPanelContent"
import { VoicePanel } from "@/features/chat-interface/components/VoicePanel"
import { Show, SignInButton, UserAvatar } from "@clerk/nextjs"
import { RTVIEvent } from "@pipecat-ai/client-js"
import { useRTVIClientEvent } from "@pipecat-ai/client-react"
import { PanelRight } from "lucide-react"
import { useRouter } from "next/navigation"
import { useCallback, useEffect, useState } from "react"
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

export function ChatLayout({
  handleDisconnect,
  initialError
}: {
  handleDisconnect: () => void | Promise<void>
  initialError?: string | null
}) {
  const [fatalError, setFatalError] = useState<string | null>(null)
  const router = useRouter()

  useRTVIClientEvent(
    RTVIEvent.Error,
    useCallback((message) => {
      const { error, message: msg, fatal } = message.data as any
      const text = error ?? msg  // use "error" first, fall back to "message"

      toast.error("An error occurred: " + text, {
        duration: 10000,
      })

      if (fatal) {
        setFatalError(text)
      }
    }, [])
  )

  const redirectOnDisconnect = async () => {
    await handleDisconnect()
    router.push("/")
  }

  useEffect(() => {
    if (initialError) {
      setFatalError(initialError)
    }
  }, [initialError])

  return (
    <SidebarProvider
      defaultOpen={false}
      style={{ "--sidebar-width": "60vh" } as React.CSSProperties}
    >
      <SidebarInset>
        <VoicePanel handleDisconnect={redirectOnDisconnect}>
          <HistoryTrigger />

          <AlertDialog
            open={!!fatalError}
            onOpenChange={(open) => {
              if (!open && fatalError) return
            }}
          >
            <AlertDialogContent
              onEscapeKeyDown={(event) => event.preventDefault()}
            >
              <AlertDialogHeader>
                <AlertDialogTitle>Session ended</AlertDialogTitle>
                <AlertDialogDescription>
                  {fatalError || "The bot runtime stopped unexpectedly."}
                </AlertDialogDescription>
              </AlertDialogHeader>

              <AlertDialogFooter>
                <AlertDialogAction onClick={redirectOnDisconnect}>
                  Return home
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </VoicePanel>
      </SidebarInset>

      <Sidebar side="right">
        <SidebarContent>
          <HistoryPanelContent />
        </SidebarContent>
        <SidebarFooter className="border-t-[0.5px]">
          <SidebarFooterUI />
        </SidebarFooter>
      </Sidebar>
    </SidebarProvider>
  )
}