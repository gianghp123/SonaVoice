"use client"

import { useEffect } from "react"

const DEFAULT_MESSAGE =
  "Leaving this page may interrupt your current voice session. Continue?"

interface BrowserNavigationGuardProps {
  enabled: boolean
  message?: string
}

export function BrowserNavigationGuard({
  enabled,
  message = DEFAULT_MESSAGE,
}: BrowserNavigationGuardProps) {
  useEffect(() => {
    if (!enabled) return

    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      event.preventDefault()

      // Required for browser compatibility despite being deprecated.
      ;(event as BeforeUnloadEvent & { returnValue: boolean }).returnValue = true
    }

    window.addEventListener("beforeunload", handleBeforeUnload)

    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload)
    }
  }, [enabled])

  useEffect(() => {
    if (!enabled) return

    const currentUrl =
      window.location.pathname + window.location.search + window.location.hash

    window.history.pushState({ browserNavigationGuard: true }, "", currentUrl)

    const handlePopState = () => {
      const confirmed = window.confirm(message)

      if (!confirmed) {
        window.history.pushState(
          { browserNavigationGuard: true },
          "",
          window.location.href
        )
        return
      }

      window.removeEventListener("popstate", handlePopState)
      window.history.back()
    }

    window.addEventListener("popstate", handlePopState)

    return () => {
      window.removeEventListener("popstate", handlePopState)
    }
  }, [enabled, message])

  return null
}