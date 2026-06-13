"use client"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardTitle
} from "@/components/ui/card"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import { Separator } from "@/components/ui/separator"
import type { IGrammarAIResult } from "@/lib/types/grammar-analysis.interface"
import { ChevronRight } from "lucide-react"
import { useT } from "next-i18next/client"

interface GrammarAnalysisCardProps {
  grammar: IGrammarAIResult
  originalText?: string
}

export function GrammarAnalysisCard({ grammar, originalText }: GrammarAnalysisCardProps) {
  const { t } = useT("chat")

  const hasCleanedTranscript =
    grammar.cleanedTranscript &&
    grammar.cleanedTranscript.trim() &&
    (!originalText || grammar.cleanedTranscript !== originalText)

  return (
    <Collapsible defaultOpen className="w-full min-w-0">
      <Card className="w-full min-w-0 overflow-hidden p-0">
        <CollapsibleTrigger asChild>
          <Button
            variant="ghost"
            className="group w-full justify-start"
          >
            <ChevronRight className="h-4 w-4 shrink-0 transition-transform group-data-[state=open]:rotate-90" />
            <CardTitle className="truncate text-sm">
              {t("grammar")}
            </CardTitle>
          </Button>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <CardContent className="min-w-0 space-y-4 p-3 pt-0">
            <div className="min-w-0 space-y-3">
              {originalText && (
                <div className="space-y-1">
                  <p className="text-xs font-medium text-muted-foreground">
                    {t("original")}
                  </p>
                  <p className="wrap-break-word text-sm text-muted-foreground">
                    {originalText}
                  </p>
                </div>
              )}

              {hasCleanedTranscript && (
                <div className="space-y-1">
                  <p className="text-xs font-medium text-muted-foreground">
                    {t("normalized")}
                  </p>
                  <div className="w-full rounded-md border px-2.5 py-1.5 text-sm leading-relaxed whitespace-pre-wrap break-words [overflow-wrap:anywhere]">
                    {grammar.cleanedTranscript}
                  </div>
                </div>
              )}

              {grammar.hasCorrection ? (
                <div className="space-y-1">
                  <p className="text-xs font-medium text-muted-foreground">
                    {t("corrected")}
                  </p>
                  <div className="w-full rounded-md bg-secondary px-2.5 py-1.5 text-sm leading-relaxed text-secondary-foreground whitespace-pre-wrap break-words [overflow-wrap:anywhere]">
                    {grammar.correctedText}
                  </div>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">
                  {t("no_grammar_issues")}
                </p>
              )}
            </div>

            {grammar.explanation && (
              <p className="wrap-break-word text-sm leading-relaxed">
                {grammar.explanation}
              </p>
            )}

            {grammar.practiceSentence && (
              <>
                <Separator />

                <div className="min-w-0 space-y-2">
                  <p className="text-sm font-medium">{t("practice")}</p>

                  <div className="min-w-0 rounded-lg border bg-muted/40 p-3">
                    <p className="wrap-break-word text-sm italic">
                      {grammar.practiceSentence}
                    </p>

                    {grammar.practiceFocus && (
                      <p className="mt-2 wrap-break-word text-xs text-muted-foreground">
                        {t("focus")} {grammar.practiceFocus}
                      </p>
                    )}

                    {grammar.practiceReason && (
                      <p className="mt-1 wrap-break-word text-xs text-muted-foreground">
                        {grammar.practiceReason}
                      </p>
                    )}
                  </div>
                </div>
              </>
            )}
          </CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}