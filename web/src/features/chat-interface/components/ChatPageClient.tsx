"use client"

import { LoadingScreen } from "@/components/common/LoadingScreen"
import { ChatInterface } from "@/features/chat-interface/components/ChatInterface"
import { ErrorListener } from "@/features/chat-interface/components/ErrorListener"
import { cancelSession } from "@/features/session-history/services/session.actions"
import { PAGE_ROUTES, PROXY_ROUTES } from "@/lib/routes"
import { EventsPanel, PipecatAppBase } from "@pipecat-ai/voice-ui-kit"
import * as Sentry from "@sentry/nextjs"
import { useRouter } from "next/navigation"
import { useCallback, useMemo, useState } from "react"

interface ChatPageClientProps {
  sessionId: string
}

export function ChatPageClient({ sessionId }: ChatPageClientProps) {
  const router = useRouter()
  const [maxDuration, setMaxDuration] = useState(0)

  const startBotParams = useMemo(() => ({
    endpoint: PROXY_ROUTES.WEBRTC.START(sessionId),
  }), [sessionId])

  const clientOptions = useMemo(() => ({
    enableMic: true,
  }), [])

  const startBotResponseTransformer = useCallback((response: unknown) => {
    const maxDuration =
      typeof response === "object" &&
        response !== null &&
        typeof (response as { maxDuration?: unknown }).maxDuration === "number"
        ? (response as { maxDuration: number }).maxDuration
        : 0
    Sentry.logger.info("Voice session started", {
      area: "chat-page",
      sessionId,
      maxDuration: maxDuration,
    })
    setMaxDuration(maxDuration)
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
          router.push(PAGE_ROUTES.HOME)
        }

        const handleSessionDisconnect = async () => {
          Sentry.logger.info("Voice session disconnect initiated", {
            area: "chat-page",
            sessionId,
          })

          const result = await cancelSession(sessionId)

          if (result.error) {
            if (result.error.code === 400) {
              Sentry.logger.info("Session already closed", {
                area: "chat-page",
                sessionId,
              })
            } else {
              Sentry.logger.error("Failed to cancel session", {
                area: "chat-page",
                sessionId,
                error: result.error.message,
              })
            }
          }

          await handleDisconnect?.()

          router.push(PAGE_ROUTES.HOME)
        }

        return (
          <>
            <ErrorListener
              handleError={handleSessionError}
              initialError={error}
            />
            <ChatInterface
              maxDuration={maxDuration}
              handleDisconnect={handleSessionDisconnect}
            />
            {/* <EventsPanel /> */}
          </>
        )
      }}
    </PipecatAppBase>
  )
}
