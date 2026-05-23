"use client"

import { useRouter } from "next/navigation"
import { useAuth } from "@clerk/nextjs"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { createSession } from "@/features/chat-interface/services/session.actions"
import { useState } from "react"
import { Loader2 } from "lucide-react"
import { toast } from "sonner"
import { PAGE_ROUTES } from "@/lib/routes"
import { cn } from "@/lib/utils"

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
        toast.error(res.error.message || "Failed to create session")
        return
      }

      const sessionId = res.data?.id

      if (!sessionId) {
        toast.error("Failed to create session")
        return
      }

      router.push(PAGE_ROUTES.CHAT.SESSION(sessionId))
    } catch {
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