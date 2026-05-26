"use client"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { createSession } from "@/features/chat-interface/services/session.actions"
import { PAGE_ROUTES } from "@/lib/routes"
import { cn } from "@/lib/utils"
import { useAuth } from "@clerk/nextjs"
import * as Sentry from "@sentry/nextjs"
import { Loader2 } from "lucide-react"
import { useRouter } from "next/navigation"
import { useState } from "react"
import { toast } from "sonner"

type ConnectNowProps = React.ComponentProps<typeof Button> & {
  text?: string
}

export function ConnectNow({
  className,
  size = "lg",
  text = "Connect Now",
  disabled,
  children,
  ...props
}: ConnectNowProps) {
  const router = useRouter()
  const { isSignedIn, isLoaded } = useAuth()
  const [loading, setLoading] = useState(false)

  const handleConnect = async () => {
    setLoading(true)

    try {
      if (!isSignedIn) {
        router.push("/")
        return
      }

      const res = await createSession()

      if (res.error) {
        Sentry.logger.warn("Create session failed", {
          area: "connect-now",
          action: "create-session",
          message: res.error.message,
          code: res.error.code,
        })

        toast.error(res.error.message || "Failed to create session")
        return
      }

      const sessionId = res.data?.id

      if (!sessionId) {
        Sentry.logger.error("Create session succeeded but session id is missing", {
          area: "connect-now",
          action: "create-session",
        })

        toast.error("Failed to create session")
        return
      }

      router.push(PAGE_ROUTES.CHAT.SESSION(sessionId))
    } catch (error) {
      Sentry.captureException(error, {
        tags: {
          area: "connect-now",
          action: "create-session",
        },
      })

      toast.error("Something went wrong. Please try again or contact support.")
    } finally {
      setLoading(false)
    }
  }

  if (!isLoaded) {
    return <Skeleton className={cn("h-10 w-full", className)} />
  }

  return (
    <Button
      onClick={handleConnect}
      disabled={loading || disabled}
      size={size}
      className={cn("w-full justify-center gap-2", className)}
      {...props}
    >
      {loading && <Loader2 className="size-4 animate-spin" />}
      {children ?? text}
    </Button>
  )
}