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
import React, { useCallback, useEffect, useState } from "react"
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

  useRTVIClientEvent(
    RTVIEvent.Error,
    useCallback((message: RTVIMessage) => {
      const { error, message: msg, fatal } = message.data as RTVIErrorData
      const text = error ?? msg ?? "Unknown RTVI error"

      Sentry.logger[fatal ? "error" : "warn"]("RTVI client error", {
        area: "chat-layout",
        stage: "rtvi",
        fatal: Boolean(fatal),
        error: text,
      })

      if (fatal) {
        Sentry.captureException(new Error(text), {
          tags: {
            area: "chat-layout",
            type: "rtvi-fatal-error",
          },
          extra: {
            rtviEventData: message.data,
          },
        })
      }

      toast.error(t('an_error_occurred', { text: text }), {
        duration: 10000,
      })

      if (fatal) {
        setFatalError(text.includes("408") ? t('timeout') : text)
      }
    }, [t])
  )

  useRTVIClientEvent(
    RTVIEvent.Disconnected,
    useCallback(() => {

      if (isUserDisconnecting?.current) {
        return
      }
      Sentry.logger.info("Disconnected from RTVI", {
        area: "chat-layout",
      })

      setFatalError((prev) => prev ?? t("bot_stopped"))
    }, [t, isUserDisconnecting?.current])
  )

  useEffect(() => {
    if (initialError) {
      Sentry.captureException(new Error(initialError), {
        tags: {
          area: "chat-layout",
          type: "pipecat-initial-error",
        },
      })
    }
  }, [initialError])

  return (
    <AlertDialog
      open={!!fatalError}
      onOpenChange={(open) => {
        if (!open && fatalError) return
      }}
    >
      <AlertDialogContent
        onEscapeKeyDown={(event) => event.preventDefault()}
      >
        <AlertDialogHeader>
          <AlertDialogTitle>{t('session_ended')}</AlertDialogTitle>
          <AlertDialogDescription>
            {fatalError || t('bot_stopped')}
          </AlertDialogDescription>
        </AlertDialogHeader>

        <AlertDialogFooter>
          <AlertDialogAction onClick={handleError}>
            {t('return_home')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

export { ErrorListener }
