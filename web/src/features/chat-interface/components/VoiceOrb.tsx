"use client"

import { usePipecatClientMediaTrack } from "@pipecat-ai/client-react"
import { usePipecatClientTransportState } from "@pipecat-ai/client-react"
import { CircularWaveform } from "@pipecat-ai/voice-ui-kit"

export function VoiceOrb() {
  const botAudioTrack = usePipecatClientMediaTrack("audio", "bot")
  const transportState = usePipecatClientTransportState()
  const isThinking = transportState === "connecting" || transportState === "authenticating"

  return (
    <div className="relative flex items-center justify-center mb-8">
      <div
        className="absolute size-[360px] rounded-full border border-primary animate-orb-ring"
        style={{ animationDelay: "1.6s", opacity: 0 }}
      />
      <div
        className="absolute size-[300px] rounded-full border border-primary animate-orb-ring"
        style={{ animationDelay: "0.8s", opacity: 0 }}
      />
      <div
        className="absolute size-[240px] rounded-full border border-primary animate-orb-ring"
        style={{ animationDelay: "0s", opacity: 0 }}
      />
      <div className="relative z-10 flex size-48 items-center justify-center">
        <CircularWaveform
          audioTrack={botAudioTrack ?? null}
          isThinking={isThinking}
          backgroundColor="transparent"
          numBars={64}
          barWidth={3}
          size={192}
        />
      </div>
    </div>
  )
}
