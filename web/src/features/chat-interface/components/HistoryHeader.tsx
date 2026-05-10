import { PanelRight } from "lucide-react"
import { Button } from "@/components/ui/button"

interface HistoryHeaderProps {
  onClose: () => void
}

export function HistoryHeader({ onClose }: HistoryHeaderProps) {
  return (
    <div className="flex items-center justify-between p-4 border-b border-border bg-card sticky top-0 z-10">
      <div className="flex items-center gap-2">
        <h2 className="text-[13px] font-bold leading-relaxed tracking-tight text-primary">
          History
        </h2>
      </div>
      <Button variant="ghost" size="icon-sm" onClick={onClose}>
        <PanelRight />
      </Button>
    </div>
  )
}
