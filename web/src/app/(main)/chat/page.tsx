"use client"

import { ChatLayout } from "@/features/chat-interface/components/ChatLayout"
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit"


export default function ChatPage() {
  return (
    <PipecatAppBase
      transportType="smallwebrtc"
      startBotParams={{ endpoint: "api/proxy/webrtc/model-gateway/start" }}
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