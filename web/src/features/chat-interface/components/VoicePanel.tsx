import { PanelRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Logo } from "@/components/common/Logo"
import { SessionIndicator } from "./SessionIndicator"
import { SessionTimer } from "./SessionTimer"
import { VoiceOrb } from "./VoiceOrb"
import { WaveformVisualization } from "./WaveformVisualization"
import { VoiceToolbar } from "./VoiceToolbar"

interface VoicePanelProps {
  showHistory: boolean
  onToggleHistory: () => void
}

export function VoicePanel({ showHistory, onToggleHistory }: VoicePanelProps) {
  return (
    <section
      className={`relative flex flex-col items-center justify-center border-r border-border bg-card ${
        showHistory ? "w-[60%]" : "flex-1"
      }`}
    >
      <div className="absolute top-4 left-4 flex items-center gap-2">
        <Logo />
        <span className="text-muted-foreground">/</span>
        <SessionIndicator />
      </div>

      {!showHistory && (
        <div className="absolute top-4 right-4">
          <Button variant="ghost" size="icon-sm" onClick={onToggleHistory}>
            <PanelRight />
          </Button>
        </div>
      )}

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
