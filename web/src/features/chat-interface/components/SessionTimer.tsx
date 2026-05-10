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
      <span className="size-1.5 rounded-full bg-green-500 animate-pulse" />
      <span className="text-lg font-bold text-primary">
        {time}
      </span>
    </div>
  )
}
