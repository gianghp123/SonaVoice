import { AudioLines } from "lucide-react"
import { cn } from "@/lib/utils"
import Link from "next/link"
import { PAGE_ROUTES } from "@/lib/routes"

export function Logo({ className }: { className?: string }) {
  return (
    <Link
      className={cn(
        "flex items-center gap-2 font-semibold text-primary",
        className
      )}

      href={PAGE_ROUTES.HOME}
    >
      <AudioLines className="size-[1em]" strokeWidth={2.2} />

      <span className="text-[1em] leading-none">
        Sona
      </span>
    </Link>
  )
}
