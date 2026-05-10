import { Mic, Database } from "lucide-react"

export function BottomNavBar() {
  return (
    <nav className="md:hidden fixed bottom-0 left-0 z-50 flex w-full items-center justify-center gap-6 border-t border-border bg-background py-2">
      <div className="flex flex-col items-center justify-center rounded-xl border border-secondary bg-secondary px-4 py-1 text-primary">
        <Mic className="size-[18px]" />
        <span className="text-[11px] font-semibold">Session</span>
      </div>
      <div className="flex flex-col items-center justify-center px-4 py-1 text-muted-foreground">
        <Database className="size-[18px]" />
        <span className="text-[11px] font-semibold">History</span>
      </div>
    </nav>
  )
}
