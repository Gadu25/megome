# Polish What Exists — Production Polish for Megome

**Date:** 2026-08-06
**Status:** draft
**Scope:** Backend foundation + frontend UX/DX polish across existing features

## 1. Backend Foundation

### 1.1 Pagination on all list endpoints

Only `/api-logs/token/{id}` supports pagination. Every other list endpoint returns all rows.

Add `limit` (default 20, max 100) and `offset` (default 0) query params to:
projects, skills, education, experience, certifications, technologies, PATs — both internal and public endpoints.

Response envelope:
```json
{
  "data": [...],
  "pagination": { "limit": 20, "offset": 0, "total": 47 }
}
```

Each repository `GetByUserID` gets `LIMIT ? OFFSET ?` + a `CountByUserID` companion.

### 1.2 Soft delete with `deleted_at`

Add `deleted_at TIMESTAMP NULL DEFAULT NULL` to: projects, skills, education, experiences, certifications, project_images, project_techs, experience_tech.

Delete handlers write `deleted_at = NOW()` instead of `DELETE`. All queries add `AND deleted_at IS NULL`. R2 objects left in place.

Excluded: users, refresh_tokens, personal_access_tokens (existing revoke patterns), api_usage_logs (append-only), password_reset_tokens (expiry), technologies (seed data), profiles (1:1 user).

## 2. Frontend Error Handling

- `app/error.tsx` — branded "Something went wrong" + retry
- `app/not-found.tsx` — branded 404 with dashboard/docs/login links
- `components/ui/ErrorBoundary.tsx` — class component wrapper for profile sections and project list

## 3. Dashboard Overhaul

- Move API playground to `/api` docs section
- Recent activity timeline — unions entity tables by created_at
- API usage bar chart — daily counts from api_usage_logs
- Keep stat cards + completion checklist

## 4. Form Polish

- Unsaved changes: `useDirtyGuard` hook with `beforeunload` + confirmation modal
- ProfileForm: 1s debounced autosave with "Saved" indicator
- Rich text preview toggle on bio, experience, and education editors

## 5. Image Handling

- `<canvas>`-based `resizeImage()` — 1920px covers/screenshots, 800px logos, 400px avatars, 0.85 JPEG
- Drag-and-drop on cover/screenshot drop zones
- Minimal circular avatar crop modal
- 10MB max, image-only file validation

## 6. Empty State Consistency

- Project list uses shared `EmptyState` component
- Form sections use `EmptyState` instead of plain text

## Files Changed

### Backend (megome)
- `internal/domain/*/repository.go` — pagination + soft delete (7 entities)
- `internal/domain/*/model.go` — PaginatedResponse type, extended interfaces
- `internal/api/handler/dashboard.go` — activity + usage stats endpoints
- `internal/api/handler/*.go` — limit/offset parsing (7 handlers)
- `internal/api/router.go` — new dashboard routes
- `cmd/migrate/migrations/` — soft-delete migrations (7 up/down pairs)

### Frontend (megome-front)
- `app/error.tsx`, `app/not-found.tsx` — new
- `components/ui/ErrorBoundary.tsx` — new
- `app/(app)/dashboard/page.tsx` — reworked
- `features/profile/components/ProfileForm.tsx` — autosave + dirty guard + preview
- `features/project/components/stepperForm/StepImages.tsx` — resize + drag-drop + validation
- `features/project/components/ProjectsClient.tsx` — empty state
- `features/profile/components/*Form.tsx` — empty states
- Profile section components — ErrorBoundary wrappers
- New files: `useDirtyGuard`, `useImageResize`, `RichTextPreview`, `ActivityTimeline`, `UsageChart`, `AvatarCrop`
- New API proxy routes for dashboard activity/usage-stats

## Risks

| Risk | Mitigation |
|---|---|
| Soft-delete ALTER TABLE on existing data | Safe (NULL default), test on dev |
| Pagination breaks frontend assuming full arrays | Fallback to current behavior if no params |
| Image resize slow on mobile for large files | Debounced, show progress |
| Error boundaries mask bugs in dev | Only active in production |
