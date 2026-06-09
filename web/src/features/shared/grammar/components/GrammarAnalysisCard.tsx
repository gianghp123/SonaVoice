"use client"

import { Badge } from "@/components/ui/badge"
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
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"
import { ChevronRight } from "lucide-react"
import { useT } from "next-i18next/client"

interface GrammarAnalysisCardProps {
  grammar: IGrammarAnalysis
}

export function GrammarAnalysisCard({ grammar }: GrammarAnalysisCardProps) {
  const { t } = useT("chat")

  return (
    <Collapsible defaultOpen className="w-full">
      <Card className="p-0">
        <CollapsibleTrigger asChild>
          <Button
            variant="ghost"
            className="justify-start"
          >
            <ChevronRight className="h-4 w-4 transition-transform group-data-[state=open]:rotate-90" />
            <CardTitle className="text-sm">{t("grammar")}</CardTitle>
          </Button>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <CardContent className="space-y-4 p-3 pt-0">
            {grammar.hasCorrection ? (
              <>
                <div className="space-y-2">
                  <p className="text-sm text-muted-foreground">
                    {grammar.originalText}
                  </p>

                  <Badge variant="secondary" className="text-sm">
                    {grammar.correctedText}
                  </Badge>
                </div>

                {grammar.explanation && (
                  <p className="text-sm leading-relaxed">
                    {grammar.explanation}
                  </p>
                )}

                {grammar.practiceSentence && (
                  <>
                    <Separator />

                    <div className="space-y-2">
                      <p className="text-sm font-medium">{t("practice")}</p>

                      <div className="rounded-lg border bg-muted/40 p-3">
                        <p className="text-sm italic">
                          {grammar.practiceSentence}
                        </p>

                        {grammar.practiceFocus && (
                          <p className="mt-2 text-xs text-muted-foreground">
                            {t("focus")} {grammar.practiceFocus}
                          </p>
                        )}

                        {grammar.practiceReason && (
                          <p className="mt-1 text-xs text-muted-foreground">
                            {grammar.practiceReason}
                          </p>
                        )}
                      </div>
                    </div>
                  </>
                )}
              </>
            ) : (
              <p className="text-sm text-muted-foreground">
                {t("no_grammar_issues")}
              </p>
            )}
          </CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}