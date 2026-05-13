"use client"

import { usePipecatClientTransportState } from "@pipecat-ai/client-react"
import { cn } from "@/lib/utils"
import { useEffect, useRef, useState } from "react"

function formatElapsed(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`
}

export function SessionTimer({ className }: { className?: string }) {
  const transportState = usePipecatClientTransportState()
  const [elapsed, setElapsed] = useState(0)
  const startTimeRef = useRef<number | null>(null)
  const isConnected = transportState === "connected" || transportState === "ready"

  useEffect(() => {
    if (isConnected) {
      startTimeRef.current = Date.now()
      const id = setInterval(() => {
        setElapsed(Math.floor((Date.now() - startTimeRef.current!) / 1000))
      }, 1000)
      return () => clearInterval(id)
    } else {
      startTimeRef.current = null
    }
  }, [isConnected])

  const displayElapsed = isConnected ? elapsed : 0

  return (
    <div className={cn("flex items-center gap-1.5", className)}>
      <span
        className={cn(
          "size-1.5 rounded-full",
          isConnected ? "bg-green-500 animate-pulse" : "bg-muted-foreground"
        )}
      />
      <span className="text-lg font-bold text-primary">
        {formatElapsed(displayElapsed)}
      </span>
    </div>
  )
}
