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
        <CircularWaveform
          audioTrack={botAudioTrack ?? null}
          isThinking={isThinking}
          backgroundColor="transparent"
          numBars={64}
          barWidth={3}
          size={250}
        />
    </div>
  )
}
