"use client"

import { VoiceVisualizer } from "@pipecat-ai/voice-ui-kit"

export function WaveformVisualization() {
  return (
    <div className="flex items-end justify-center h-10 gap-0 mb-8">
      <VoiceVisualizer
        participantType="bot"
        barCount={12}
        barWidth={3}
        barGap={3}
        barMaxHeight={40}
        barLineCap="round"
        barOrigin="center"
        backgroundColor="transparent"
      />
    </div>
  )
}
