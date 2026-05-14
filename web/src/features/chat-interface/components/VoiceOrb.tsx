"use client"

import { usePipecatClientMediaTrack, usePipecatClientTransportState } from "@pipecat-ai/client-react"
import { CircularWaveform } from "@pipecat-ai/voice-ui-kit"

export function VoiceOrb() {
  const botAudioTrack = usePipecatClientMediaTrack("audio", "bot")
  const transportState = usePipecatClientTransportState()

  const isThinking = 
    transportState === "connecting" ||
    transportState === "authenticating" ||
    transportState === "connected" // bot joined but not yet speaking
  
  const isReady = transportState === "ready"

  return (
    <div className="relative flex items-center justify-center mb-8">
      <CircularWaveform
        audioTrack={isReady ? (botAudioTrack ?? null) : null}
        isThinking={isThinking}
        backgroundColor="transparent"
        numBars={64}
        barWidth={3}
        size={250}
        color1="#00D3F2"
        color2="#E12AFB"
      />
    </div>
  )
}