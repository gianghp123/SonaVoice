"use client"

import { usePipecatClientTransportState } from "@pipecat-ai/client-react"
import { cn } from "@/lib/utils"
import { useEffect, useRef, useState } from "react"

function formatElapsed(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60

  return `${String(h).padStart(2, "0")}:${String(m).padStart(
    2,
    "0"
  )}:${String(s).padStart(2, "0")}`
}

interface SessionTimerProps {
  className?: string
  maxDuration?: number
}

export function SessionTimer({
  className,
  maxDuration,
}: SessionTimerProps) {
  const transportState = usePipecatClientTransportState()

  const [elapsed, setElapsed] = useState(0)

  const startTimeRef = useRef<number | null>(null)
  const intervalRef = useRef<NodeJS.Timeout | null>(null)

  const isConnected =
    transportState === "connected" ||
    transportState === "ready"

  useEffect(() => {
    if (!isConnected) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }

      intervalRef.current = null
      startTimeRef.current = null

      return
    }

    if (!startTimeRef.current) {
      startTimeRef.current = Date.now()
      setElapsed(0)

      intervalRef.current = setInterval(() => {
        setElapsed(
          Math.floor(
            (Date.now() - startTimeRef.current!) / 1000
          )
        )
      }, 1000)
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    }
  }, [isConnected])

  const displayedElapsed = isConnected ? elapsed : 0

  const progress =
    maxDuration && maxDuration > 0
      ? Math.min((displayedElapsed / maxDuration) * 100, 100)
      : 0

  return (
    <div className={cn("flex flex-col items-center gap-1", className)}>
      <div className="flex items-center gap-1.5">
        <span
          className={cn(
            "size-1.5 rounded-full",
            isConnected
              ? "animate-pulse bg-green-500"
              : "bg-muted-foreground"
          )}
        />

        <span className="text-lg font-bold text-primary">
          {formatElapsed(displayedElapsed)}
        </span>

        {maxDuration ? (
          <span className="text-sm text-muted-foreground">
            / {formatElapsed(maxDuration)}
          </span>
        ) : null}
      </div>

      {maxDuration ? (
        <div className="h-1 w-40 overflow-hidden rounded-full bg-muted">
          <div
            className="h-full rounded-full bg-primary transition-all"
            style={{ width: `${progress}%` }}
          />
        </div>
      ) : null}
    </div>
  )
}