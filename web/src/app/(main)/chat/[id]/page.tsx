"use client"

import { LoadingScreen } from "@/components/common/LoadingScreen"
import { ChatLayout } from "@/features/chat-interface/components/ChatLayout"
import { cancelSession } from "@/features/chat-interface/services/session.actions"
import { PROXY_ROUTES } from "@/lib/routes"
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit"
import * as Sentry from "@sentry/nextjs"
import { useParams, useRouter } from "next/navigation"
import { useCallback, useMemo, useState } from "react"

export default function ChatPage() {
  const params = useParams()
  const router = useRouter()
  const sessionId = params.id as string
  const [maxDuration, setMaxDuration] = useState(0)

  const startBotParams = useMemo(() => ({
    endpoint: PROXY_ROUTES.WEBRTC.START(sessionId),
  }), [sessionId])

  const clientOptions = useMemo(() => ({
    enableMic: true,
  }), [])

  const startBotResponseTransformer = useCallback((response: any) => {
    Sentry.logger.info("Voice session started", {
      area: "chat-page",
      sessionId,
      maxDuration: response.maxDuration,
    })
    setMaxDuration(response.maxDuration)
    return {
      webrtcUrl: PROXY_ROUTES.WEBRTC.OFFER(sessionId),
    }
  }, [sessionId])

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
          Sentry.logger.error("Voice session fatal error handled", {
            area: "chat-page",
            sessionId,
          })

          await cancelSession(sessionId)
          await handleDisconnect?.()
          router.push("/")
        }

        const handleSessionDisconnect = async () => {
          Sentry.logger.info("Voice session disconnected by user", {
            area: "chat-page",
            sessionId,
          })

          await handleDisconnect?.()
          router.push("/")
        }

        return (
          <ChatLayout
            maxDuration={maxDuration}
            handleError={handleSessionError}
            handleDisconnect={handleSessionDisconnect}
            initialError={error}
          />
        )
      }}
    </PipecatAppBase>
  )
}