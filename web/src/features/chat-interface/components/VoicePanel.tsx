import { Logo } from "@/components/common/Logo"
import { SessionIndicator } from "./SessionIndicator"
import { SessionTimer } from "./SessionTimer"
import { VoiceOrb } from "./VoiceOrb"
import { WaveformVisualization } from "./WaveformVisualization"
import { VoiceToolbar } from "./VoiceToolbar"

export function VoicePanel({ children }: { children?: React.ReactNode }) {
  return (
    <section className="relative flex flex-1 flex-col items-center justify-center bg-card">
      <div className="absolute top-4 left-4 flex items-center gap-2">
        <Logo />
        <span className="text-muted-foreground">/</span>
        <SessionIndicator />
      </div>

      {children}

      <div className="absolute top-4 left-1/2 -translate-x-1/2">
        <SessionTimer />
      </div>

      <div className="relative flex flex-col items-center">
        <VoiceOrb />
        <WaveformVisualization />
      </div>

      <div className="absolute bottom-4">
        <VoiceToolbar />
      </div>
    </section>
  )
}
