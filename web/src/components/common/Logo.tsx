import { cn } from "@/lib/utils"

export function Logo({ className }: { className?: string }) {
  return (
    <span
      className={cn(
        "text-lg font-semibold tracking-tight text-primary",
        className
      )}
    >
      Sona
    </span>
  )
}
