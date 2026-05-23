"use client"

import { Logo } from "@/components/common/Logo"
import { SessionIndicator } from "./SessionIndicator"
import { SessionTimer } from "./SessionTimer"
import { VoiceOrb } from "./VoiceOrb"
import { WaveformVisualization } from "./WaveformVisualization"
import { VoiceToolbar } from "./VoiceToolbar"
import { usePipecatClientTransportState } from "@pipecat-ai/client-react"

export function VoicePanel({ maxDuration, children, handleDisconnect }: {maxDuration: number, children?: React.ReactNode, handleDisconnect: () => void | Promise<void> }) {
  const transportState = usePipecatClientTransportState()

  return (
    <section className="relative flex flex-1 flex-col items-center justify-center">
      <div className="absolute top-4 left-4 flex items-center gap-2">
        <Logo />
        <span className="text-muted-foreground">/</span>
        <SessionIndicator transportType={transportState === "ready" ? "Active" : null} />
      </div>

      {children}

      <div className="absolute top-4 left-1/2 -translate-x-1/2">
        <SessionTimer maxDuration={maxDuration}/>
      </div>

      <div className="relative flex flex-col items-center">
        <VoiceOrb />
        <WaveformVisualization />
      </div>

      <div className="absolute bottom-4">
        <VoiceToolbar handleDisconnect={handleDisconnect} />
      </div>
    </section>
  )
}
