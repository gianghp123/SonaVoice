'use client'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { RTVIEvent, RTVIMessage } from "@pipecat-ai/client-js"
import { useRTVIClientEvent } from "@pipecat-ai/client-react"
import * as Sentry from "@sentry/nextjs"
import { useT } from "next-i18next/client"
import React, { useCallback, useRef, useState } from "react"
import { toast } from "sonner"

type RTVIErrorData = {
  error?: string
  message?: string
  fatal?: boolean
}

function ErrorListener({
  handleError,
  initialError,
  isUserDisconnecting,
}: {
  handleError: () => void | Promise<void>
  initialError?: string | null
  isUserDisconnecting?: React.RefObject<boolean>
}) {
  const { t } = useT('chat')
  const [fatalError, setFatalError] = useState<string | null>(
    initialError ?? null
  )
  const [open, setOpen] = useState<boolean>(Boolean(initialError))
  const lastErrorRef = useRef<string>("")

  useRTVIClientEvent(
    RTVIEvent.Error,
    useCallback((message: RTVIMessage) => {
      const { error, message: msg, fatal } = message.data as RTVIErrorData

      const text = error ?? msg ?? ""

      Sentry.logger[fatal ? "error" : "warn"]("RTVI client error", {
        area: "chat-layout",
        stage: "rtvi",
        fatal: Boolean(fatal),
        error: text,
      })

      Sentry.captureException(new Error(text || "Unknown RTVI error"), {
        tags: {
          area: "chat-layout",
          type: "rtvi-fatal-error",
        },
        extra: {
          rtviEventData: message.data,
        },
      })

      toast.error(t('an_error_occurred', { text }), {
        duration: 10000,
      })

      const finalText = text.includes("408") ? t('timeout') : text

      lastErrorRef.current = finalText
      setFatalError(finalText)
      setOpen(true)
    }, [t])
  )

  useRTVIClientEvent(
    RTVIEvent.Disconnected,
    useCallback(() => {
      if (isUserDisconnecting?.current) return

      Sentry.logger.info("Disconnected from RTVI", {
        area: "chat-layout",
      })

      const msg = t("bot_stopped")

      lastErrorRef.current = msg
      setFatalError((prev) => prev ?? msg)
      setOpen(true)
    }, [t, isUserDisconnecting?.current])
  )

  return (
    <AlertDialog
      open={open}
      onOpenChange={(value) => {
        if (!value) return
      }}
    >
      <AlertDialogContent
        onEscapeKeyDown={(e) => e.preventDefault()}
      >
        <AlertDialogHeader>
          <AlertDialogTitle>
            {t('session_ended')}
          </AlertDialogTitle>

          <AlertDialogDescription>
            {/* IMPORTANT: empty string is valid, so only fallback if null */}
            {fatalError ?? t('bot_stopped')}
          </AlertDialogDescription>
        </AlertDialogHeader>

        <AlertDialogFooter>
          <AlertDialogAction
            onClick={() => {
              setOpen(false)
              handleError()
            }}
          >
            {t('return_home')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

export { ErrorListener }
