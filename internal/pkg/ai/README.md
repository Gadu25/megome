# AI Package

Provides AI-powered content generation using Google Gemini.

## Setup

Add to your `.env`:

```
GEMINI_API_KEY=<key from aistudio.google.com/apikey>
GEMINI_MODEL=gemini-flash-latest
GEMINI_QUOTA_COOLDOWN=75
```

## Components

### `Provider` (`client.go`)

Interface with a single method:

```go
GenerateText(ctx context.Context, prompt string) (string, error)
```

### `GeminiClient` (`gemini.go`)

REST client for the Gemini API. Calls `generativelanguage.googleapis.com` with the API key sent via `X-goog-api-key` header. Handles 429 rate limiting with retry delay parsing from `RetryInfo`.

### `StatusTracker` (`status.go`)

Thread-safe tracker that manages quota cooldowns. When Gemini returns 429, marks the service as unavailable for a configurable duration (or the `retryDelay` from the response). Disabled entirely when no API key is configured.

### `Service` (`service.go`)

Orchestrates prompt building, text generation, and JSON parsing. Returns structured fields or typed errors (`UnavailableError`, `ErrUnknownTask`, `ErrGeneration`).

### `Prompts` (`prompts.go`)

Builds structured prompts for tasks: `generate_bio`, `generate_tagline`, `generate_project_description`, `generate_experience`, `generate_education`. Parses JSON from Gemini responses, stripping code fences and extracting the first JSON object.

## Task Types

| Task | Returns | Context Fields |
|------|---------|----------------|
| `generate_bio` | `{tagline, bio}` | title, tagline, location |
| `generate_tagline` | `{tagline}` | title, tagline, location |
| `generate_project_description` | `{description}` | title, githubLink |
| `generate_experience` | `{description}` | title, company, startDate, endDate, technologies |
| `generate_education` | `{description}` | school, degree, fieldOfStudy, startDate, endDate |
