import { cn } from "@/lib/utils"

export function SessionTimer({
  time = "00:12:45",
  className,
}: {
  time?: string
  className?: string
}) {
  return (
    <div className={cn("flex items-center gap-1.5", className)}>
      <span className="size-1.5 rounded-full bg-destructive animate-pulse" />
      <span className="font-mono text-[11px] font-bold text-primary">
        {time}
      </span>
    </div>
  )
}
