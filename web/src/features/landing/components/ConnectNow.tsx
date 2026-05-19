"use client"

import { useRouter } from "next/navigation"
import { useAuth } from "@clerk/nextjs"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { createSession } from "@/features/chat-interface/services/session.actions"
import { useState } from "react"
import { Loader2 } from "lucide-react"
import { toast } from "sonner"

export function ConnectNow() {
  const router = useRouter()
  const { isSignedIn, isLoaded } = useAuth()
  const [loading, setLoading] = useState(false)

  const handleConnect = async () => {
    setLoading(true)

    try {
      if (!isSignedIn) {
        router.push("/chat")
        return
      }

      const res = await createSession()

      console.log(res)

      if (res.error) {
        toast.error(res.error.message || "Failed to create session")
        return
      }

      const sessionId = res.data?.id

      if (!sessionId) {
        toast.error("Failed to create session")
        return
      }

      router.push(`/chat/${sessionId}`)
    } catch {
      toast.error("Something went wrong. Please try again or contact support.")
    } finally {
      setLoading(false)
    }
  }

  if (!isLoaded) {
    return <Skeleton className="h-10 w-full" />
  }

  return (
    <Button onClick={handleConnect} disabled={loading} size="lg" className="w-full">
      {loading && <Loader2 className="animate-spin" />}
      Connect Now
    </Button>
  )
}