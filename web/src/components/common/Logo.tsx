import { AudioLines } from "lucide-react"
import { cn } from "@/lib/utils"

export function Logo({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "flex items-center gap-2 font-semibold text-primary",
        className
      )}
    >
      <AudioLines className="size-[1em]" strokeWidth={2.2} />

      <span className="text-[1em] leading-none">
        Sona
      </span>
    </div>
  )
}