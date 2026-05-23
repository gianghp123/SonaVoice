import { Volume2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

interface Suggestion {
  original: string
  corrected: string
  hint?: string
}

interface Pronunciation {
  word: string
  phonetic: string
}

interface AnalysisCardProps {
  suggestions?: Suggestion
  pronunciation?: Pronunciation
}

export function AnalysisCard({
  suggestions,
  pronunciation,
}: AnalysisCardProps) {
  return (
    <Card className="border-secondary shadow-sm">
      <CardContent className="flex flex-col gap-4">
        {suggestions && (
          <div className="flex flex-col gap-2">
            <span className="font-mono text-[9px] font-bold uppercase tracking-wider text-muted-foreground">
              Suggestions
            </span>
            <div className="flex flex-wrap items-center gap-1 rounded-md border border-secondary/50 bg-muted p-2 text-sm">
              {suggestions.hint && (
                <span className="italic text-muted-foreground">
                  {suggestions.hint}
                </span>
              )}
              <span>{suggestions.original}</span>
              <span className="font-bold rounded bg-secondary/50 px-1 text-primary">
                {suggestions.corrected}
              </span>
            </div>
          </div>
        )}

        {pronunciation && (
          <div className="flex flex-col gap-2">
            <span className="font-mono text-[9px] font-bold uppercase tracking-wider text-muted-foreground">
              Pronunciation
            </span>
            <div className="flex items-center justify-between rounded-md border border-secondary/50 bg-muted p-2">
              <div className="flex items-center gap-2">
                <span className="text-sm font-bold">
                  {pronunciation.word}
                </span>
                <span className="font-mono text-xs font-medium text-primary">
                  {pronunciation.phonetic}
                </span>
              </div>
              <Button variant="ghost" size="icon-sm">
                <Volume2 />
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
