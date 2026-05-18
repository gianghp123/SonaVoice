# Model Gateway Service — Flow Documentation

**File:** `services/model-gateway.service.go`
**Interface:** `IModelGatewayService`

---

## Overview

`modelGatewayService` orchestrates the voice session lifecycle. It coordinates 4 services:

| Service | Role |
|---|---|
| `IGlobalConfigService` | Reads the JSONB global config payload |
| `ISessionService` | Create + Get + GetBySpeechSessionID + MarkSessionActive |
| `ISpeechProxyService` | HTTP proxy to external speech engine |
| `IStartConnectionService` | Atomic quota reserve + speech connection + DB updates in one UoW |
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

**Trade-off:** Under extreme load with a slow speech engine, the DB connection is tied up for up to 15s. This is acceptable because:
- The connection pool is sized for this workload
- Concurrent offers for the same session are inherently sequential (WebRTC handshake)
- The alternative (HTTP call outside tx + optimistic locking) adds complexity for minimal gain

### Why CloseSession has no ownership check
**Decision:** `CloseSession` does not extract `requesterID` or call `EnforceOwnership`.

**Rationale:** `CloseSession` is called by the speech engine via webhook/callback, not by the end user. The endpoint is on a private network with auth middleware at the controller/gateway layer.

### Why CreateSession and StartConnection are separate endpoints
**Decision:** `CreateSession` only creates a PENDING row. `StartConnection` reserves quota and starts WebRTC.

**Rationale:**
- `CreateSession` must be fast so the FE can redirect to `/chat/:sessionId` immediately.
- `StartConnection` is slower (reserves quota, calls speech engine) and happens after the user lands on the chat page.
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

### Why quota_date and reserved_amount are cleared on terminal transitions
**Decision:** `SetSessionFailed` and `SetSessionInactive` set `quota_date = nil` and `reserved_amount = 0`.

**Rationale:** Defense in depth. Even though `markSessionFailedAndReleaseQuota` releases quota before calling `SetSessionFailed`, clearing the fields prevents any future code from accidentally releasing the same quota again.

---

## End-to-End User Flow

```
┌─────────┐    POST /sessions      ┌─────────────┐    redirect
│   User  │ ──────────────────────►│  CreateSession│───────────► /chat/:sessionId
└─────────┘                        └─────────────┘
                                          │
                                          ▼
                                    ┌──────────┐
                                    │  PENDING  │
                                    └────┬─────┘
                                          │
                           POST /sessions/:id/start
                                          │
                                          ▼
                           ┌──────────────────────┐
                           │  StartConnection       │
                           │  - Reserve quota       │
                           │  - Start WebRTC        │
                           └──────────┬─────────────┘
                                          │
                                          ▼
                                    ┌──────────┐
                                    │  PENDING  │ (quota reserved, WebRTC ready)
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
                                    │ INACTIVE  │
                                    └──────────┘
```

**Flow explanation:**
1. **CreateSession** — lightweight, only creates a PENDING session row. Returns immediately so the FE can redirect.
2. **StartConnection** — heavier operation that reserves quota and initiates the WebRTC connection with the speech engine. Happens after the user lands on the chat page.
3. **ProxyOffer** — forwards the WebRTC offer to the speech engine. On success, transitions PENDING → ACTIVE.
4. **CloseSession** — called by the speech engine when the call ends. Releases unused quota.
5. **CancelSession** — called by the user if they want to abort. Releases all reserved quota.

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
      ProxyOffer failure ──► transactional: release quota + mark FAILED
             │
             ▼
        ┌──────────┐
        │  FAILED   │
        └──────────┘
```
```

**Note:** Sessions enter PENDING on creation. The first successful ProxyOffer transitions PENDING→ACTIVE. Failed ProxyOffer calls release reserved quota and mark the session FAILED atomically within the transaction. CloseSession and CancelSession transition any non-terminal session to INACTIVE.

---

## Method Flows

### 1. `CreateSession(ctx) → (*res.CreateSessionRes, error)`

Entry point: `POST /sessions`

```
1. Extract requesterID from context (JWT/Clerk)

   └─ 1b. Opportunistic stale PENDING cleanup:
   └─     └─ Check if user has an existing PENDING session older than 2 minutes
   └─     └─ If stale: release reserved quota + mark INACTIVE (within UoW with FOR UPDATE)
   └─     └─ If cleanup fails: log warning, proceed with creation anyway

2. SessionService.Create(requesterID)         → INSERT with Status=PENDING
   └─ on unique_violation (active/pending exists) → Conflict error

3. Return CreateSessionRes{ID, MaxDuration: 0, WebRTCConnectionRes: nil}
   └─ Client must call StartConnection separately to reserve quota and start WebRTC
```

**Note:** CreateSession only creates the session record. Quota reservation and WebRTC connection establishment happen in the separate `StartConnection` call.

**Why split CreateSession and StartConnection?**
- **CreateSession** must be fast (creates a DB row) so the FE can redirect immediately.
- **StartConnection** is slower (reserves quota, calls speech engine) and happens after the user is already on the chat page.
- This prevents the user from staring at a loading screen while quota is reserved and WebRTC is set up.

### 2. `StartConnection(ctx, sessionID) → (*res.WebRTCConnectionRes, error)`

Entry point: `POST /sessions/:sessionId/start`

```
1. Extract requesterID from context
2. SessionService.Get(sessionID, requesterID)  → ownership check
3. domain.NewSessionFromModel() → CanBeStarted()
   └─ validates Status == PENDING
4. GlobalConfigService.Get()                    → typed config
5. StartConnectionService.Start(session)        → same as CreateSession step 4
6. Return WebRTCConnectionRes{SessionID, MaxDuration, ...}
```

### 3. `ProxyOffer(ctx, sessionId, method, body) → ([]byte, int, error)`

Entry point: `POST|PATCH /sessions/:sessionId/api/offer`

```
1. Validate sessionId is not empty
2. Extract requesterID from context
3. SessionService.GetBySpeechSessionID(speechSessionID, requesterID)
   └─ looks up by speech_session_id (the external speech engine's session ID)
   └─ this is the `:sessionId` path parameter, which is actually the speech session ID
   └─ ownership check via speech_session_id
4. UnitOfWork.Do() [transaction]:
   a. SessionRepo.GetForUpdate(sessionID)       ── row-level lock for duration of offer
   b. domain.NewSessionFromModel() → CanBeStarted()
      └─ validates Status == PENDING
   c. SpeechProxyService.ProxyOffer(speechSessionID, method, body)
      └─ forwards HTTP request to speech engine, returns raw response
      └─ on speech error OR non-2xx:
         └─ QuotaRepo.Release(userID, "voice", QuotaDate, ReservedAmount) if applicable
         └─ SessionRepo.SetSessionFailed(sessionID) ── WHERE status IN ('pending','active')
         └─ tx commits (quota released, session marked FAILED)
   d. On 2xx response:
      └─ SessionRepo.SetSessionActive(sessionID) ── WHERE status = 'pending'
      └─ on fail: release quota + mark FAILED (same as 4c)
5. Return response body and status code
```

**Note:** ProxyOffer is now fully transactional. The row lock prevents concurrent offers for the same session. There are no compensations — success or failure is determined atomically within the transaction.

### 4. `CloseSession(ctx, reqBody) → error`

Entry point: `POST /sessions/:sessionId/close`

```
1. Validate reqBody (not nil, sessionId non-empty, actualUsage >= 0)
2. UnitOfWork.Do() [transaction]:
   a. SessionRepo.GetForUpdate(sessionId)       ── row-level lock
   b. domain.NewSessionFromModel(session) → CanBeClosed()
      └─ validates Status != INACTIVE
   c. domain.WantsQuotaRelease()?
      └─ true when quota_date IS NOT NULL (quota was reserved)
      └─ quotaRepo.Release(UserID, "voice", session.QuotaDate, unused)
         └─ unused = max(0, reserved - ClampActualUsage(actualUsage))
         └─ ClampActualUsage: [0, reservedAmount]
      └─ false when quota_date IS NULL (no quota reserved, e.g. test sessions)
   d. SessionRepo.SetSessionInactive() → INACTIVE + ended_at
```

### 5. `CancelSession(ctx, sessionID) → error`

Entry point: `POST /sessions/:sessionId/cancel`

```
1. Validate sessionID is not empty
2. Extract requesterID from context
3. UnitOfWork.Do() [transaction]:
   a. SessionRepo.GetForUpdate(sessionID)       ── row-level lock
   b. EnforceOwnership(session.UserID, requesterID)
   c. domain.NewSessionFromModel(session) → CanBeCancelled()
      └─ validates Status != INACTIVE and != FAILED
   d. domain.WantsQuotaRelease()?
      └─ true when quota_date IS NOT NULL
      └─ quotaRepo.Release(UserID, "voice", session.QuotaDate, session.ReservedAmount)
         └─ releases ALL reserved quota (not just unused)
   e. SessionRepo.SetSessionInactive() → INACTIVE + ended_at
```

**Key difference from CloseSession:**
- Client-facing endpoint (requires Clerk auth)
- Releases ALL reserved quota (not just unused), since actual usage is unknown
- Can cancel both PENDING and ACTIVE sessions

---

## Quota Lifecycle

```
Reserve (in StartConnectionService.Start)
   │
   ├── Success ──► quota_date stored on session
   │                  │
   │                  ├── Offer success ──► session becomes ACTIVE
   │                  │                    └─ SetSessionActive clears nothing (quota stays reserved)
   │                  │
   │                  ├── Offer failure ──► SetSessionFailed releases all quota + clears quota_date/reserved_amount
   │                  │
   │                  └── Close session ──► Release unused = max(0, reserved - actualUsed)
   │                                       └─ SetSessionInactive clears quota_date/reserved_amount
   │
   ├── Speech connection fails ──► UoW rolls back automatically, no cleanup needed
   │
   └── Reserve returns 0 ──► Forbidden("quota exceeded"), nothing to clean up
```

**Key differences from previous flow:**
- Reserve is a single atomic CTE query (no 3-query race condition)
- `quota_date` is stored on the session at reserve time (not computed on release)
- Release uses the stored `quota_date` — fixes day-boundary bug (quota reserved on day N, released on day N+1)

---

## Error Handling Patterns

| Situation | Response |
|---|---|
| Session not found | `NotFound()` |
| Quota exhausted | `Forbidden("quota exceeded")` |
| Active/pending session exists | `Conflict()` (from unique partial index violation) |
| Session not startable (not PENDING) | `BadRequest("session is not startable")` |
| Session already inactive | `BadRequest("session is already inactive")` |
| Not session owner | `Forbidden()` |
| Session not cancellable (already inactive/failed) | `BadRequest("session is already closed")` |
| DB/Infra error | `Internal()` (logged with details) |

---

## Key Design Decisions

- **No advisory locks**: The unique partial index `uq_one_active_session_per_user ON sessions(user_id) WHERE status IN ('active','pending')` enforces one active/pending session per user at the DB level, replacing the old `pg_advisory_xact_lock` approach.
- **No stale session cleanup**: Removed `SessionJanitor` entirely. Stale sessions are handled implicitly — if a session is stuck in PENDING, the next `CreateSession` call gets a unique constraint violation and the client must close the stale session first. This is simpler and serverless-friendly (no background cron).
- **Atomic quota reserve**: Single CTE query (`INSERT ... ON CONFLICT ... RETURNING remaining`) eliminates the race between check-and-update in the old 3-query approach.
- **quota_date on session**: The date is captured at reserve time and stored on the session row, fixing the day-boundary bug where `today()` at reserve time differs from `today()` at release time.
- **Two-phase session creation**: `CreateSession` creates the row quickly for FE redirect. `StartConnection` does the heavy lifting (quota + WebRTC) after the user lands on the chat page. This keeps the critical path fast.
- **ProxyOffer is fully transactional**: Uses `UoW.Do()` with `GetForUpdate` row-level locking. Quota release and session status change happen atomically — no compensations needed. If the speech engine fails, the session is marked FAILED and quota is released within the same transaction.
- **Context timeouts for speech engine calls**: All HTTP calls to the speech engine use a 15-second `context.WithTimeout` to prevent indefinite hangs from exhausting DB connections.
- **Speech errors are logged before cleanup**: In `ProxyOffer`, the original speech engine error is logged before attempting quota release and session cleanup. This ensures the real cause is preserved even if the cleanup itself fails.
- **Repository errors mapped correctly**: `gorm.ErrRecordNotFound` from `Get` and `GetBySpeechSessionID` is mapped to `NotFound()` (HTTP 404) instead of `Internal()` (HTTP 500).
- **CloseSession uses GetForUpdate**: Row-level locking prevents double quota release from concurrent close callbacks.
- **No `cleanupFailedSession` after StartConnection UoW failure**: The UoW transaction is atomic. If it fails, PostgreSQL rolls back everything including the quota reserve. Running external cleanup would inflate quota by releasing what was never deducted.
- **ProxyOffer failure releases quota**: Failed offers release reserved quota and mark the session FAILED within the same transaction.
- **CloseSession is strictly transactional**: If quota release fails, the entire transaction rolls back. The speech engine can retry CloseSession.
- **Stale PENDING cleanup in CreateSession**: Serverless-friendly approach. No background cron needed — stale sessions are cleaned up opportunistically when the user tries to create a new session.

---

## Common Misconceptions (For Future Reviewers)

1. **"ProxyOffer rolls back its own cleanup"** — Fixed. The inner UoW callback captures the speech error in a closure variable (`proxyErr`) and returns `nil` so the cleanup (quota release + FAILED status) commits atomically. The captured error is returned after the transaction succeeds.

2. **"StartConnection leaves quota reserved on speech failure"** — False. `Reserve`, `SetQuotaDate`, `SetReservedAmount`, and `StartConnection` are all inside the same `UoW.Do()`. If any step fails, the entire transaction rolls back, undoing the quota reserve.

3. **"CloseSession can release quota twice"** — False. `CanBeClosed()` rejects INACTIVE and FAILED sessions. The second call fails with `BadRequest("session is already closed")` before reaching quota release.

4. **"SetSessionFailed clears fields redundantly"** — By design. Clearing `quota_date` and `reserved_amount` is defense in depth. It prevents any future code path from releasing already-released quota.

5. **"The 2-minute threshold cancels in-progress sessions"** — False. The stale cleanup only runs during `CreateSession` (when user tries to create a new session). A session that has reached `StartConnection` has a `quota_date` and `speech_session_id` set. The cleanup only affects truly abandoned PENDING sessions.

6. **"GetCtx returns empty string for missing keys"** — This is a utility concern, not model-gateway specific. The context is populated by auth middleware before any service method is called. If it's missing, the auth layer fails first.
