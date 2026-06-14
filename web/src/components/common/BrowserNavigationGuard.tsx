"use client"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { useRouter } from "next/navigation"
import { useT } from "next-i18next/client"
import { useEffect, useState } from "react"
import { PAGE_ROUTES } from "@/lib/routes"

interface BrowserNavigationGuardProps {
  enabled: boolean
  onConfirmLeave?: () => Promise<void>
}

export function BrowserNavigationGuard({
  enabled,
  onConfirmLeave,
}: BrowserNavigationGuardProps) {
  const router = useRouter()
  const { t } = useT("chat")
  const [showModal, setShowModal] = useState(false)
  const [isLeaving, setIsLeaving] = useState(false)

  // Tab close/refresh - native browser dialog
  useEffect(() => {
    if (!enabled) return

    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      event.preventDefault()
        ; (event as BeforeUnloadEvent & { returnValue: boolean }).returnValue = true
    }

    window.addEventListener("beforeunload", handleBeforeUnload)
    return () => window.removeEventListener("beforeunload", handleBeforeUnload)
  }, [enabled])

  // Browser back/forward - custom modal
  useEffect(() => {
    if (!enabled) return

    history.pushState({ guard: true }, "", location.href)

    const handlePopState = () => {
      setShowModal(true)
    }

    window.addEventListener("popstate", handlePopState)
    return () => window.removeEventListener("popstate", handlePopState)
  }, [enabled])

  const handleStay = () => {
    history.pushState({ guard: true }, "", location.href)
    setShowModal(false)
  }

  const handleLeave = async () => {
    setIsLeaving(true)
    await onConfirmLeave?.()
    router.push(PAGE_ROUTES.HOME)
  }

  return (
    <AlertDialog open={showModal}>
      <AlertDialogContent
        onEscapeKeyDown={(e) => e.preventDefault()}
      >
        <AlertDialogHeader>
          <AlertDialogTitle>{t("leave_session")}</AlertDialogTitle>
          <AlertDialogDescription>
            {t("leave_session_description")}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={handleStay} disabled={isLeaving}>
            {t("stay")}
          </AlertDialogCancel>
          <AlertDialogAction onClick={handleLeave} disabled={isLeaving}>
            {isLeaving ? t("leaving") : t("leave")}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
