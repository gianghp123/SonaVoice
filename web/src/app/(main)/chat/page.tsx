"use client"

import { ChatLayout } from "@/features/chat-interface/components/ChatLayout"
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit"

const PIPECAT_ENDPOINT =
  process.env.NEXT_PUBLIC_PIPECAT_ENDPOINT ?? "http://localhost:7860/api/offer"


export default function ChatPage() {
  return (
    <PipecatAppBase
      transportType="smallwebrtc"
      connectParams={{ webrtcUrl: PIPECAT_ENDPOINT }}
      noThemeProvider
      clientOptions={{ enableMic: true }}
      initDevicesOnMount={true}
      connectOnMount={true}
    >
      {({ client, error, handleDisconnect }) => {
        if (error) return <div>Error: {error}</div>
        if (!client) return <div>Initializing...</div>
        return <ChatLayout
          handleDisconnect={handleDisconnect ?? (() => { })}
        />
      }}
    </PipecatAppBase>
  )
}