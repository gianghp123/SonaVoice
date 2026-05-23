"use client"

import { LoadingScreen } from "@/components/common/LoadingScreen"
import { ChatLayout } from "@/features/chat-interface/components/ChatLayout"
import { cancelSession } from "@/features/chat-interface/services/session.actions"
import { PROXY_ROUTES } from "@/lib/routes"
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit"
import { useParams, useRouter } from "next/navigation"
import { useCallback, useMemo } from "react"

export default function ChatPage() {
  const params = useParams()
  const router = useRouter()
  const sessionId = params.id as string

  const startBotParams = useMemo(() => ({
    endpoint: PROXY_ROUTES.WEBRTC.START(sessionId),
  }), [sessionId])

  const clientOptions = useMemo(() => ({
    enableMic: true,
  }), [])

  const startBotResponseTransformer = useCallback((_response: any) => ({
    webrtcUrl: PROXY_ROUTES.WEBRTC.OFFER(sessionId),
  }), [sessionId])

  return (
    <PipecatAppBase
      key={sessionId}
      transportType="smallwebrtc"
      startBotParams={startBotParams}
      startBotResponseTransformer={startBotResponseTransformer}
      clientOptions={clientOptions}
      initDevicesOnMount
      connectOnMount
      noThemeProvider
    >
      {({ client, error, handleDisconnect }) => {
        if (!client) return <LoadingScreen />

        const handleSessionError = async () => {
          await cancelSession(sessionId)
          await handleDisconnect?.()
          router.push("/")
        }

        const handleSessionDisconnect = async () => {
          await handleDisconnect?.()
          router.push("/")
        }

        return (
          <ChatLayout
            handleError={handleSessionError}
            handleDisconnect={handleSessionDisconnect}
            initialError={error}
          />
        )
      }}
    </PipecatAppBase>
  )
}