"use client"

import { LoadingScreen } from "@/components/common/LoadingScreen"
import { ChatLayout } from "@/features/chat-interface/components/ChatLayout"
import { cancelSession } from "@/features/chat-interface/services/session.actions"
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit"
import { useParams, useRouter } from "next/navigation"

export default function AuthenticatedChatPage() {
  const params = useParams()
  const router = useRouter()

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

        const handleSessionError = async () => {
          await cancelSession(sessionId)
          if (handleDisconnect) {
            await handleDisconnect()
          }
          router.push("/")
        }

        const handleSessionDisconnect = async () => {
          if (handleDisconnect) {
            await handleDisconnect()
          }
          router.push("/")
        }

        return (
          <ChatLayout
            handleError={handleSessionDisconnect}
            handleDisconnect={handleSessionError}
            initialError={error}
          />
        )
      }}
    </PipecatAppBase>
  )
}