"use client"

import { LoadingScreen } from "@/components/common/LoadingScreen"
import { ChatLayout } from "@/features/chat-interface/components/ChatLayout"
import { IWebRTCConnection } from "@/lib/types/webtc-connection.interface"
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit"
import { notFound, useParams, useRouter } from "next/navigation"
import { useEffect, useMemo, useState } from "react"
import { toast } from "sonner"

export default function AuthenticatedChatPage() {
  const params = useParams()
  const router = useRouter()

  const sessionId = params.id as string

  const [connection, setConnection] = useState<IWebRTCConnection | null>(null)
  const [isLoaded, setIsLoaded] = useState(false)

  useEffect(() => {
    if (!sessionId) {
      toast.error("Session not found")
      notFound()
    }

    const storageKey = `webrtcConnection:${sessionId}`
    const rawConnection = sessionStorage.getItem(storageKey)

    if (!rawConnection) {
      toast.error("Connection data was lost. Please start a new chat.")
      return
    }

    try {
      const parsedConnection = JSON.parse(rawConnection) as IWebRTCConnection

      if (!parsedConnection.sessionId) {
        toast.error("Invalid connection data. Please start a new chat.")
        return
      }

      setConnection(parsedConnection)
      setIsLoaded(true)
    } catch {
      toast.error("Invalid connection data. Please start a new chat.")
    }
  }, [sessionId, router])

  const connectParams = useMemo(() => {
    if (!connection) return null

    return {
      sessionId: connection.sessionId,
      iceConfig: connection.iceConfig,
    }
  }, [connection])

  if (!isLoaded || !connectParams) {
    return <LoadingScreen />
  }

  return (
    <PipecatAppBase
      transportType="smallwebrtc"
      connectParams={connectParams}
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
            handleDisconnect={handleDisconnect ?? (() => {})}
            initialError={error}
          />
        )
      }}
    </PipecatAppBase>
  )
}