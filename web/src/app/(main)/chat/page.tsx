"use client"

import { LoadingScreen } from "@/components/common/LoadingScreen"
import { ChatLayout } from "@/features/chat-interface/components/ChatLayout"
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit"


export default function ChatPage() {
  return (
    <PipecatAppBase
      transportType="smallwebrtc"
      startBotParams={{ 
        endpoint: "api/proxy/webrtc/model-gateway/start",
        requestData: {
          session_id: "session-id",
        }
      }}
      noThemeProvider
      clientOptions={{ enableMic: true }}
      initDevicesOnMount={true}
      connectOnMount={true}
    >
      {({ client, handleDisconnect }) => {
        if (!client) {
          return <LoadingScreen />
        }

        return <ChatLayout
          handleDisconnect={handleDisconnect ?? (() => { })}
        />
      }}
    </PipecatAppBase>
  )
}