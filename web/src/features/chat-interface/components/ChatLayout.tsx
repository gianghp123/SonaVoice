'use client'

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
import { RTVIEvent, RTVIMessage } from "@pipecat-ai/client-js"
import { useRTVIClientEvent } from "@pipecat-ai/client-react"
import * as Sentry from "@sentry/nextjs"
import { PanelRight } from "lucide-react"
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

type RTVIErrorData = {
  error?: string
  message?: string
  fatal?: boolean
}

export function ChatLayout({
  maxDuration,
  handleDisconnect,
  handleError,
  initialError
}: {
  maxDuration: number
  handleDisconnect: () => void | Promise<void>
  handleError: () => void | Promise<void>
  initialError?: string | null
}) {
  const [fatalError, setFatalError] = useState<string | null>(
    initialError ?? null
  )

  useRTVIClientEvent(
    RTVIEvent.Error,
    useCallback((message: RTVIMessage) => {
      const { error, message: msg, fatal } = message.data as RTVIErrorData
      const text = error ?? msg ?? "Unknown RTVI error"

      Sentry.logger[fatal ? "error" : "warn"]("RTVI client error", {
        area: "chat-layout",
        stage: "rtvi",
        fatal: Boolean(fatal),
        error: text,
      })

      if (fatal) {
        Sentry.captureException(new Error(text), {
          tags: {
            area: "chat-layout",
            type: "rtvi-fatal-error",
          },
          extra: {
            rtviEventData: message.data,
          },
        })
      }

      toast.error("An error occurred: " + text, {
        duration: 10000,
      })

      if (fatal) {
        setFatalError(text)
      }
    }, [])
  )

  useEffect(() => {
    if (initialError) {
      Sentry.captureException(new Error(initialError), {
        tags: {
          area: "chat-layout",
          type: "pipecat-initial-error",
        },
      })
    }
  }, [initialError])

  return (
    <SidebarProvider
      defaultOpen={false}
      style={{ "--sidebar-width": "60vh" } as React.CSSProperties}
    >
      <SidebarInset>
        <VoicePanel maxDuration={maxDuration} handleDisconnect={handleDisconnect}>
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
                <AlertDialogAction onClick={handleError}>
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