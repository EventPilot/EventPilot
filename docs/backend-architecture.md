# Event Pilot — Backend Architecture Design

**Version:** 0.1  
**Last updated:** February 2025  
**Status:** Draft for review

---

## 1. Executive Summary

Event Pilot is a calendar and marketing agent for small engineering startups and student orgs. It schedules major milestones/events, prompts event stakeholders (owner, photographer, customers) for information and media after the event, uses an LLM to generate social posts, and publishes to X (Twitter) first, with LinkedIn and Instagram planned later.

This document describes the **backend architecture** for the API, integrations, and deployment—with a **Go API running serverless on Vercel** and **Supabase (PostgreSQL)** for persistence.

**Database vs. persistence:** In this architecture they are the same thing. The **database** is where we persist all durable state (events, prompts, responses, posts). There is no separate “persistence layer”—Supabase *is* the persistence layer. Caches (e.g. Redis) or file storage (e.g. for media) would be additive later.

---

## 2. Goals & Constraints

| Goal | Constraint |
|------|------------|
| Simple deployment | Serverless on Vercel (no long-lived servers) |
| Cost-conscious LLM usage | Prefer Claude ($50 free credits), fallback to OpenAI |
| Reliable storage | Supabase (PostgreSQL) for events, prompts, responses, and post drafts |
| Multi-role workflows | Prompt Owner, Photographer, Customers per event |
| Social distribution | Start with X API; extend to LinkedIn, Instagram |
| Local/dev consistency | Docker for local backend + dependencies |

---

## 3. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              VERCEL                                          │
│  ┌─────────────────────┐    ┌──────────────────────────────────────────┐   │
│  │   Next.js Frontend  │───▶│   Go API (Serverless Functions)           │   │
│  │   (App Router)      │    │   /api/events, /api/prompts, /api/posts   │   │
│  └─────────────────────┘    └───────────────────┬──────────────────────┘   │
└─────────────────────────────────────────────────┼───────────────────────────┘
                                                  │
                    ┌─────────────────────────────┼─────────────────────────────┐
                    │                             │                             │
                    ▼                             ▼                             ▼
            ┌───────────────┐             ┌───────────────┐             ┌───────────────┐
            │  Supabase     │             │  LLM Provider │             │  X (Twitter)  │
            │  (PostgreSQL) │             │  Claude /     │             │  API v2       │
            │  Events,      │             │  OpenAI       │             │  (post)       │
            │  Prompts,     │             │               │             │               │
            │  Responses,   │             │  - Prompt     │             │  (future:     │
            │  Posts        │             │    generation │             │   LinkedIn,    │
            │  + Storage    │             │  - Post       │             │   Instagram)  │
            │  (optional)   │             │    generation │             └───────────────┘
            └───────────────┘             └───────────────┘
                    │
                    │  (optional: Supabase Storage for media)
                    ▼
            ┌───────────────┐
            │  Supabase     │  Media uploads for event assets (optional)
            │  Storage      │
            └───────────────┘
```

- **Vercel** hosts the Next.js app and the Go serverless API (single project or monorepo).
- **Go API** is the only backend; it talks to Supabase (Postgres), the LLM, and X API. No separate “backend service” unless you later move off Vercel.
- **Docker** is used for local development (Go API + local Postgres or Supabase local dev) and for any future deployment (e.g., Lambda/ECS) if you leave Vercel.

---

## 4. Task Flow (Backend Perspective)

Your task breakdown maps to these backend flows:

| # | Task | Backend responsibility |
|---|------|-------------------------|
| 1 | **Database calendar → LLM** | Store events in Supabase; when “event done” or “request post,” load event + metadata and call LLM to generate **prompt text** (and optionally which role to ask). |
| 2 | **LLM → send prompt to users** | Persist “prompt requests” (per event, per role); expose endpoint so frontend/email can show “Event Pilot is asking you for input.” Optionally send email/link with prompt. |
| 3 | **LLM + user response → create post** | When user submits text/media links, store in Supabase; call LLM with event context + all collected responses to **generate post copy** (and optionally image selection). |
| 4 | **Post on X** | Call X API v2 to create a tweet (text + optional media). Store post ID and status in Supabase. |

End-to-end:

1. **Create event** → API writes to Supabase (calendar).
2. **Mark event “done”** (or cron/scheduled trigger) → API loads event, calls LLM to generate prompts per role (Owner, Photographer, Customers).
3. **Prompt delivery** → API creates “prompt request” records; frontend/notifications show “please submit your input” with the prompt text.
4. **User submits response** → API stores response (and optional media URLs) linked to event + role.
5. **Generate post** → When enough responses (e.g., owner + at least one other) or manual “generate,” API calls LLM with event + all responses → draft post stored.
6. **Publish to X** → API calls X API, stores tweet ID and status.

---

## 5. API Surface (Go on Vercel)

Vercel supports **Go serverless functions** via the Go runtime. Recommended route layout (Vercel / Next.js API or `/api` in same repo):

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/events` | List events (with optional filters: upcoming, past, by org). |
| `POST` | `/api/events` | Create event. |
| `GET` | `/api/events/:id` | Get single event. |
| `PATCH` | `/api/events/:id` | Update event (e.g. mark “done”). |
| `POST` | `/api/events/:id/request-prompts` | Trigger “event done” → LLM generates prompts per role; persist prompt requests. |
| `GET` | `/api/events/:id/prompts` | List prompt requests for an event (for owner dashboard). |
| `GET` | `/api/prompts/:id` | Get one prompt (e.g. for shared link “submit your input”). |
| `POST` | `/api/prompts/:id/response` | Submit user response (text + optional media URLs). |
| `POST` | `/api/events/:id/generate-post` | LLM generates post from event + responses; store draft. |
| `GET` | `/api/events/:id/post` | Get current draft or published post. |
| `POST` | `/api/events/:id/post/publish` | Publish draft to X (and store result). |

Authentication can be minimal at first (e.g. API key or simple JWT for “org” or “owner”), then refined (OAuth, Vercel Auth, or Cognito) later.

---

## 7. LLM Integration

- **Provider choice**: Prefer **Claude** (Anthropic) using the $50 free credits; **OpenAI** as fallback (e.g. when credits exhausted or for A/B).
- **Two LLM use cases**:
  1. **Prompt generation** (event → prompts per role): Input = event name, date, description, role. Output = one prompt text per role (and optionally instructions for media).
  2. **Post generation** (event + responses → post): Input = event summary + all response texts (and optional media list). Output = short post copy (e.g. tweet length) and optional image ordering.

Implementation in Go:

- Use official SDKs: `github.com/anthropics/anthropic-sdk-go` (or REST) and `github.com/sashabaranov/go-openai`.
- One small **provider abstraction** (interface) so you can switch or fallback:
  - `GeneratePrompts(ctx, event, roles) (map[role]string, error)`
  - `GeneratePost(ctx, event, responses) (copy string, mediaOrder []string, error)`
- Keep prompts in code or in a small config layer (no need to store prompt templates in the DB in v1).

---

## 8. X (Twitter) API

- Use **X API v2** (OAuth 2.0 or 1.0a as per X’s current requirements).
- **Create Tweet** endpoint: text + optional media IDs (upload media first if you support images).
- Store in Supabase: `xTweetId`, `publishedAt`, and optionally `xMediaIds`.
- Rate limits: design for “one post per event” and optional retry/backoff in Go.

Later: LinkedIn and Instagram as separate adapters (same “publish” abstraction: `Publish(ctx, post, platform) (id string, err error)`).

---

## 9. Vercel + Go Serverless

- **Vercel Go runtime**: Each handler is a serverless function. Cold starts for Go are acceptable; keep handlers thin and reuse a shared Supabase/HTTP client where possible.
- **Project layout** (example):
  - `frontend/` — Next.js (or root if monorepo).
  - `api/` — Go handlers; each file or subpath maps to a route (see Vercel docs for Go).
- **Secrets**: Store in Vercel env (e.g. `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `DATABASE_URL` or `SUPABASE_URL` + `SUPABASE_SERVICE_ROLE_KEY` for Supabase, `X_API_*`). Use Vercel’s env per environment (dev/preview/prod).
- **Supabase**: From Vercel, use **pgx** (or Supabase Go client) with `DATABASE_URL` (Postgres connection string from Supabase dashboard). No AWS credentials needed for the database.

---

## 10. Docker & Local Development

- **Docker Compose** for local stack:
  - **Postgres** in Docker (official image) or **Supabase local** (`supabase start`) for table creation and dev data.
  - Optional: **LocalStack** if you later add S3.
  - **Go API**: either run `go run ./cmd/api` against local Postgres or run the same handlers under a small HTTP server (e.g. `net/http`) that mimics Vercel’s request shape.
- **Scripts**: `make dev` or `docker compose up` to start Postgres (or `supabase start`); `make run-api` to run the Go API locally. Next.js can point to `http://localhost:8080` for API in dev.

This keeps “Docker for infrastructure” without requiring Docker on Vercel; Vercel only runs the built serverless functions.

---

## 11. Security & Credentials

- **API auth**: Start with a simple API key or JWT (org-scoped). Later: OAuth or Vercel Auth.
- **Supabase**: Use `DATABASE_URL` or `SUPABASE_SERVICE_ROLE_KEY` with minimal privileges; optional Row Level Security (RLS) for multi-tenant isolation.
- **X / LLM keys**: Only in server-side env; never expose to the frontend.
- **Prompt links**: Use unguessable IDs for `GET /api/prompts/:id` so “submit response” links are shareable but not enumerable.

---

## 12. Future Considerations

- **Media**: If you need uploads (not just URLs), add **Supabase Storage** (or S3) with presigned URLs or a small upload API; store object keys in Postgres and pass them to the LLM/post flow.
- **Scheduling**: “Event done” can be manual (owner clicks “Request inputs”) or a cron (Vercel Cron) that finds events with `date <= today` and `status = scheduled`, then calls `request-prompts`.
- **Multi-channel**: Same post payload, multiple adapters (X, LinkedIn, Instagram); store per-platform post IDs in Supabase.
- **Moving off Vercel**: If you later move the Go API to AWS (Lambda + API Gateway or ECS + Docker), the same Go code and Supabase model can be reused; only the deployment and trigger (HTTP vs Lambda event) change.

---

## 13. Summary

| Layer | Choice |
|-------|--------|
| **Hosting** | Vercel (Next.js + Go serverless API) |
| **Database** | Supabase (PostgreSQL / relational) |
| **LLM** | Claude first, OpenAI fallback; thin abstraction in Go |
| **Social** | X API v2 first; LinkedIn/Instagram later |
| **Local dev** | Docker (Postgres or Supabase local) + Go API run locally |
| **Auth** | Simple API key/JWT → refine later |

The backend is a **single Go serverless API on Vercel** that implements the four tasks: calendar in Supabase → LLM-generated prompts → stored user responses → LLM-generated post → publish to X, with a clear path to more channels and optional AWS migration later.
