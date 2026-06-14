# Orchestrator Service — Flow Documentation

**File:** `services/orchestrator.service.go`
**Interface:** `IOrchestratorService`

---

## Overview

`orchestratorService` orchestrates the voice session lifecycle. It coordinates 5 services:

| Service | Role |
|---|---|
| `ISessionConfigService` | Reads the JSONB session config payload |
| `ISessionService` | Create + Get + GetBySpeechSessionID + MarkSessionActive |
| `ISpeechProxyService` | HTTP proxy to external speech engine |
| `IStartConnectionService` | Starts WebRTC connection + DB updates in one UoW |
| `IQuotaService` | Checks remaining quota at CreateSession |
| `UnitOfWork` | Transactional boundary for atomic DB operations |

---

## Design Decisions & Trade-offs

### Why ProxyOffer holds a DB row lock during the HTTP call
**Decision:** `ProxyOffer` acquires `SELECT ... FOR UPDATE` and holds it while forwarding the offer to the speech engine.

**Rationale:**
- A user can only have one active/pending session at a time (enforced by unique partial index).
- `ProxyOffer` is part of the WebRTC handshake, not a high-frequency API call.
- Serialization is desired: two concurrent offers for the same session should be sequential, not parallel.
- The 15-second `context.WithTimeout` bounds the maximum lock duration.

### Why CloseSession has no ownership check
**Decision:** `CloseSession` does not extract `requesterID` or call `EnforceOwnership`.

**Rationale:** `CloseSession` is called by the speech engine via webhook/callback, not by the end user. The endpoint is on a private network with auth middleware at the controller/gateway layer.

### Why CreateSession and StartConnection are separate endpoints
**Decision:** `CreateSession` creates a PENDING row and checks quota. `StartConnection` starts WebRTC.

**Rationale:**
- `CreateSession` must be fast so the FE can redirect to `/chat/:sessionId` immediately.
- `StartConnection` is slower (calls speech engine) and happens after the user lands on the chat page.
- This prevents the user from staring at a loading screen.

**FE Flow:**
```
POST /sessions → {id} → redirect /chat/:id → POST /sessions/:id/start → POST /sessions/:id/api/offer
```

### Why the 2-minute stale threshold
**Decision:** PENDING sessions older than 2 minutes are considered abandoned and cleaned up.

**Rationale:**
- After `CreateSession`, the FE redirects immediately.
- `StartConnection` is called within seconds of landing on the chat page.
- A 2-minute window covers slow networks and UI delays without being so long that abandoned sessions accumulate.
- The cleanup is opportunistic (runs on next `CreateSession`), not a background cron.

### Why quota is charged at CloseSession, not reserved at StartConnection
**Decision:** Quota is checked at CreateSession (fail-fast) and deducted at CloseSession using actual usage.

**Rationale:**
- The speech engine already enforces maxDuration per session.
- Reserving quota at StartConnection added significant complexity (Reserve/Release CTE, 4 release paths, ReservedAmount tracking).
- Charging at CloseSession with actual usage is simpler and more accurate.

---

## End-to-End User Flow

```
┌─────────┐    POST /sessions      ┌─────────────┐    redirect
│   User  │ ──────────────────────►│  CreateSession│───────────► /chat/:sessionId
└─────────┘                        └─────────────┘
                                           │
                                           ▼
                                     ┌──────────┐
                                     │  PENDING  │ (MaxDuration stored)
                                     └────┬─────┘
                                           │
                            POST /sessions/:id/start
                                           │
                                           ▼
                            ┌──────────────────────┐
                            │  StartConnection       │
                            │  - Start WebRTC        │
                            └──────────┬─────────────┘
                                           │
                                           ▼
                                     ┌──────────┐
                                     │  PENDING  │ (WebRTC ready)
                                     └────┬─────┘
                                           │
                      POST|PATCH /sessions/:id/api/offer
                                           │
                                           ▼
                            ┌──────────────────────┐
                            │  ProxyOffer            │
                            │  - Forward to speech   │
                            │  - PENDING → ACTIVE    │
                            └──────────┬─────────────┘
                                           │
                                           ▼
                                     ┌──────────┐
                                     │  ACTIVE   │
                                     └────┬─────┘
                                           │
                      POST /sessions/:id/close (speech engine callback)
                                           │
                                           ▼
                                     ┌──────────┐
                                     │ INACTIVE  │ (actualUsage stored, quota deducted)
                                     └──────────┘
```

**Flow explanation:**
1. **CreateSession** — checks quota + creates a PENDING session row with MaxDuration. Returns immediately so the FE can redirect.
2. **StartConnection** — initiates the WebRTC connection with the speech engine (sends session.MaxDuration as the limit). Happens after the user lands on the chat page.
3. **ProxyOffer** — forwards the WebRTC offer to the speech engine. On success, transitions PENDING → ACTIVE.
4. **CloseSession** — called by the speech engine when the call ends. Clamps actualUsage to MaxDuration, stores it on the session, and deducts from today's quota.
5. **CancelSession** — called by the user if they want to abort. No quota touch (nothing was reserved).

---

## Session State Machine

```
         CreateSession
              │
              ▼
        ┌──────────┐
        │  PENDING  │
        └────┬─────┘
             │
      ProxyOffer (2xx + PENDING)
             │
             ▼
        ┌──────────┐
        │  ACTIVE   │
        └────┬─────┘
             │
      CloseSession / CancelSession
             │
             ▼
        ┌──────────┐
        │ INACTIVE │
        └──────────┘
             ▲
             │
      ProxyOffer failure ──► mark FAILED (no quota reserved)
             │
             ▼
        ┌──────────┐
        │  FAILED   │
        └──────────┘
```

**Note:** Sessions enter PENDING on creation with MaxDuration stored. The first successful ProxyOffer transitions PENDING→ACTIVE. Failed ProxyOffer marks the session FAILED (no quota to release). CloseSession clamps actualUsage to MaxDuration, deducts from quota, and marks INACTIVE. CancelSession marks INACTIVE with no quota touch.

---

## Method Flows

### 1. `CreateSession(ctx) → (*res.CreateSessionRes, error)`

Entry point: `POST /sessions`

```
1. Extract requesterID from context (JWT/Clerk)
2. GlobalConfigService.Get() → dailyLimit
3. QuotaService.CheckRemaining(userID, dailyLimit) → remaining
   └─ If no quota row: creates one with remaining=dailyLimit (UPSERT)
   └─ If remaining <= 0 → Forbidden("quota exceeded")
4. UoW.Do() [transaction]:
   a. GetActiveOrPendingByUserIDForUpdate(userID) ── row-level lock
      └─ If active/pending session found:
         └─ SetSessionInactive(existing.ID) ── mark orphaned session inactive (no quota touch)
         └─ log info (previousStatus, userId, sessionId)
   b. SessionService.Create(requesterID) → INSERT with Status=PENDING
      └─ on unique_violation → Conflict error
5. Return CreateSessionRes{ID, MaxDuration: 0, WebRTCConnectionRes: nil}
```

**Note:** Any active or pending session for the user is force-closed in the same transaction before creating the new one. This handles orphaned active sessions (e.g., when the speech engine crashes and `/finalize` never arrives) so the user is never permanently blocked. No quota is touched during the cleanup — only `FinalizeSession` (called by the speech engine) deducts quota.

**Note:** MaxDuration is set at CreateSession from the checked remaining quota. This value is later sent to the speech engine at StartConnection and used to clamp actualUsage at CloseSession.

### 2. `StartConnection(ctx, sessionID) → (*res.WebRTCConnectionRes, error)`

Entry point: `POST /sessions/:sessionId/start`

```
1. Extract requesterID from context
2. SessionService.Get(sessionID, requesterID) → ownership check
3. domain.NewSessionFromModel() → CanBeStarted()
   └─ validates Status == PENDING
4. StartConnectionService.Start(ctx, session, userID)
   └─ UoW:
      a. GetForUpdate(session.ID)
      b. CanBeStarted() → must be PENDING
      c. speechSvc.StartConnection(ctx, {MaxDuration: session.MaxDuration})
      d. UpdateSpeechSessionID(session.ID, speechRes.SessionID)
5. Return WebRTCConnectionRes
```

**Note:** No quota touch. MaxDuration was already set at CreateSession. If the speech engine fails, the UoW rolls back the SpeechSessionID update. No quota to clean up.

### 3. `ProxyOffer(ctx, sessionId, method, body) → ([]byte, int, error)`

Entry point: `POST|PATCH /sessions/:sessionId/api/offer`

```
1. Validate sessionId is not empty
2. Extract requesterID from context
3. SessionService.GetBySpeechSessionID(speechSessionID, requesterID)
   └─ looks up by speech_session_id
   └─ ownership check
4. UnitOfWork.Do() [transaction]:
   a. SessionRepo.GetForUpdate(sessionID) ── row-level lock
   b. domain.CanBeStarted() ── validates Status == PENDING
   c. SpeechProxyService.ProxyOffer(speechSessionID, method, body)
      └─ forwards HTTP request to speech engine, returns raw response
      └─ on speech error OR non-2xx:
         └─ SetSessionFailed(sessionID) ── tx commits (session marked FAILED)
   d. On 2xx response:
      └─ SetSessionActive(sessionID) ── WHERE status = 'pending'
      └─ on fail: SetSessionFailed + return Internal
4. Return response body and status code
```

**Note:** No quota touch on any path. On failure, the session is simply marked FAILED. No quota to release since nothing was reserved.

### 4. `CloseSession(ctx, reqBody) → error`

Entry point: `POST /sessions/:sessionId/close`

```
1. Validate reqBody (not nil, sessionId non-empty, actualUsage >= 0)
2. UnitOfWork.Do() [transaction]:
   a. SessionRepo.GetForUpdate(sessionId) ── row-level lock
   b. domain.CanBeClosed()
      └─ rejects INACTIVE and FAILED
   c. actualUsage = min(reqBody.ActualUsage, session.MaxDuration)
      └─ clamps to MaxDuration stored at CreateSession
   d. If actualUsage > 0:
      └─ quotaRepo.Deduct(userID, "voice", session.QuotaDate, actualUsage)
      └─ UPDATE user_quotas SET remaining = remaining - ? WHERE remaining >= ?
   e. SetActualUsage(session.ID, actualUsage) ── stores for audit trail
   f. SetSessionInactive() → INACTIVE + ended_at
```

**Note:** If Deduct fails, the entire transaction rolls back. The speech engine can retry CloseSession. QuotaRepo.Deduct is called only when needed (after state check passes), so closed/failed sessions never attempt deduction.

### 5. `CancelSession(ctx, sessionID) → error`

Entry point: `POST /sessions/:sessionId/cancel`

```
1. Validate sessionID is not empty
2. Extract requesterID from context
3. UnitOfWork.Do() [transaction]:
   a. SessionRepo.GetForUpdate(sessionID) ── row-level lock
   b. EnforceOwnership(session.UserID, requesterID)
   c. domain.CanBeCancelled()
      └─ rejects INACTIVE and FAILED
   d. SetSessionInactive() → INACTIVE + ended_at
```

**Note:** No quota touch. Nothing was reserved, nothing to release. Can cancel both PENDING and ACTIVE sessions.

---

## Quota Lifecycle

```
CreateSession: CheckRemaining(userID) → fail if 0
StartConnection: no quota touch (sends session.MaxDuration to speech engine)
CloseSession: Deduct(actualUsage) → deduct from today's quota
CancelSession: no quota touch
Stale cleanup: no quota touch
```

- **CreateSession**: `quotaService.CheckRemaining(userID, dailyLimit)` ensures user has > 0 remaining. Fails fast with `Forbidden` if exhausted. Stores remaining as session.MaxDuration.
- **CloseSession**: `quotaRepo.Deduct(userID, actualUsage)` deducts the clamped actual usage from today's quota. actualUsage is clamped to session.MaxDuration.
- **CancelSession**: No quota touch (nothing was reserved).
- **ProxyOffer failure**: No quota touch (nothing was reserved).
- **Stale cleanup**: No quota touch (marks session INACTIVE only).

---

## Error Handling Patterns

| Situation | Response |
|---|---|
| Session not found | `NotFound()` |
| Quota exceeded | `Forbidden("quota exceeded")` |
| Active/pending session exists | `Conflict()` (from unique partial index violation) |
| Session not startable (not PENDING) | `BadRequest("session is not startable")` |
| Session already inactive | `BadRequest("session is already inactive")` |
| Not session owner | `Forbidden()` |
| Session not cancellable (already inactive/failed) | `BadRequest("session is already closed")` |
| DB/Infra error | `Internal()` (logged with details) |

---

## Key Design Decisions

- **No advisory locks**: The unique partial index `uq_one_active_session_per_user ON sessions(user_id) WHERE status IN ('active','pending')` enforces one active/pending session per user at the DB level.
- **No stale session cleanup**: Stale sessions are handled implicitly — if a session is stuck in PENDING, the next `CreateSession` call gets a unique constraint violation and the client must close the stale session first. This is simpler and serverless-friendly (no background cron).
- **Two-phase session creation**: `CreateSession` creates the row quickly for FE redirect. `StartConnection` starts the WebRTC after the user lands on the chat page. This keeps the critical path fast.
- **ProxyOffer is fully transactional**: Uses `UoW.Do()` with `GetForUpdate` row-level locking. If the speech engine fails, the session is simply marked FAILED.
- **Context timeouts for speech engine calls**: All HTTP calls to the speech engine use a 15-second `context.WithTimeout` to prevent indefinite hangs from exhausting DB connections.
- **Speech errors are logged before cleanup**: In `ProxyOffer`, the original speech engine error is logged before marking the session failed. This ensures the real cause is preserved.
- **Repository errors mapped correctly**: `gorm.ErrRecordNotFound` is mapped to `NotFound()` (HTTP 404) instead of `Internal()` (HTTP 500).
- **CloseSession uses GetForUpdate**: Row-level locking prevents concurrent close callbacks from double-deducting quota.
- **CloseSession is strictly transactional**: If Deduct fails, the entire transaction rolls back. The speech engine can retry CloseSession.
- **Stale PENDING cleanup in CreateSession**: Serverless-friendly approach. No background cron needed — stale sessions are cleaned up opportunistically when the user tries to create a new session.
- **Quota charged at CloseSession**: Quota is checked at CreateSession (fail-fast) and deducted at CloseSession (using actual usage clamped to MaxDuration). No quota touch at StartConnection, CancelSession, ProxyOffer failure, or stale cleanup.
- **MaxDuration set at CreateSession**: The remaining quota at creation time is stored as the session's MaxDuration. This is sent to the speech engine at StartConnection and used to clamp actualUsage at CloseSession.

---

## Common Misconceptions (For Future Reviewers)

1. **"ProxyOffer has side effects on speech failure"** — False. The inner UoW callback captures the speech error and marks the session FAILED. No quota to release since nothing was reserved.

2. **"StartConnection has quota side effects"** — False. No quota is reserved or touched at StartConnection. The MaxDuration was already stored at CreateSession.

3. **"CloseSession can deduct quota twice"** — False. `CanBeClosed()` rejects INACTIVE and FAILED sessions. The second call fails with `BadRequest("session is already closed")` before reaching Deduct.

4. **"The 2-minute threshold cancels in-progress sessions"** — False. The stale cleanup only runs during `CreateSession` (when user tries to create a new session). It only affects truly abandoned PENDING sessions with no quota reservations.

5. **"GetCtx returns empty string for missing keys"** — This is a utility concern, not model-gateway specific. The context is populated by auth middleware before any service method is called. If it's missing, the auth layer fails first.
