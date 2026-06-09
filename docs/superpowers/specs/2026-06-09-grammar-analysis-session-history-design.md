# Grammar Analysis Integration for Session History

## Overview

Integrate grammar analysis into the session history view, allowing users to trigger and persist grammar analyses for past conversation messages. The system will reuse the existing backend endpoint with `messageId` for persistence and follow established patterns for data fetching and revalidation.

## Requirements

1. Display grammar analysis sparkle button on ALL user messages (since endpoint uses upsert)
2. Trigger analysis via `POST /learning/grammar/messages/:messageId` (persists to DB)
3. Fetch existing analyses at server component level, join with messages into flat ordered list
4. Use `updateTag` + `refresh()` for revalidation after trigger (full refresh pattern)
5. Extract shared logic into reusable `useGrammarAnalysis` hook at `web/src/hooks/`

## Architecture

### Data Flow

```
Server Component (sessions/[id]/page.tsx)
  ├── getMessages(sessionId)           → IMessage[]
  ├── getGrammarAnalyses(sessionId)    → IGrammarAnalysis[]
  │
  ├── Build analysisMap: Map<messageId, IGrammarAnalysis>
  │
  ├── Build flat ordered items list:
  │   messages.forEach(msg => {
  │     items.push({ type: 'message', data: msg })
  │     if (analysisMap.has(msg.id))
  │       items.push({ type: 'analysis', data: analysis })
  │   })
  │
  └── <SessionMessageList items={items} />

SessionMessageList (client component)
  ├── for each item:
  │   ├── type === 'message' → MessageBubble
  │   │   └── if role === MessageRole.User → sparkle button (always shown)
  │   └── type === 'analysis' → MessageBubble(role=Analysis) + AnalysisCard
  │
  └── on SparkleButton click:
      └── analyzeGrammarByMessage(messageId)
            ├── POST /learning/grammar/messages/:messageId
            ├── updateTag("grammarAnalyses")
            └── refresh()
```

### SessionItem Type

```typescript
type SessionItem =
  | { type: 'message'; data: IMessage }
  | { type: 'analysis'; data: IGrammarAnalysis }
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

**File:** `web/src/hooks/useGrammarAnalysis.ts`

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

### 6. Server component: Update page

**File:** `web/src/app/(main)/(sidebar)/sessions/[id]/page.tsx`

```typescript
import { getMessages } from "@/features/session-history/services/messages.get"
import { getGrammarAnalyses } from "@/features/session-history/services/grammar.get"
import { SessionMessageList } from "@/features/session-history/components/SessionMessageList"
import type { IMessage } from "@/lib/types/message.interface"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"

type SessionItem =
  | { type: 'message'; data: IMessage }
  | { type: 'analysis'; data: IGrammarAnalysis }

export default async function SessionPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  const [messagesRes, analysesRes] = await Promise.all([
    getMessages(id),
    getGrammarAnalyses(id),
  ])

  if (messagesRes.error) {
    // ... existing error handling
  }

  // Build analysis lookup map
  const analysisMap = new Map(
    (analysesRes.data ?? []).map((a) => [a.messageId, a])
  )

  // Build flat ordered items list
  const items: SessionItem[] = []
  for (const msg of messagesRes.data ?? []) {
    items.push({ type: 'message', data: msg })
    const analysis = analysisMap.get(msg.id)
    if (analysis) {
      items.push({ type: 'analysis', data: analysis })
    }
  }

  return (
    // ... existing layout
    <SessionMessageList items={items} />
  )
}
```

### 7. Client component: Update `SessionMessageList`

**File:** `web/src/features/session-history/components/SessionMessageList.tsx`

```typescript
"use client"

import { MessageBubble } from "@/components/common/MessageBubble"
import { AnalysisCard } from "@/components/common/AnalysisCard"
import { MessageAction, MessageActions } from "@/components/prompt-kit/message"
import { Button } from "@/components/ui/button"
import { MessageRole } from "@/lib/enums/message-role.enum"
import { useGrammarAnalysis } from "@/hooks/useGrammarAnalysis"
import type { IMessage } from "@/lib/types/message.interface"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"
import { Loader2, Sparkle } from "lucide-react"
import { useT } from "next-i18next/client"

type SessionItem =
  | { type: 'message'; data: IMessage }
  | { type: 'analysis'; data: IGrammarAnalysis }

interface SessionMessageListProps {
  items: SessionItem[]
}

export function SessionMessageList({ items }: SessionMessageListProps) {
  const { t } = useT("session")
  const { t: tChat } = useT("chat")
  const { pendingId, triggerAnalysis } = useGrammarAnalysis()

  if (items.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <p className="text-sm text-muted-foreground">{t("no_messages")}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6 px-5 w-full max-w-7/12">
      {items.map((item, index) => {
        if (item.type === 'analysis') {
          return (
            <MessageBubble
              key={`analysis-${item.data.messageId ?? index}`}
              role={MessageRole.Analysis}
              asChild
            >
              <AnalysisCard grammar={item.data} />
            </MessageBubble>
          )
        }

        // type === 'message'
        const msg = item.data
        const isUser = msg.role === MessageRole.User

        return (
          <MessageBubble
            key={msg.id}
            role={msg.role}
            timestamp={new Date(msg.createdAt)}
            wasInterrupted={msg.wasInterrupted}
            action={
              isUser && (
                <MessageActions className="self-end!">
                  <MessageAction tooltip={tChat("analyze_grammar")}>
                    <Button
                      variant="ghost"
                      size="icon"
                      type="button"
                      disabled={pendingId !== null}
                      onClick={() => triggerAnalysis(msg.id)}
                      className="h-6 w-6 text-muted-foreground hover:text-foreground"
                    >
                      {pendingId === msg.id ? (
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
            {msg.transcript}
          </MessageBubble>
        )
      })}
    </div>
  )
}
```

### 8. Refactor: Update `HistoryPanelContent` (optional follow-up)

**File:** `web/src/features/chat-interface/components/HistoryPanelContent.tsx`

Replace local state management with shared hook:

```typescript
// Remove:
// const [analyses, setAnalyses] = useState<Record<number, IGrammarAnalysis>>({})
// const [pendingIndex, setPendingIndex] = useState<number | null>(null)
// const handleAnalyzeGrammar = async (index, transcript) => { ... }

// Add:
import { useGrammarAnalysis } from "@/hooks/useGrammarAnalysis"

// In component:
const { pendingId, triggerAnalysis } = useGrammarAnalysis()

// Update button onClick to use triggerAnalysis(message.id) or triggerAnalysis(String(i))
// Note: HistoryPanelContent uses index-based messages from Pipecat, may need adaptation
```

**Note:** `HistoryPanelContent` uses Pipecat's `usePipecatConversation()` which returns messages without IDs. The refactoring may need to use index-based approach or wait for Pipecat to provide message IDs. This can be a follow-up task.

## Interface Definitions

### SessionItem

```typescript
type SessionItem =
  | { type: 'message'; data: IMessage }
  | { type: 'analysis'; data: IGrammarAnalysis }
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
| Create | `web/src/hooks/useGrammarAnalysis.ts` | Shared hook |
| Modify | `web/src/app/(main)/(sidebar)/sessions/[id]/page.tsx` | Fetch, join, build flat items list |
| Modify | `web/src/features/session-history/components/SessionMessageList.tsx` | Accept items, render by type |
| Modify | `web/src/features/chat-interface/components/HistoryPanelContent.tsx` | Refactor to use shared hook (follow-up) |
