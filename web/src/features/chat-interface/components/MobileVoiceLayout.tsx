"use client"

import { Button } from "@/components/ui/button"
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer"
import { ScrollArea } from "@/components/ui/scroll-area"
import { HistoryPanelContent } from "@/features/chat-interface/components/HistoryPanelContent"
import { VoicePanel } from "@/features/chat-interface/components/VoicePanel"
import { ArrowLeft, History } from "lucide-react"
import { useT } from "next-i18next/client"
import Link from "next/link"
import { useState } from "react"

interface MobileVoiceLayoutProps {
  maxDuration: number
  handleDisconnect: () => void | Promise<void>
}

export function MobileVoiceLayout({ maxDuration, handleDisconnect }: MobileVoiceLayoutProps) {
  const [historyOpen, setHistoryOpen] = useState(false)
  const { t } = useT("chat")

  return (
    <div className="flex h-screen flex-col">
      {/* Minimal header */}
      <header className="flex h-14 shrink-0 items-center justify-between border-b px-4">
        <Link href="/" className="flex items-center gap-2 text-muted-foreground">
          <ArrowLeft className="size-4" />
          <span className="text-sm">{t("back")}</span>
        </Link>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setHistoryOpen(true)}
        >
          <History className="size-5" />
        </Button>
      </header>

      {/* Voice area */}
      <div className="flex flex-1 flex-col items-center justify-center gap-4 px-4">
        <VoicePanel maxDuration={maxDuration} handleDisconnect={handleDisconnect} />
      </div>

      {/* History Drawer */}
      <Drawer direction="bottom" open={historyOpen} onOpenChange={setHistoryOpen}>
        <DrawerContent>
          <DrawerHeader className="text-left">
            <DrawerTitle>{t("conversation_history")}</DrawerTitle>
            <DrawerDescription>{t("session_messages")}</DrawerDescription>
          </DrawerHeader>
          <ScrollArea className="h-[60vh] px-4">
            <HistoryPanelContent />
          </ScrollArea>
        </DrawerContent>
      </Drawer>
    </div>
  )
}
