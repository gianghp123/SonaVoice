"use client"

import { Conversation, type FunctionCallRenderer } from "@pipecat-ai/voice-ui-kit"
import { AnalysisCard } from "./AnalysisCard"
import { HistoryHeader } from "./HistoryHeader"

interface FunctionCallArgs {
  original?: string
  corrected?: string
  word?: string
  phonetic?: string
}

const AnalysisRenderer: FunctionCallRenderer = (functionCall) => {
  const args = (functionCall.args ?? {}) as FunctionCallArgs
  return (
    <AnalysisCard
      suggestions={{
        hint: "Correction:",
        original: args.original ?? "",
        corrected: args.corrected ?? "",
      }}
      pronunciation={{
        word: args.word ?? "",
        phonetic: args.phonetic ?? "",
      }}
    />
  )
}

export function HistoryPanelContent() {
  return (
    <>
      <HistoryHeader />
      <Conversation
        noTextInput
        functionCallRenderer={AnalysisRenderer}
        classNames={{
          container: "flex flex-col gap-6",
        }}
      />
    </>
  )
}
