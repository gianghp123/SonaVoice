"use client"

import { LoadingScreen } from "@/components/common/LoadingScreen"
import { ChatLayout } from "@/features/chat-interface/components/ChatLayout"
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit"
import { useParams } from "next/navigation"

export default function AuthenticatedChatPage() {
  const params = useParams()

  const sessionId = params.id as string

  return (
    <PipecatAppBase
      transportType="smallwebrtc"
      startBotParams={
        {
          endpoint: `/api/proxy/webrtc/sessions/${sessionId}/start`,
        }
      }
      startBotResponseTransformer={(response: any) => {
        return {
          webrtcUrl: `/api/proxy/webrtc/sessions/${response.sessionId}/api/offer`
        }
      }
      }
      noThemeProvider
      clientOptions={{
        enableMic: true,
      }}
      initDevicesOnMount={true}
      connectOnMount={true}
    >
      {({ client, error, handleDisconnect }) => {
        if (!client) {
          return <LoadingScreen />
        }

        return (
          <ChatLayout
            handleDisconnect={handleDisconnect ?? (() => { })}
            initialError={error}
          />
        )
      }}
    </PipecatAppBase>
  )
}