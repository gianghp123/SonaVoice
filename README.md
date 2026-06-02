# 🎙️ Sona — AI Voice Companion

<p align="center" style="background-color: #ffffff;">
  <img src="docs/images/logo.png" alt="Sona Logo" width="300" />
</p>

<p align="center">
  <strong>Speak naturally. Learn confidently.</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/status-active-brightgreen" alt="Status" />
  <img src="https://img.shields.io/badge/license-Apache%202.0-blue" alt="License" />
  <img src="https://img.shields.io/badge/go-%3E%3D1.25-00ADD8?logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/python-%3E%3D3.12-3776AB?logo=python" alt="Python" />
  <img src="https://img.shields.io/badge/next.js-16-000000?logo=nextdotjs" alt="Next.js" />
</p>

---

> **Sona is under active development.** Core MVP is complete — staging environment is live at [sona-go-api-staging.vercel.app](https://sona-go-api-staging.vercel.app). See [Progress](#-roadmap--progress) for what's done and what's next.

---

## Why Sona?

I wanted to improve my English speaking skills. I tried existing AI voice tutors — they're either paywalled, or can't be customized to my needs. So I built my own.

Sona is an end-to-end **AI voice companion platform** that combines **speech-to-text**, a **large language model**, and **text-to-speech** into a WebRTC pipeline. It remembers conversation context across sessions via long-term memory.

This is also an **portfolio project** — showcasing full-stack development across Go, Python, Next.js, WebRTC, infrastructure-as-code, and CI/CD.

**Current persona:** English speaking teacher.

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Real-time voice chat** | Response pipeline: Deepgram STT → LLM → Piper TTS over WebRTC |
| **Persistent memory** | Mem0 + pgvector remembers conversation context across sessions |
| **User authentication** | Clerk provides sign-up, sign-in, and session management |
| **Session management** | Create, start, and review past voice sessions with full message history |
| **Rate limiting & quotas** | Configurable per-user rate limits and session duration caps |
| **Observability** | Sentry error tracking across all services; Zap structured logging (API) |
| **Infrastructure as code** | Terraform modules for Neon, Redis, Sentry, and Vercel |

---

## Architecture Overview

<div align="center">
  <img src="docs/images/infra.png" alt="Infrastructure Architecture" width="100%">
  <p><em>Figure 1. Infrastructure Architecture</em></p>
</div>

### Session Lifecycle

<div align="center">
  <img src="docs/images/session-flow.png" alt="Session Lifecycle" width="60%">
  <p><em>Figure 2. Session Lifecycle</em></p>
</div>


---

## Quickstart

```bash
# Clone the repo
git clone https://github.com/gianghp123/SonaVoice.git
cd SonaVoice

# Install all dependencies
make install

# Set up environment variables (see below)
cp services/api/.env.example services/api/.env
cp services/speech-engine/.env.example services/speech-engine/.env
cp web/.env.example web/.env

# Start all services
make dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

---

## Installation

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | >= 1.25 | API backend |
| [Python](https://www.python.org/downloads/) | >= 3.12 | Speech engine |
| [Node.js](https://nodejs.org/) | >= 20 | Web frontend |
| [PostgreSQL](https://www.postgresql.org/) | 15+ | Database (with pgvector extension) |
| [Redis](https://redis.io/) | 7+ | Caching & rate limiting |

### Services

```bash
# Web frontend
cd web && npm install

# API backend
cd services/api && go mod download

# Speech engine
cd services/speech-engine
pip install -r requirements.txt
# Download Piper voice model
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx.json
```

---

## Environment Variables

### Web (`web/.env`)

| Variable | Required | Description |
|----------|----------|-------------|
| `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` | Yes | Clerk publishable key |
| `CLERK_SECRET_KEY` | Yes | Clerk secret key |
| `API_URL` | Yes | Go API base URL (default: `http://localhost:8080`) |
| `SENTRY_DSN` | No | Sentry DSN for error tracking |

### API (`services/api/.env`)

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `REDIS_URL` | Yes | Redis connection string |
| `CLERK_SECRET_KEY` | Yes | Clerk secret key |
| `SPEECH_SERVICE_URL` | Yes | Speech engine URL (default: `http://localhost:7860`) |
| `MODE` | No | Gin mode: `debug` or `release` |
| `PORT` | No | Server port (default: `8080`) |
| `SENTRY_DSN` | No | Sentry DSN for error tracking |
| `REDIS_KEY_PREFIX` | No | Key prefix for Redis (e.g., `dev`) |
| `ENVIRONMENT` | No | Environment name |

### Speech Engine (`services/speech-engine/.env`)

| Variable | Required | Description |
|----------|----------|-------------|
| `DEEPGRAM_API_KEY` | Yes | Deepgram API key for STT |
| `GOOGLE_API_KEY` | Yes | Gemini/LLM API key |
| `OPENAI_BASE_URL` | No | OpenAI-compatible base URL |
| `DATABASE_URL` | Yes | PostgreSQL connection string (for Mem0) |
| `API_URL` | Yes | Go API base URL |
| `ICE_SERVERS` | No | JSON array of STUN/TURN server configs |
| `SENTRY_DSN` | No | Sentry DSN for error tracking |
| `MODAL_APP_NAME` | No | Modal app name (for deployment) |
| `MODAL_SECRET_NAME` | No | Modal secret name (for deployment) |
| `MODEL_ENVIRONMENT` | No | `local` or `modal` |

---

## Running Locally

### All services at once

```bash
make dev
```

This starts the speech engine, Go API, and Next.js frontend in parallel.

### Individual services

```bash
# Speech engine (port 7860)
make speech

# Go API (port 8080)
make api

# Next.js frontend (port 3000)
make web
```

### Local database & Redis

A minimal PostgreSQL container is provided for local development:

```bash
cd services/api/docker
docker compose up -d
```

> For full local dev (including Redis), a multi-service `docker compose` is planned — see [Roadmap](#-roadmap--progress).

---

## Deployment

### Infrastructure (Terraform)

Sona uses Terraform to manage infrastructure across environments:

```
infrastructure/
├── modules/
│   ├── database/    # Neon PostgreSQL (with pgvector)
│   ├── redis/       # Upstash Redis
│   ├── sentry/      # Sentry projects + DSNs
│   └── vercel/      # Vercel project + env vars
└── environments/
    ├── shared/      # Cross-environment resources (Redis)
    ├── dev/         # Development
    └── staging/     # Staging (auto-deployed from main)
```

```bash
cd infrastructure/environments/staging
terraform init
terraform plan
terraform apply
```

### Services

| Service | Platform | Notes |
|---------|----------|-------|
| Web frontend | **Vercel** | Next.js with Edge runtime |
| Go API | **Vercel** | Go runtime via `vercel.json` |
| Speech engine | **Modal** | CPU serverless containers |

### CI/CD

GitHub Actions handles:

- **PR checks** (`ci-workflow.yaml`): lint, build, and test all three services
- **Staging deploy** (`deploy-staging.yaml`): on merge to `main` — runs Terraform apply, syncs secrets to Modal, deploys the speech engine, and syncs env vars to Vercel

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Frontend** | Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS v4, shadcn/ui |
| **Auth** | Clerk |
| **Voice Client** | Pipecat AI Client JS + React + SmallWebRTC Transport |
| **API Backend** | Go, Gin, GORM, pgx |
| **Speech Pipeline** | Python, Pipecat AI, FastAPI |
| **STT** | Deepgram |
| **LLM** | OpenAI-compatible API (Deepseek via OpenCode proxy) |
| **TTS** | Piper (local onnx runtime) |
| **Memory** | Mem0 with pgvector extension |
| **Database** | Neon PostgreSQL |
| **Cache** | Upstash Redis |
| **Observability** | Sentry, Zap |
| **Infrastructure** | Terraform (Neon, Upstash, Sentry, Vercel providers) |
| **CI/CD** | GitHub Actions |
| **Deployment** | Vercel (web + API), Modal (speech engine) |

---

## Project Structure

```
sona-voice/
├── web/                          # Next.js frontend
│   ├── src/
│   │   ├── app/                  # App Router pages
│   │   │   ├── (auth)/           # Sign-in / sign-up
│   │   │   ├── (main)/           # Main app (chat, sessions, landing)
│   │   │   └── api/proxy/webrtc/ # WebRTC proxy to backend for client side
│   │   ├── components/           # Shared UI components
│   │   │   ├── common/           # AnalysisCard, LoadingScreen, Logo, etc.
│   │   │   └── ui/               # shadcn/ui primitives
│   │   ├── features/             # Feature modules
│   │   │   ├── chat-interface/   # VoiceOrb, VoicePanel, VoiceToolbar, etc.
│   │   │   ├── landing/          # LandingHero, ConnectNow
│   │   │   └── session-history/  # SessionMessageList, services
│   │   ├── hooks/                # Custom React hooks
│   │   ├── lib/                  # API client, types, utilities
│   │   └── instrumentation.ts    # Sentry setup
│   ├── package.json
│   └── next.config.ts
│
├── services/
│   ├── api/                      # Go API backend
│   │   ├── cmd/server/           # Entry point + Swagger docs
│   │   ├── internal/
│   │   │   ├── configs/          # Typed config (DB, Clerk, Redis, Sentry)
│   │   │   ├── core/             # Enums, errors, response envelope, logger
│   │   │   ├── database/         # GORM models, migrations, repos
│   │   │   ├── middlewares/      # Auth, role, rate limiter
│   │   │   ├── modules/
│   │   │   │   ├── session/      # Session CRUD + WebRTC orchestration
│   │   │   │   └── message/      # Message persistence
│   │   │   └── utils/            # Helpers
│   │   ├── go.mod
│   │   └── vercel.json
│   │
│   └── speech-engine/            # Python voice pipeline
│       ├── src/
│       │   ├── agents/           # System prompts + tools
│       │   ├── core/             # Pydantic config
│       │   ├── pipelines/        # STT/LLM/TTS factories, tasks, handlers
│       │   ├── services/         # API clients to Go backend
│       │   ├── types/            # Message types
│       │   └── utils/            # Error helpers
│       ├── main.py               # Bot entry point (Pipecat pipeline)
│       ├── runner.py              # Custom FastAPI WebRTC runner
│       ├── modal_app.py           # Modal serverless deployment
│       ├── requirements.txt
│       └── pyproject.toml
│
├── infrastructure/               # Terraform IaC
│   ├── modules/                  # Reusable modules
│   └── environments/             # Per-environment configs
│
└── .github/workflows/            # CI/CD pipelines
```

---

## Roadmap / Progress

### Completed

* [x] Real-time voice pipeline (STT → LLM → TTS over WebRTC)
* [x] User authentication (Clerk)
* [x] Session CRUD + message history
* [x] Mem0 long-term memory with pgvector
* [x] Rate limiting & user quotas
* [x] Sentry error tracking (all services)
* [x] Zap structured logging (API)
* [x] Terraform IaC (dev + staging)
* [x] CI/CD with auto-deploy to staging
* [x] Modal serverless deployment (speech engine)
* [x] Vercel deployment (web + API)
* [x] Unit tests for API backend

### In Progress / Planned

* [ ] Reduce end-to-end voice latency
* [ ] Production Terraform environment
* [ ] Frontend test suite
* [ ] Speech engine test suite
* [ ] E2E / integration tests
* [ ] Pronunciation feedback & scoring
* [ ] Speaking fluency analysis
* [ ] Vocabulary suggestions after conversation
* [ ] Grammar correction & explanation
* [ ] Personalized speaking recommendations
* [ ] Conversation summary + learning insights
* [ ] Adaptive difficulty by learner level
* [ ] Custom AI tutor personas
* [ ] Multi-language support
