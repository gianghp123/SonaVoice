# Grammar Analysis Integration for Session History

## Overview

Integrate grammar analysis into the session history view, allowing users to trigger and persist grammar analyses for past conversation messages. The system will reuse the existing backend endpoint with `messageId` for persistence and follow established patterns for data fetching and revalidation.

## Requirements

1. Display grammar analysis button on user messages in session history
2. Trigger analysis via `POST /learning/grammar/messages/:messageId` (persists to DB)
3. Fetch existing analyses at server component level, pass as props
4. Use `updateTag` + `refresh()` for revalidation after trigger (full refresh pattern)
5. Extract shared logic from `HistoryPanelContent` into reusable hook

## Architecture

### Data Flow

```
Server Component (sessions/[id]/page.tsx)
  ├── getMessages(sessionId)           → IMessage[]
  ├── getGrammarAnalyses(sessionId)    → IGrammarAnalysis[]
  │
  ├── Join: messages.map(msg => ({
  │     ...msg,
  │     analysis: analysisMap.get(msg.id)
  │   }))
  │
  └── <SessionMessageList messages={messagesWithAnalysis} />

SessionMessageList (client component)
  ├── for each message:
  │   ├── render MessageBubble with transcript
  │   ├── if user role && no analysis → show SparkleButton
  │   └── if analysis exists → show AnalysisCard below
  │
  └── on SparkleButton click:
      └── analyzeGrammarByMessage(messageId)
            ├── POST /learning/grammar/messages/:messageId
            ├── updateTag("grammarAnalyses")
            └── refresh()
```

## Backend Changes

### 1. Repository: Add `GetBySessionID`

**File:** `services/api/internal/modules/learning/repositories/grammar-analysis.repository.go`

```go
func (r *grammarAnalysisRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*models.GrammarAnalysis, error) {
    var analyses []*models.GrammarAnalysis
    if err := r.db.WithContext(ctx).
        Where("session_id = ? AND deleted_at IS NULL", sessionID).
        Find(&analyses).Error; err != nil {
        return nil, err
    }
    return analyses, nil
}
```

### 2. Repository Interface: Add method signature

**File:** `services/api/internal/database/repository-interfaces/grammar-analysis-repository.interface.go`

```go
type IGrammarAnalysisRepository interface {
    Upsert(ctx context.Context, m *models.GrammarAnalysis) error
    GetByID(ctx context.Context, id string) (*models.GrammarAnalysis, error)
    GetByMessageID(ctx context.Context, messageID string) (*models.GrammarAnalysis, error)
    GetBySessionID(ctx context.Context, sessionID string) ([]*models.GrammarAnalysis, error)  // NEW
}
```

### 3. Service: Add `GetBySessionID`

**File:** `services/api/internal/modules/learning/services/grammar.service.go`

```go
// Add to interface
type IGrammarService interface {
    Analyze(ctx context.Context, messageID string, explanationLanguage string) (*models.GrammarAnalysis, *errors.AppError)
    AnalyzeText(ctx context.Context, transcript string, explanationLanguage string) (*models.GrammarAnalysis, *errors.AppError)
    GetBySessionID(ctx context.Context, sessionID string) ([]*models.GrammarAnalysis, *errors.AppError)  // NEW
}

// Implementation
func (s *grammarService) GetBySessionID(ctx context.Context, sessionID string) ([]*models.GrammarAnalysis, *errors.AppError) {
    analyses, err := s.grammarRepo.GetBySessionID(ctx, sessionID)
    if err != nil {
        return nil, errors.MapRepoError(err)
    }
    return analyses, nil
}
```

### 4. Controller: Add `HandleGetBySession`

**File:** `services/api/internal/modules/learning/controllers/grammar.controller.go`

```go
func (ctrl *GrammarController) HandleGetBySession(c *gin.Context) {
    sessionID := c.Param("sessionId")

    models, appErr := ctrl.svc.GetBySessionID(c.Request.Context(), sessionID)
    if appErr != nil {
        c.JSON(appErr.Code, response.Fail(appErr))
        return
    }

    var dtos []res.GrammarAIResult
    if err := utils.MapToDTO(models, &dtos); err != nil {
        sentry.CaptureException(err)
        c.JSON(http.StatusInternalServerError, response.Fail(errors.Internal()))
        return
    }

    c.JSON(http.StatusOK, response.Success(&dtos))
}
```

### 5. Module: Register new route

**File:** `services/api/internal/modules/learning/learning.module.go`

```go
sessionGroup := router.Group("/learning/grammar/sessions/:sessionId")
sessionGroup.GET("", authMiddlware, grammarCtrl.HandleGetBySession)
```

## Frontend Changes

### 1. Tags: Add `grammarAnalyses`

**File:** `web/src/lib/tags.ts`

```typescript
export const tags = {
  sessions: "sessions",
  profile: "profile",
  grammarAnalyses: "grammarAnalyses",  // NEW
} as const
```

### 2. Routes: Add grammar endpoints

**File:** `web/src/lib/routes.ts`

```typescript
LEARNING: {
  GRAMMAR: {
    ANALYZE: "/learning/grammar/analyze",
    BY_SESSION: (sessionId: string) => `/learning/grammar/sessions/${sessionId}`,  // NEW
    BY_MESSAGE: (messageId: string) => `/learning/grammar/messages/${messageId}`,  // NEW
  },
},
```

### 3. Server-only fetch: `grammar.get.ts`

**File:** `web/src/features/session-history/services/grammar.get.ts`

```typescript
import "server-only"

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import { tags } from "@/lib/tags"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"

export async function getGrammarAnalyses(sessionId: string) {
  return apiFetch<IGrammarAnalysis[]>(
    API_ROUTES.LEARNING.GRAMMAR.BY_SESSION(sessionId),
    {
      withCredentials: true,
      next: { tags: [tags.grammarAnalyses] },
    }
  )
}
```

### 4. Server action: `grammar.actions.ts`

**File:** `web/src/features/session-history/services/grammar.actions.ts`

```typescript
"use server"

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import { tags } from "@/lib/tags"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"
import { updateTag, refresh } from "next/cache"

export async function analyzeGrammarByMessage(
  messageId: string,
  explanationLanguage?: string
) {
  const result = await apiFetch<IGrammarAnalysis>(
    API_ROUTES.LEARNING.GRAMMAR.BY_MESSAGE(messageId),
    {
      method: "POST",
      withCredentials: true,
      body: { explanationLanguage },
    }
  )

  if (!result.error) {
    updateTag(tags.grammarAnalyses)
    refresh()
  }

  return result
}
```

### 5. Shared hook: `useGrammarAnalysis`

**File:** `web/src/lib/hooks/useGrammarAnalysis.ts`

```typescript
import { useT } from "next-i18next/client"
import { useState } from "react"
import { toast } from "sonner"
import { analyzeGrammarByMessage } from "@/features/session-history/services/grammar.actions"
import { FALLBACK_LANGUAGE, isSupportedLanguage, LANGUAGE_FULL_NAMES, SupportedLanguage } from "@/lib/i18n/i18n"

export function useGrammarAnalysis() {
  const { i18n } = useT()
  const [pendingId, setPendingId] = useState<string | null>(null)

  const explanationLanguage =
    LANGUAGE_FULL_NAMES[
      isSupportedLanguage(i18n.language)
        ? (i18n.language as SupportedLanguage)
        : FALLBACK_LANGUAGE
    ]

  const triggerAnalysis = async (messageId: string) => {
    setPendingId(messageId)

    try {
      const response = await analyzeGrammarByMessage(
        messageId,
        explanationLanguage
      )

      if (response.error) {
        toast.error(response.error.message)
      }
    } catch {
      toast.error("Failed to analyze grammar")
    } finally {
      setPendingId(null)
    }
  }

  return {
    pendingId,
    triggerAnalysis,
  }
}
```

### 6. Component: `GrammarAnalysisMessage`

**File:** `web/src/components/common/GrammarAnalysisMessage.tsx`

```typescript
"use client"

import { MessageBubble } from "@/components/common/MessageBubble"
import { AnalysisCard } from "@/components/common/AnalysisCard"
import { MessageAction, MessageActions } from "@/components/prompt-kit/message"
import { Button } from "@/components/ui/button"
import { MessageRole } from "@/lib/enums/message-role.enum"
import { useGrammarAnalysis } from "@/lib/hooks/useGrammarAnalysis"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"
import { Loader2, Sparkle } from "lucide-react"
import { useT } from "next-i18next/client"

interface GrammarAnalysisMessageProps {
  messageId: string
  transcript: string
  analysis?: IGrammarAnalysis
}

export function GrammarAnalysisMessage({
  messageId,
  transcript,
  analysis,
}: GrammarAnalysisMessageProps) {
  const { t } = useT("chat")
  const { pendingId, triggerAnalysis } = useGrammarAnalysis()
  const isLoading = pendingId === messageId

  return (
    <div className="flex flex-col">
      <MessageBubble
        role={MessageRole.User}
        action={
          !analysis && (
            <MessageActions className="self-end!">
              <MessageAction tooltip={t("analyze_grammar")}>
                <Button
                  variant="ghost"
                  size="icon"
                  type="button"
                  disabled={pendingId !== null}
                  onClick={() => triggerAnalysis(messageId)}
                  className="h-6 w-6 text-muted-foreground hover:text-foreground"
                >
                  {isLoading ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    <Sparkle className="h-3 w-3 text-purple-500" />
                  )}
                </Button>
              </MessageAction>
            </MessageActions>
          )
        }
      >
        {transcript}
      </MessageBubble>

      {analysis && (
        <MessageBubble role={MessageRole.Analysis} asChild>
          <AnalysisCard grammar={analysis} />
        </MessageBubble>
      )}
    </div>
  )
}
```

### 7. Server component: Update page

**File:** `web/src/app/(main)/(sidebar)/sessions/[id]/page.tsx`

```typescript
import { getMessages } from "@/features/session-history/services/messages.get"
import { getGrammarAnalyses } from "@/features/session-history/services/grammar.get"
import { SessionMessageList } from "@/features/session-history/components/SessionMessageList"

export default async function SessionPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  const [messagesRes, analysesRes] = await Promise.all([
    getMessages(id),
    getGrammarAnalyses(id),
  ])

  if (messagesRes.error) {
    // ... existing error handling
  }

  const analysisMap = new Map(
    (analysesRes.data ?? []).map((a) => [a.messageId, a])
  )

  const messagesWithAnalysis = (messagesRes.data ?? []).map((msg) => ({
    ...msg,
    analysis: analysisMap.get(msg.id),
  }))

  return (
    // ... existing layout
    <SessionMessageList messages={messagesWithAnalysis} />
  )
}
```

### 8. Client component: Update `SessionMessageList`

**File:** `web/src/features/session-history/components/SessionMessageList.tsx`

```typescript
"use client"

import { GrammarAnalysisMessage } from "@/components/common/GrammarAnalysisMessage"
import { MessageBubble } from "@/components/common/MessageBubble"
import { MessageRole } from "@/lib/enums/message-role.enum"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"
import type { IMessage } from "@/lib/types/message.interface"
import { useT } from "next-i18next/client"

interface IMessageWithAnalysis extends IMessage {
  analysis?: IGrammarAnalysis
}

interface SessionMessageListProps {
  messages: IMessageWithAnalysis[]
}

export function SessionMessageList({ messages }: SessionMessageListProps) {
  const { t } = useT("session")

  if (messages.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <p className="text-sm text-muted-foreground">{t("no_messages")}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6 px-5 w-full max-w-7/12">
      {[...messages]
        .sort(
          (a, b) =>
            new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        )
        .map((message) => {
          if (message.role === MessageRole.User) {
            return (
              <GrammarAnalysisMessage
                key={message.id}
                messageId={message.id}
                transcript={message.transcript}
                analysis={message.analysis}
              />
            )
          }

          return (
            <MessageBubble
              key={message.id}
              role={message.role}
              timestamp={new Date(message.createdAt)}
              wasInterrupted={message.wasInterrupted}
            >
              {message.transcript}
            </MessageBubble>
          )
        })}
    </div>
  )
}
```

### 9. Refactor: Update `HistoryPanelContent`

**File:** `web/src/features/chat-interface/components/HistoryPanelContent.tsx`

Replace local state management with shared hook:

```typescript
// Remove:
// const [analyses, setAnalyses] = useState<Record<number, IGrammarAnalysis>>({})
// const [pendingIndex, setPendingIndex] = useState<number | null>(null)
// const handleAnalyzeGrammar = async (index, transcript) => { ... }

// Add:
import { useGrammarAnalysis } from "@/lib/hooks/useGrammarAnalysis"

// In component:
const { pendingId, triggerAnalysis } = useGrammarAnalysis()

// Update button onClick to use triggerAnalysis(message.id) or triggerAnalysis(String(i))
// Note: HistoryPanelContent uses index-based messages from Pipecat, may need adaptation
```

**Note:** `HistoryPanelContent` uses Pipecat's `usePipecatConversation()` which returns messages without IDs. The refactoring may need to use index-based approach or wait for Pipecat to provide message IDs. This can be a follow-up task.

## Interface Definitions

### IMessageWithAnalysis

```typescript
interface IMessageWithAnalysis extends IMessage {
  analysis?: IGrammarAnalysis
}
```

### IGrammarAnalysis (existing)

```typescript
export interface IGrammarAnalysis {
  id?: string
  messageId?: string
  originalText: string
  correctedText: string
  explanation: string
  hasCorrection: boolean
  severity: "low" | "medium" | "high"
  practiceSentence?: string
  practiceFocus?: string
  practiceReason?: string
}
```

## File Changes Summary

| Action | File | Purpose |
|--------|------|---------|
| Modify | `services/api/internal/modules/learning/repositories/grammar-analysis.repository.go` | Add `GetBySessionID` |
| Modify | `services/api/internal/database/repository-interfaces/grammar-analysis-repository.interface.go` | Add interface method |
| Modify | `services/api/internal/modules/learning/services/grammar.service.go` | Add `GetBySessionID` |
| Modify | `services/api/internal/modules/learning/controllers/grammar.controller.go` | Add `HandleGetBySession` |
| Modify | `services/api/internal/modules/learning/learning.module.go` | Register new route |
| Modify | `web/src/lib/tags.ts` | Add `grammarAnalyses` tag |
| Modify | `web/src/lib/routes.ts` | Add grammar routes |
| Create | `web/src/features/session-history/services/grammar.get.ts` | Server-only fetch |
| Create | `web/src/features/session-history/services/grammar.actions.ts` | Server action |
| Create | `web/src/lib/hooks/useGrammarAnalysis.ts` | Shared hook |
| Create | `web/src/components/common/GrammarAnalysisMessage.tsx` | Wrapper component |
| Modify | `web/src/app/(main)/(sidebar)/sessions/[id]/page.tsx` | Fetch and join data |
| Modify | `web/src/features/session-history/components/SessionMessageList.tsx` | Accept and render analyses |
| Modify | `web/src/features/chat-interface/components/HistoryPanelContent.tsx` | Refactor to use shared hook |
