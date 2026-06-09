"use client"

import { Button } from "@/components/ui/button"
import { Loader2, Sparkle } from "lucide-react"

interface GrammarAnalysisButtonProps {
  isLoading?: boolean
  disabled?: boolean
  tooltip: string
  onClick: () => void
}

export function GrammarAnalysisButton({
  isLoading = false,
  disabled = false,
  tooltip,
  onClick,
}: GrammarAnalysisButtonProps) {
  return (
    <Button
      variant="ghost"
      size="icon"
      type="button"
      title={tooltip}
      aria-label={tooltip}
      disabled={disabled}
      onClick={onClick}
      className="h-6 w-6 text-muted-foreground hover:text-foreground"
    >
      {isLoading ? (
        <Loader2 className="h-3 w-3 animate-spin" />
      ) : (
        <Sparkle className="h-3 w-3 text-purple-500" />
      )}
    </Button>
  )
}