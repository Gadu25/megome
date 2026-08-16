# OTP Email Verification Design

## Overview

Add email verification to the sign-up flow. When a user registers with email/password, the backend creates the account in an **unverified** state, emails a 6-digit OTP, and blocks login until the email is verified. The user enters the OTP on a dedicated frontend page and is **auto-logged-in** on success. A resend option with a 60-second cooldown is included.

The implementation follows the existing password-reset pattern exactly (`passwordforgot` domain, `mailer` package with HTML templates, handler wiring in `internal/api/handler/user.go`).

Decisions locked in with the user:
- **Flow:** account is created but pending until email verified; login is blocked for unverified accounts.
- **OTP:** 6-digit numeric, 10-minute expiry, resend allowed with a server-enforced 60s cooldown.
- **Existing users + Google OAuth:** both auto-verified. Existing users are backfilled via migration; Google users are marked verified at creation (Google already verified the email).
- **After verify:** auto-login — the verify endpoint issues access/refresh tokens.

## Backend Changes (megome)

### Migration: `cmd/migrate/migrations/20260817000001_add-email-verification.up.sql`

- `ALTER TABLE users ADD COLUMN emailVerifiedAt DATETIME NULL`
- Backfill existing rows: `UPDATE users SET emailVerifiedAt = NOW() WHERE emailVerifiedAt IS NULL`
- Create `email_verification_otps` table (mirrors `password_reset_tokens`):

```sql
CREATE TABLE email_verification_otps (
	id INT UNSIGNED NOT NULL AUTO_INCREMENT,
	userId INT UNSIGNED NOT NULL,
	email VARCHAR(255) NOT NULL,
	otpHash CHAR(64) NOT NULL,
	expiresAt DATETIME NOT NULL,
	usedAt DATETIME NULL,
	createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

	PRIMARY KEY (id),
	FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE
)
```

- `.down.sql` drops the table and removes the `emailVerifiedAt` column.

### Mailer

- New template `internal/pkg/mailer/templates/verify_email.html` styled like `reset_password.html` (card, header, footer, `.Year`). Shows the 6-digit OTP in a prominent code block and notes the 10-minute expiry.
- `service.go` additions:
  - `VerifyEmailData { OTP string; ExpiresInMinutes int; Year int }`
  - `SendVerifyEmail(to string, otp string) error` — renders `verify_email.html` and sends with subject `Verify your email`.

### New domain: `internal/domain/emailverification`

Mirror of `internal/domain/passwordforgot`:

- `model.go` — `EmailVerificationStore` interface:
  - `SaveOTP(userId int, email string, otpHash string, exp time.Time) error`
  - `LastOTPSentAt(userId int) (time.Time, error)` — for the resend cooldown
  - `DeleteOTPs(userId int) error` — removes all OTP rows (used by resend and by `MarkVerified`)
  - `VerifyOTP(email string, otpHash string) (int, error)` — returns userId, rejects invalid/expired/used codes
  - `MarkVerified(userId int) error` — in a transaction: sets `users.emailVerifiedAt = NOW()` and calls `DeleteOTPs(userId)`
- `repository.go` — OTP is hashed with SHA-256 before storage (same approach as `password_reset_tokens`). `VerifyOTP` scans by `email + otpHash`, enforces `usedAt IS NULL` and `expiresAt > NOW()`.

### OTP generation

- `internal/pkg/auth`: add `GenerateOTP() (string, error)` — 6-digit numeric via `crypto/rand`.

### Handlers: `internal/api/handler/user.go`

- `handleRegister` (modified): after `CreateUser`, generate + hash + `SaveOTP`, then `SendVerifyEmail`. Responds `201 { success: true, message: "verification code sent to your email", email }`. **No tokens issued at registration.**
- `handleVerifyEmail` (new, `POST /auth/verify-email`): payload `{ email, otp }`. Validates, calls `VerifyOTP` + `MarkVerified`, then `getTokens(userId)` and responds with the standard `AuthResponse` (tokens) — auto-login.
- `handleResendOTP` (new, `POST /auth/resend-otp`): payload `{ email }`. Loads user; 404 if missing; 400 if already verified. Enforces 60s cooldown via `LastOTPSentAt` (409 with a `retryAfterSeconds` value otherwise). Calls `DeleteOTPs`, saves a fresh OTP, sends the email.
- `handleLogin` (modified): after password check, if `emailVerifiedAt IS NULL` → `403 { error: "email not verified", email: <user email> }` so the frontend can route the user to verification.
- Google OAuth creation (`handleGoogleCallback`): set `emailVerifiedAt = NOW()` when creating a new user via OAuth.

Payload additions in `internal/domain/user/model.go`:
- `VerifyEmailPayload { Email string; OTP string }`
- `ResendOTPPayload { Email string }`

### Routes

- Register `verify-email` and `resend-otp` in `UserHandler.RegisterRoutes`.

## Frontend Changes (megome-front)

### API proxy routes (mirror existing `/app/api/auth/*` pattern)

- `app/api/auth/register/route.ts` — **stop setting cookies** (register no longer returns tokens). Include `email` in the success response so the page can navigate to verification.
- `app/api/auth/verify-email/route.ts` — new; POST → backend `/api/v1/auth/verify-email`. On success, set `access_token`/`refresh_token` httpOnly cookies (same as `login/route.ts`) and return `{ success, user }`.
- `app/api/auth/resend-otp/route.ts` — new; POST → backend `/api/v1/auth/resend-otp`, passthrough response.

### Client

- `lib/api/client/auth.ts`: add `verifyEmailClient(email, otp)` and `resendOtpClient(email)` using `handleResponse`.

### UI

- `features/auth/components/AuthForm.tsx` — `handleRegister`: on success, toast the message, then `router.push("/auth/verify-email?email=<email>")` instead of fetching init + redirecting to dashboard.
- `features/auth/schema.ts` — add `verifyEmailSchema` (`email` valid, `otp` exactly 6 digits).
- New page `app/(auth)/auth/verify-email/page.tsx` modeled on `forgot-password/page.tsx`:
  - Header, email shown (prefilled from `?email=` query, editable), 6-digit OTP input, submit button.
  - Resend button with 60s client countdown; disabled during cooldown; surfaces server errors via `withRequest`.
  - "Back to Sign In" link.
  - Success state after verification, then redirect to `/profile-setup` (if no profile via `getInitClient`) or `/dashboard`.

### Login flow for unverified users

- When login returns `403 "email not verified"` (with `email` in the payload), `AuthForm` redirects to `/auth/verify-email?email=<email>` so the user can complete verification. Errors from the request are still surfaced via `withRequest` for other failures.

## Error Handling

- Backend errors flow through `httputil.WriteError` → `{ error: "..." }`; proxy routes map to `{ success: false, message }`; client throws via `handleResponse`; UI surfaces via `withRequest`.
- Specific cases: wrong/expired OTP → 400 with clear message; resend too soon → 409 with `retryAfterSeconds`; login while unverified → 403 with the user's `email`.

## Security

- OTP stored only as a SHA-256 hash; never returned or logged.
- Single-use: marked used on successful verification; all OTP rows cleared after verification and on resend (only one active OTP per user).
- 10-minute expiry, 60-second resend cooldown enforced server-side.
- Unverified accounts cannot log in.
- OTP is bound to the email it was sent to (`email` column + lookup by `email + otpHash`).

## Testing / Verification

- Backend: `go build ./...` and `make test`.
- Frontend: `npm run lint`, `npx tsc --noEmit`, `npm run build`.
- Manual: register → receive OTP → wrong code rejected → correct code logs in and lands on dashboard/profile-setup; resend respects cooldown; login blocked before verification; Google OAuth unaffected.
