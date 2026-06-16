"use client"

import { Logo } from "@/components/common/Logo"
import { useMediaQuery } from "@/hooks/use-media-query"
import { SessionTimer } from "./SessionTimer"
import { VoiceOrb } from "./VoiceOrb"
import { VoiceToolbar } from "./VoiceToolbar"
import { WaveformVisualization } from "./WaveformVisualization"

export function VoicePanel({
  maxDuration,
  children,
  handleDisconnect,
}: {
  maxDuration: number
  children?: React.ReactNode
  handleDisconnect: () => void | Promise<void>
}) {

  const isMobile = useMediaQuery("(max-width: 767px)")
  return (
    <section className="flex h-full flex-col">
      {
        !isMobile ? (
          <header className="relative h-16">
            <div className="absolute left-4 top-1/2 -translate-y-1/2">
              <Logo />
            </div>

            <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2">
              <SessionTimer maxDuration={maxDuration} />
            </div>

            {children}
          </header>
        ) : (
          <header className="flex flex-col gap-2 pt-4">
            <div className="flex items-center justify-between">
              <Logo />
              {children}
            </div>
            <div className="flex justify-center">
              <SessionTimer maxDuration={maxDuration} />
            </div>
          </header>
        )
      }

      <main className="relative flex flex-1 flex-col items-center justify-center">
        <VoiceOrb />
        <WaveformVisualization />
      </main>

      <footer className="flex items-center justify-center">
        <VoiceToolbar handleDisconnect={handleDisconnect} />
      </footer>
    </section>
  )
}
