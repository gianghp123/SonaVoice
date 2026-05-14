"use client"

import { LoadingScreen } from "@/components/common/LoadingScreen"
import { ChatLayout } from "@/features/chat-interface/components/ChatLayout"
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit"
import { toast } from "sonner"


export default function ChatPage() {
  return (
    <PipecatAppBase
      transportType="smallwebrtc"
      startBotParams={{ 
        endpoint: "/api/proxy/webrtc/model-gateway/start",
      }}
      noThemeProvider
      clientOptions={{ enableMic: true }}
      initDevicesOnMount={true}
      connectOnMount={true}
    >
      {({ client, error, handleDisconnect }) => {
        if (!client) {
          return <LoadingScreen />
        }

        return <ChatLayout
          handleDisconnect={handleDisconnect ?? (() => { })}
          initialError={error}
        />
      }}
    </PipecatAppBase>
  )
}