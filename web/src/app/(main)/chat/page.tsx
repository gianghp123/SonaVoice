"use client"

import { useState } from "react"
import { VoicePanel } from "@/features/chat-interface/components/VoicePanel"
import { HistoryPanel } from "@/features/chat-interface/components/HistoryPanel"
import { BottomNavBar } from "@/features/chat-interface/components/BottomNavBar"

export default function ChatPage() {
  const [showHistory, setShowHistory] = useState(false)

  return (
    <>
      <VoicePanel
        showHistory={showHistory}
        onToggleHistory={() => setShowHistory((v) => !v)}
      />
      {showHistory && (
        <HistoryPanel onClose={() => setShowHistory(false)} />
      )}
      <BottomNavBar />
    </>
  )
}
