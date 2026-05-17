# Model Gateway Service — Flow Documentation

**File:** `services/model-gateway.service.go`
**Interface:** `IModelGatewayService`

---

## Overview

`modelGatewayService` is the orchestrator for the voice session lifecycle. It coordinates 4 sub-services:

| Service | Role |
|---|---|
| `IGlobalConfigService` | Reads the JSONB global config payload |
| `ISessionService` | CRUD + state transitions on `Session` DB model |
| `ISpeechProxyService` | HTTP proxy to external speech engine |
| `IUserQuotaRepository` | Daily voice-second quota (reserve/release via Postgres) |
| `UnitOfWork` | Transactional boundary for atomic DB operations |

---

## Session State Machine

```
         CreateSession / ResumeSession
                    │
                    ▼
              ┌──────────┐
              │  PENDING  │
              └────┬─────┘
                   │
            ProxyOffer (success)
                   │
                   ▼
              ┌──────────┐
              │  ACTIVE   │
              └────┬─────┘
                   │
            CloseSession / ProxyOffer (fail)
                   │
                   ▼
              ┌──────────┐
              │ INACTIVE │
              └──────────┘

   PENDING ──(timeout/error)──► FAILED
   ACTIVE  ──(error)─────────► FAILED
```

---

## Method Flows

### 1. `CreateSession(ctx) → (*res.CreateSessionRes, error)`

Entry point: `POST /sessions`

```
1. Extract requesterID from context (JWT/Clerk)
2. GlobalConfigService.Get()                     → typed config (daily quota, TTL)
3. UnitOfWork.Do() [transaction]:
   a. AcquireLock(requesterID)                   → PG advisory lock (prevent race)
   b. FindStaleByUserID(requesterID, TTL)        → stale pending/active sessions
      For each stale session:
        i.   domain.NewSessionFromModel()
        ii.  ShouldReleaseQuota()? → releaseQuotaInTx() + mark QuotaReleased
        iii. UpdateStatus() → INACTIVE
   c. FindActiveByUserID(requesterID)            → conflict check
      If active exists → return Conflict error
   d. Create(&Session{UserID, Status=PENDING})  → new DB row
4. startOrResume(session, requesterID, dailyQuota):
   a. QuotaRepo.ReserveAll(requesterID, "voice", today, dailyQuota)
      └─ if 0 returned → Forbidden("quota exceeded")
   b. SessionService.SetReservation(reservedAmount, dailyQuota)
      └─ on fail → releaseQuota() (rollback)
   c. connectToSpeech(sessionID, reservedAmount, requesterID):
        i.   SpeechProxyService.StartConnection({UserID, SessionID, MaxDuration=reserved})
        ii.  SessionService.SetSpeechSessionID(speechSessionID)
        └─ on fail → releaseQuota() + MarkSessionFailed() + MarkQuotaReleased()
   d. Return CreateSessionRes{ID, MaxDuration, WebRTCConnectionRes}
```

### 2. `ResumeSession(ctx, sessionID) → (*res.CreateSessionRes, error)`

Entry point: `POST /sessions/:sessionId/resume`

```
1. Extract requesterID from context
2. SessionService.GetInternal(sessionID)           → raw session (no ownership check)
3. domain.NewSessionFromModel(session).CanBeResumedBy(requesterID)
   └─ validates: owner match + Status == INACTIVE
4. GlobalConfigService.Get()                       → typed config
5. UnitOfWork.Do() [transaction]:
   a. AcquireLock(requesterID)
   b. FindStaleByUserID + cleanup (same as CreateSession)
   c. FindActiveByUserID → conflict if active.ID != sessionID
   d. UpdateStatus(session.ID) → PENDING
6. startOrResume(session, requesterID, dailyQuota)
   └─ (same as CreateSession — reserves quota, connects to speech engine)
```

### 3. `ProxyOffer(ctx, sessionId, method, body) → ([]byte, int, error)`

Entry point: `POST|PATCH /sessions/:sessionId/api/offer`

```
1. Validate sessionId is not empty
2. Extract requesterID from context
3. SessionService.GetBySpeechSessionID(speechSessionID, requesterID)
   └─ ownership enforced
4. SpeechProxyService.ProxyOffer(speechSessionID, method, body)
   └─ forwards HTTP request to speech engine
5. On failure OR non-2xx response:
   a. MarkSessionFailed()
   b. releaseQuota(session.UserID, reservedAmount, 0)
   c. MarkQuotaReleased()
6. On success (2xx):
   a. MarkSessionActive() → Status = ACTIVE, set started_at
7. Return (responseBody, statusCode, nil)
```

### 4. `CloseSession(ctx, reqBody) → error`

Entry point: `POST /sessions/:sessionId/close`

```
1. Validate reqBody (not nil, sessionId non-empty, actualUsage >= 0)
2. UnitOfWork.Do() [transaction]:
   a. SessionRepo.Get(sessionId)
   b. domain.NewSessionFromModel(session).CanBeClosed()
      └─ validates Status != INACTIVE
   c. ShouldReleaseQuota()?
        └─ reserveAmount/dailyQuota fallback (load config if both <= 0)
        └─ quotaRepo.Release(UserID, "voice", today, unused)
           └─ unused = max(0, reserved - clampedActualUsage)
        └─ sessionRepo.UpdateQuotaReleased()
   d. sessionRepo.UpdateStatus() → INACTIVE
```

---

## Quota Lifecycle

```
Reserve (startOrResume)
   │
   ├── Success ──► Connect to speech engine
   │                  │
   │                  ├── Offer success ──► Active session (quota consumed)
   │                  │
   │                  ├── Offer failure ──► Release all quota
   │                  │
   │                  └── Close session ──► Release unused = max(0, reserved - actualUsed)
   │
   └── Failure ──► Return error, quota not reserved
```

---

## Error Handling Patterns

| Situation | Response |
|---|---|
| Quota exhausted | `Forbidden("quota exceeded")` |
| Active session exists | `Conflict("close current session...")` |
| Session not resumable | `BadRequest("session is not resumable")` |
| Session already inactive | `BadRequest("session is already inactive")` |
| Not session owner | `Forbidden()` |
| DB/Infra error | `Internal()` (logged with details) |

---

## Key Design Decisions

- **Advisory lock** (`AcquireLock`): Uses `pg_advisory_xact_lock(hashtext(userID))` to serialize session creation per user — prevents race conditions when two create requests arrive simultaneously.
- **Stale session cleanup**: Runs inside every CreateSession/ResumeSession transaction. Sessions past `maxSessionLockTTL` in PENDING, or past their duration in ACTIVE, are force-closed.
- **Quota in transaction vs out**: `startOrResume` runs quota outside the transaction (it calls the speech engine), so `releaseQuota` is needed as rollback on failure. `CloseSession` runs quota release *inside* the transaction for atomicity.
- **ID as `speechSessionID`**: The external speech engine's session ID is stored on the `Session` model. `ProxyOffer` looks up by this ID, not by the internal session ID.
