# AI Assist Design (Gemini Free Tier)

## Overview

Add AI-powered writing assistance to Megome using the **free tier of the Google Gemini API** (no token spend). A single generic assist endpoint in the Go backend generates draft content for five forms: profile (bio + tagline), project description, experience, education, and certificate. When the free-tier quota runs out (HTTP 429 / RESOURCE_EXHAUSTED), the backend enters an "unavailable" state with a cooldown; the frontend shows a global banner and disables all generate buttons until the cooldown clears.

Architecture chosen: **Approach A** — one generic `POST /api/v1/ai/assist` endpoint with task types, plus `GET /api/v1/ai/status`. Provider-agnostic client so an Ollama provider can be swapped in later without changing app code. No streaming in v1.

## Backend Changes (megome)

### New package: `internal/pkg/ai`

- `client.go` — `Provider` interface with `GenerateText(ctx, prompt) (string, error)`. Default implementation is the Gemini HTTP client. Uses direct REST calls (no SDK dependency).
- `gemini.go` — Gemini implementation calling `POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={apiKey}`. Parses `candidates[0].content.parts[].text`. Classifies errors: HTTP 429 or `RESOURCE_EXHAUSTED` → `*QuotaError`; everything else → generic error.
- `status.go` — in-memory status tracker with two states: `available` and `unavailable`. Holds an `unavailableUntil` timestamp. Methods: `MarkUnavailable()`, `Status()` returning `{available, cooldownRemainingSeconds}`.
- `prompts.go` — one prompt builder per task (see Tasks below). Each takes the context fields + optional `extra` notes and returns the prompt string. Prompts instruct Gemini to return plain text only, no markdown.

### New package: `internal/domain/assist`

Shared types:
- `Task` string constant set: `generate_bio`, `generate_tagline`, `generate_project_description`, `generate_experience`, `generate_education`, `generate_certificate`.
- `Request { Task string, Context map[string]string, Extra string }`
- `Response { Task string, Fields map[string]string }` — `Fields` so `generate_bio` can return both `bio` and `tagline`.

### New handler: `internal/api/handler/assist.go`

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/ai/assist` | JWT-protected. Validate task, build prompt from context, call provider, return `{task, fields}`. |
| GET | `/api/v1/ai/status` | JWT-protected. Returns `{available, cooldownRemainingSeconds}`. |

Assist flow:
1. Validate task is known and required context present → 400 otherwise.
2. If status is unavailable → 429 with `{message: "ai_unavailable", data: {cooldownRemainingSeconds}}`.
3. Build prompt, call `Provider.GenerateText`.
4. On `*QuotaError` → `MarkUnavailable()` and return 429 `{message: "ai_unavailable"}`.
5. On other errors → 502 `{message: "ai_generation_failed"}`.
6. On success → 200 `{message, data: {task, fields}}`.

### Changes to existing files

- `internal/api/router.go` — register assist handler routes on the `internal` router.
- `internal/config/config.go` — add envs:

| Env | Default | Purpose |
|-----|---------|---------|
| `GEMINI_API_KEY` | empty | API key (optional; empty disables AI → always unavailable) |
| `GEMINI_MODEL` | `gemini-2.5-flash` | Model name |
| `GEMINI_QUOTA_COOLDOWN` | `1800` (seconds) | Cooldown after a quota error |

## Frontend Changes (megome-front)

### New proxy route handlers (mirror existing `/app/api/*` pattern)
- `app/api/ai/assist/route.ts` — POST → backend `/api/v1/ai/assist`, forwards `Authorization: Bearer`.
- `app/api/ai/status/route.ts` — GET → backend `/api/v1/ai/status`.

### New client + store
- `lib/api/client/assist.ts` — `assistClient(task, context, extra)`, `getAiStatusClient()`.
- `lib/store/ai-status-store.ts` — zustand store: `{ status, fetchStatus() }`. Fetched once in the `(app)` layout.

### New components (`features/ai/`)
- `AiStatusBanner.tsx` — rendered under the Navbar in `app/(app)/layout.tsx` when `available === false`. Text: "AI assist is temporarily unavailable due to quota. Try again later."
- `AiAssistButton.tsx` — reusable. Props: `task`, `context` (record of field values), `extra` placeholder, `onResult(fields)`. Renders "✨ Generate" button with a collapsible "extra notes" textarea. Disabled when status unavailable or loading. Handles 429 by refreshing the status store (flips global banner) and showing a toast.

### Wiring into forms

| Form | Task | Fields filled |
|------|------|---------------|
| `ProfileForm.tsx` | `generate_bio` | `bio`, `tagline` (RichEditor + tagline input) |
| `ProjectWizard` `stepInfo.tsx` | `generate_project_description` | `description` |
| `ExperienceForm.tsx` | `generate_experience` | `description` |
| `EducationForm.tsx` | `generate_education` | `description` |
| `CertificateForm.tsx` | `generate_certificate` | `description` |

Context sent per form: the form's current field values (e.g. title, company, location, dates) plus any `extra` notes typed by the user. Generated text is a **draft** — the user reviews and edits before saving.

## Tasks & Prompts

| Task | Context keys (examples) | Output |
|------|------------------------|--------|
| `generate_bio` | title, location, extra notes | fields `bio` (short paragraph) + `tagline` (one line) |
| `generate_tagline` | title, extra notes | field `tagline` |
| `generate_project_description` | title, technologies, extra notes | field `description` |
| `generate_experience` | role, company, location, startDate, endDate, extra notes | field `description` |
| `generate_education` | degree, institution, fieldOfStudy, startDate, endDate, extra notes | field `description` |
| `generate_certificate` | name, issuer, date, extra notes | field `description` |

Prompt builder convention: a single `buildPrompt(task, context, extra)` function switching per task. Output must be plain text, max ~200 words, no markdown, first-person voice for profile/bio tasks.

## Error Handling & Quota Flow

1. Gemini returns 429/RESOURCE_EXHAUSTED on any assist call → backend flips to unavailable (cooldown default 30 min), returns 429.
2. Frontend receives 429 → updates status store → global banner appears, all `AiAssistButton`s disable.
3. After cooldown elapses, status flips back to available on the next status poll/assist call; banner hides, buttons re-enable.
4. Non-quota errors (timeouts, 5xx, malformed response) → inline toast only; status stays available.
5. Empty `GEMINI_API_KEY` → status always unavailable (feature shows as disabled without errors).

## Testing

### Backend (Go)
- Unit tests for `internal/pkg/ai/status.go` (state transitions, cooldown remaining).
- Unit tests for `internal/pkg/ai/prompts.go` (each task builds a non-empty prompt containing context).
- Unit tests for `internal/pkg/ai/gemini.go` using `httptest` server: successful parse, 429 → `QuotaError`, `RESOURCE_EXHAUSTED` body → `QuotaError`, 500 → generic error.
- Handler test for assist: unknown task → 400; quota error → 429 + status unavailable.

### Frontend
- No test framework present. Verify with `npm run lint` and `npm run build`.
- Manual QA checklist: generate from each of the 5 forms; banner appears after simulated quota (set key empty or cooldown to 1s); buttons disable; toast on non-quota error.

## Non-Goals (v1)

- Streaming/SSE responses (future enhancement).
- GitHub README import for project descriptions.
- Persistent quota tracking across restarts (in-memory status resets on restart).
- Multi-provider selection UI (provider chosen at build via config only).
