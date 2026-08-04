# Settings Page Design

## Overview

Redesign the settings page from 6 mock tabs to 4 functional tabs: **Account**, **Security**, **API Keys**, and **Data**. Each tab needs both frontend UI and corresponding Go backend endpoints.

## Tabs

### 1. Account Tab

Editable identity fields (profile content lives on `/profile` page).

**Sections:**
- **Email** — text input pre-filled with current email. Submit requires current password for confirmation. Calls `POST /api/v1/auth/change-email`.
- **Username** — text input pre-filled with current username. Calls `POST /api/v1/auth/change-username`.
- **Danger Zone** — delete account button, requires password confirmation. Calls `DELETE /api/v1/auth/account`. Styled with error borders.

**Backend endpoints:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/change-email` | Update user email (requires current password) |
| POST | `/api/v1/auth/change-username` | Update username |
| DELETE | `/api/v1/auth/account` | Permanently delete user account and all data (requires password) |

### 2. Security Tab

**Sections:**
- **Change Password** — current password, new password, confirm new password. Calls `POST /api/v1/auth/change-password`.
- **Active Sessions** — list of non-expired, non-revoked refresh tokens showing creation date/IP/user agent, with per-session revoke button. Calls `GET /api/v1/auth/sessions` and `POST /api/v1/auth/sessions/:id/revoke`.

**Backend endpoints:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/change-password` | Change password (requires current password) |
| GET | `/api/v1/auth/sessions` | List active refresh tokens |
| POST | `/api/v1/auth/sessions/:id/revoke` | Revoke a specific refresh token session |

### 3. API Keys Tab

Already fully functional. No changes needed. Manages personal access tokens (generate, list, revoke, delete, view usage logs).

### 4. Data Tab

**Sections:**
- **Export Portfolio** — button to download all portfolio data as a single JSON file. Calls `GET /api/v1/data/export`.

**Backend endpoint:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/data/export` | Download all user data as structured JSON |

**Exported JSON structure:**
```json
{
  "exportedAt": "2026-08-04T...",
  "profile": { ... },
  "experiences": [ ... ],
  "skills": [ ... ],
  "education": [ ... ],
  "projects": [ ... ],
  "certifications": [ ... ]
}
```

## Frontend Changes

- Remove `OutputTab.tsx`, `IntegrationsTab.tsx`, `SecurityTab.tsx` (replaced)
- Create `AccountTab.tsx` (replace current placeholder)
- Create `SecurityTab.tsx` (replace current placeholder)
- Create `DataTab.tsx` (new)
- Update `page.tsx`: 4 tabs (account, security, api, data), remove old imports
- Update `index.ts` barrel export
- Update `README.md`

## Backend Changes

### New domain packages (if needed):
- None needed — all data accessed through existing user/profile refreshtoken repositories

### New handler files:
- `internal/api/handler/account.go` — change-email, change-username, delete-account handlers
- `internal/api/handler/security.go` — change-password, sessions list, session revoke handlers
- `internal/api/handler/dataexport.go` — export handler

### Changes to existing files:
- `internal/api/router.go` — register new routes
- `internal/domain/user/repository.go` — add UpdateEmail, UpdateUsername, DeleteAccount methods if missing
- `internal/domain/refreshtoken/repository.go` — add ListActiveSessions, RevokeSession methods if missing

### New API client functions (megome-front):
- `lib/api/client/account.ts` — changeEmail, changeUsername, deleteAccount
- `lib/api/client/security.ts` — changePassword, getSessions, revokeSession
- `lib/api/client/data.ts` — exportData
- `lib/api/client/server/account.ts` — server-side variants if needed

## Error Handling

- All mutation endpoints validate input via `go-playground/validator`
- Password-confirmed actions (change email, change password, delete account) return 401 if current password is wrong
- Username uniqueness is validated before update (return 409 on conflict)
- Account deletion cascades to all user-owned data (profiles, experiences, skills, education, projects, certifications, tokens)
- Data export returns 500 if any repository query fails

## Validation Rules

| Field | Rules |
|-------|-------|
| email | required, email format, unique |
| username | required, min 3, max 30, unique |
| currentPassword | required for email change / password change / account deletion |
| newPassword | required, min 8 characters |
| confirmPassword | must match newPassword |
