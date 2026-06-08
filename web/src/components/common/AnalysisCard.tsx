"use client"

import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ChevronRight, Volume2 } from "lucide-react"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"

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
  grammar?: IGrammarAnalysis
}

function SeverityBadge({ severity }: { severity: 'low' | 'medium' | 'high' }) {
  const variantMap = {
    high: 'destructive' as const,
    medium: 'default' as const,
    low: 'secondary' as const,
  }
  return (
    <Badge variant={variantMap[severity]}>
      {severity}
    </Badge>
  )
}

export function AnalysisCard({
  suggestions,
  pronunciation,
  grammar,
}: AnalysisCardProps) {
  return (
    <div className="flex flex-col gap-2">
      {grammar && (
        <Collapsible defaultOpen>
          <div className="rounded-md border border-secondary/50 bg-muted">
            <CollapsibleTrigger className="flex w-full items-center gap-2 p-2 hover:bg-secondary/30 transition-colors">
              <ChevronRight className="h-3 w-3 transition-transform [[data-state=open]>&]:rotate-90" />
              <span className="font-mono text-[9px] font-bold uppercase tracking-wider text-muted-foreground">
                Grammar
              </span>
              <SeverityBadge severity={grammar.severity} />
            </CollapsibleTrigger>
            <CollapsibleContent className="px-2 pb-2">
              {grammar.hasCorrection ? (
                <div className="flex flex-col gap-2">
                  <div className="flex flex-wrap items-center gap-1 rounded-md bg-background p-2 text-sm">
                    <span>{grammar.originalText}</span>
                    <span className="font-bold rounded bg-secondary/50 px-1 text-primary">
                      {grammar.correctedText}
                    </span>
                  </div>
                  {grammar.explanation && (
                    <p className="text-sm text-muted-foreground">{grammar.explanation}</p>
                  )}
                  {grammar.practiceSentence && (
                    <div className="flex flex-col gap-1">
                      <span className="font-mono text-[9px] font-bold uppercase tracking-wider text-muted-foreground">
                        Practice
                      </span>
                      <div className="rounded-md bg-background p-2 text-sm">
                        <p className="italic">{grammar.practiceSentence}</p>
                        {grammar.practiceFocus && (
                          <p className="mt-1 text-xs text-muted-foreground">
                            Focus: {grammar.practiceFocus}
                          </p>
                        )}
                        {grammar.practiceReason && (
                          <p className="text-xs text-muted-foreground">
                            {grammar.practiceReason}
                          </p>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <p className="text-sm text-green-600">No grammar issues</p>
              )}
            </CollapsibleContent>
          </div>
        </Collapsible>
      )}

      {suggestions && (
        <Collapsible defaultOpen>
          <div className="rounded-md border border-secondary/50 bg-muted">
            <CollapsibleTrigger className="flex w-full items-center gap-2 p-2 hover:bg-secondary/30 transition-colors">
              <ChevronRight className="h-3 w-3 transition-transform [[data-state=open]>&]:rotate-90" />
              <span className="font-mono text-[9px] font-bold uppercase tracking-wider text-muted-foreground">
                Suggestions
              </span>
            </CollapsibleTrigger>
            <CollapsibleContent className="px-2 pb-2">
              <div className="flex flex-wrap items-center gap-1 rounded-md bg-background p-2 text-sm">
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
            </CollapsibleContent>
          </div>
        </Collapsible>
      )}

      {pronunciation && (
        <Collapsible defaultOpen>
          <div className="rounded-md border border-secondary/50 bg-muted">
            <CollapsibleTrigger className="flex w-full items-center gap-2 p-2 hover:bg-secondary/30 transition-colors">
              <ChevronRight className="h-3 w-3 transition-transform [[data-state=open]>&]:rotate-90" />
              <span className="font-mono text-[9px] font-bold uppercase tracking-wider text-muted-foreground">
                Pronunciation
              </span>
            </CollapsibleTrigger>
            <CollapsibleContent className="px-2 pb-2">
              <div className="flex items-center justify-between rounded-md bg-background p-2">
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
            </CollapsibleContent>
          </div>
        </Collapsible>
      )}
    </div>
  )
}
