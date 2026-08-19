# OTP Email Verification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add email verification to sign-up so new accounts are created unverified, receive a 6-digit OTP by email, and can only log in after entering it (with auto-login on verify and a 60s resend cooldown).

**Architecture:** Mirrors the existing password-reset feature end to end. Backend: new `email_verification_otps` table + `users.emailVerifiedAt` column, a `verify_email.html` mailer template, a `GenerateOTP` helper in `internal/pkg/auth`, a new `emailverification` domain (SHA-256 hashed OTPs, expiry, cooldown, mark-verified), and three touched handlers in `internal/api/handler/user.go` (register sends OTP, new `verify-email` issues tokens, new `resend-otp`, login blocks unverified). Frontend: register no longer sets cookies, a new verify-email page with a resend countdown, new proxy routes + client functions following the existing auth patterns.

**Tech Stack:** Go 1.25 (gorilla/mux, database/sql, gomail, golang-migrate), Next.js 16 / React 19 / TypeScript / Tailwind 4 + daisyUI / zod.

**Spec:** `docs/superpowers/specs/2026-08-17-otp-email-verification-design.md`

## Global Constraints

- Backend repo: `/home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification`; frontend repo: `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification`. Commit in the repo each task belongs to.
- Go module is `megome`; imports use `megome/...` prefixes.
- No new Go or npm dependencies. OTP generation uses `crypto/rand`; hashing uses `crypto/sha256` (same as `password_reset_tokens`).
- Follow the existing password-reset pattern exactly: `internal/domain/passwordforgot` (repository), `internal/pkg/mailer` (template + service), handlers in `internal/api/handler/user.go`.
- OTP is 6 numeric digits, 10-minute expiry, 60-second resend cooldown. Cooldown and expiry constants live in the `emailverification` package (`OTPExpiry`, `ResendCooldown`).
- OTP is stored ONLY as a SHA-256 hex hash. Never return or log the plaintext OTP.
- Unverified accounts cannot log in (`403 {error: "email not verified", email}`). Google OAuth users and pre-existing users are auto-verified.
- The `verify-email` endpoint performs auto-login (issues access + refresh tokens via the existing `getTokens`).
- Backend verify command: `go build ./... && go vet ./... && go test ./...` (run in `/home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification`).
- Frontend verify command: `npm run lint && npm run build` (run in `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification`). Lint MUST exit 0 (zero errors); warnings are tolerated.
- Pre-existing lint debt (15 errors) is in scope and fixed by Task 11.
- Migration timestamp for this feature: `20260817000001` (must be lexically after existing `20260806000007`).

---

### Task 1: Migration — `users.emailVerifiedAt` + `email_verification_otps` table

**Files:**
- Create: `cmd/migrate/migrations/20260817000001_add-email-verification.up.sql`
- Create: `cmd/migrate/migrations/20260817000001_add-email-verification.down.sql`

**Interfaces:**
- Consumes: nothing.
- Produces: `users.emailVerifiedAt DATETIME NULL` column (backfilled to `NOW()` for existing rows) and table `email_verification_otps` with columns `id, userId, email, otpHash (CHAR(64)), expiresAt, usedAt, createdAt`.

- [ ] **Step 1: Create the up migration**

`cmd/migrate/migrations/20260817000001_add-email-verification.up.sql`:

```sql
ALTER TABLE users ADD COLUMN emailVerifiedAt DATETIME NULL;

UPDATE users SET emailVerifiedAt = NOW() WHERE emailVerifiedAt IS NULL;

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

- [ ] **Step 2: Create the down migration**

`cmd/migrate/migrations/20260817000001_add-email-verification.down.sql`:

```sql
DROP TABLE IF EXISTS email_verification_otps;

ALTER TABLE users DROP COLUMN emailVerifiedAt;
```

- [ ] **Step 3: Verify**

Run: `cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && go run cmd/migrate/main.go up`
Expected: applies the new migration (requires a running MySQL with `.env` configured; if no DB is available, review both files for correctness and continue — `make migrate-up` will apply it later).

- [ ] **Step 4: Commit**

```bash
cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && git add cmd/migrate/migrations/20260817000001_add-email-verification.up.sql cmd/migrate/migrations/20260817000001_add-email-verification.down.sql && git commit -m "feat: add email verification migration (users.emailVerifiedAt + otps table)"
```

---

### Task 2: OTP generator in `internal/pkg/auth`

**Files:**
- Create: `internal/pkg/auth/otp.go`
- Test: `internal/pkg/auth/otp_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: `GenerateOTP() (string, error)` — returns a 6-digit numeric string using `crypto/rand`.

- [ ] **Step 1: Write the failing test**

`internal/pkg/auth/otp_test.go`:

```go
package auth

import "testing"

func TestGenerateOTPLength(t *testing.T) {
	for i := 0; i < 100; i++ {
		otp, err := GenerateOTP()
		if err != nil {
			t.Fatal(err)
		}
		if len(otp) != 6 {
			t.Errorf("expected 6-digit code, got %q", otp)
		}
	}
}

func TestGenerateOTPNumeric(t *testing.T) {
	for i := 0; i < 100; i++ {
		otp, err := GenerateOTP()
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range otp {
			if c < '0' || c > '9' {
				t.Errorf("non-digit character %q in %q", c, otp)
			}
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && go test ./internal/pkg/auth/...`
Expected: FAIL with `undefined: GenerateOTP`.

- [ ] **Step 3: Write the implementation**

`internal/pkg/auth/otp.go`:

```go
package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateOTP returns a 6-digit numeric one-time password.
func GenerateOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && go test ./internal/pkg/auth/...`
Expected: PASS (`ok megome/internal/pkg/auth`).

- [ ] **Step 5: Commit**

```bash
cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && git add internal/pkg/auth/otp.go internal/pkg/auth/otp_test.go && git commit -m "feat(auth): add crypto-random 6-digit OTP generator"
```

---

### Task 3: `verify_email.html` template + mailer service method

**Files:**
- Create: `internal/pkg/mailer/templates/verify_email.html`
- Modify: `internal/pkg/mailer/service.go`
- Test: `internal/pkg/mailer/service_test.go`

**Interfaces:**
- Consumes: `mailer.Renderer` (existing), `mailer.Email` (existing).
- Produces: `VerifyEmailData{OTP string; ExpiresInMinutes int; Year int}`; `(*Service).SendVerifyEmail(to string, otp string) error` rendering `verify_email.html` with subject `Verify your email`.

- [ ] **Step 1: Write the failing test**

`internal/pkg/mailer/service_test.go`:

```go
package mailer

import (
	"strings"
	"testing"
)

func TestVerifyEmailTemplateRenders(t *testing.T) {
	r, err := NewRenderer("templates/*.html")
	if err != nil {
		t.Fatal(err)
	}

	body, err := r.Render("verify_email.html", VerifyEmailData{
		OTP:              "123456",
		ExpiresInMinutes: 10,
		Year:             2026,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(body, "123456") {
		t.Error("rendered body missing the OTP")
	}
	if !strings.Contains(body, "10 minutes") {
		t.Error("rendered body missing the expiry note")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && go test ./internal/pkg/mailer/...`
Expected: FAIL with `undefined: VerifyEmailData` (and the template does not exist yet).

- [ ] **Step 3: Create the template**

`internal/pkg/mailer/templates/verify_email.html` (styled like `reset_password.html`):

```html
{{ define "verify_email.html" }}

<div style="background:#f4f4f5;padding:40px 16px;font-family:Arial, sans-serif;">

  <!-- Card -->
  <div style="max-width:480px;margin:0 auto;background:#ffffff;border-radius:12px;padding:32px;">

    <!-- Header -->
    <h1 style="margin:0 0 12px 0;font-size:20px;font-weight:600;color:#111827;">
      Verify your email
    </h1>

    <p style="margin:0 0 24px 0;font-size:14px;line-height:1.6;color:#6b7280;">
      Thanks for signing up! Use the code below to verify your email address and activate your account.
    </p>

    <!-- OTP -->
    <div style="margin:24px 0;text-align:center;background:#f9fafb;border:1px solid #e5e7eb;border-radius:12px;padding:24px;">
      <p style="margin:0 0 8px 0;font-size:12px;color:#6b7280;">Your verification code</p>
      <p style="margin:0;font-size:32px;font-weight:700;letter-spacing:8px;color:#111827;">{{ .OTP }}</p>
    </div>

    <!-- Divider -->
    <div style="height:1px;background:#e5e7eb;margin:24px 0;"></div>

    <!-- Security note -->
    <p style="margin:0 0 8px 0;font-size:12px;line-height:1.6;color:#6b7280;">
      This code will expire in <strong>{{ .ExpiresInMinutes }} minutes</strong> for security reasons.
    </p>

    <p style="margin:0;font-size:12px;line-height:1.6;color:#6b7280;">
      If you did not sign up for Megome, you can safely ignore this email.
    </p>

  </div>

  <!-- Footer -->
  <p style="text-align:center;margin-top:16px;font-size:11px;color:#9ca3af;">
    © {{ .Year }} Megome. All rights reserved.
  </p>

</div>

{{ end }}
```

- [ ] **Step 4: Add the service method**

Append to `internal/pkg/mailer/service.go`:

```go
type VerifyEmailData struct {
	OTP              string
	ExpiresInMinutes int
	Year             int
}

func (s *Service) SendVerifyEmail(to string, otp string) error {
	body, err := s.renderer.Render("verify_email.html", VerifyEmailData{
		OTP:              otp,
		ExpiresInMinutes: 10,
		Year:             time.Now().Year(),
	})
	if err != nil {
		return err
	}

	return s.smtp.Send(Email{
		To:          []string{to},
		Subject:     "Verify your email",
		Body:        body,
		ContentType: "text/html",
	})
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && go test ./internal/pkg/mailer/...`
Expected: PASS (`ok megome/internal/pkg/mailer`).

- [ ] **Step 6: Commit**

```bash
cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && git add internal/pkg/mailer/templates/verify_email.html internal/pkg/mailer/service.go internal/pkg/mailer/service_test.go && git commit -m "feat(mailer): add email verification template and send method"
```

---

### Task 4: `emailverification` domain — repository + cooldown helper

**Files:**
- Create: `internal/domain/emailverification/model.go`
- Create: `internal/domain/emailverification/repository.go`
- Create: `internal/domain/emailverification/cooldown.go`
- Test: `internal/domain/emailverification/cooldown_test.go`

**Interfaces:**
- Consumes: `*sql.DB`.
- Produces: `Repository` with `NewRepository(db *sql.DB) *Repository` and methods `SaveOTP(userId int, email string, otpHash string, exp time.Time) error`, `LastOTPSentAt(userId int) (time.Time, error)`, `DeleteOTPs(userId int) error`, `VerifyOTP(email string, otpHash string) (int, error)`, `MarkVerified(userId int) error`. Package constants `OTPExpiry = 10 * time.Minute`, `ResendCooldown = 60 * time.Second`. Pure helper `RemainingCooldown(lastSentAt, now time.Time, cooldown time.Duration) int64` returning 0 when the cooldown has elapsed (or `lastSentAt` is zero) and otherwise the ceil of remaining seconds. Sentinel error `ErrInvalidOTP`. The DB-backed methods have no unit test (matches the `passwordforgot` precedent); `RemainingCooldown` is unit-tested.

- [ ] **Step 1: Write the failing test for `RemainingCooldown`**

`internal/domain/emailverification/cooldown_test.go`:

```go
package emailverification

import (
	"testing"
	"time"
)

func TestRemainingCooldown(t *testing.T) {
	cooldown := time.Minute
	now := time.Now()

	if got := RemainingCooldown(time.Time{}, now, cooldown); got != 0 {
		t.Errorf("zero lastSentAt: expected 0, got %d", got)
	}
	if got := RemainingCooldown(now.Add(-time.Minute), now, cooldown); got != 0 {
		t.Errorf("cooldown elapsed exactly: expected 0, got %d", got)
	}
	if got := RemainingCooldown(now.Add(-61*time.Second), now, cooldown); got != 0 {
		t.Errorf("cooldown exceeded: expected 0, got %d", got)
	}
	if got := RemainingCooldown(now.Add(-30*time.Second), now, cooldown); got != 30 {
		t.Errorf("30s elapsed: expected 30, got %d", got)
	}
	if got := RemainingCooldown(now.Add(-10*time.Millisecond), now, cooldown); got != 60 {
		t.Errorf("10ms elapsed: expected 60 (ceil), got %d", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && go test ./internal/domain/emailverification/...`
Expected: FAIL with `undefined: RemainingCooldown` (package does not exist).

- [ ] **Step 3: Create `model.go`**

`internal/domain/emailverification/model.go`:

```go
package emailverification

import "time"

const (
	OTPExpiry      = 10 * time.Minute
	ResendCooldown = 60 * time.Second
)

type EmailVerificationStore interface {
	SaveOTP(userId int, email string, otpHash string, exp time.Time) error
	LastOTPSentAt(userId int) (time.Time, error)
	DeleteOTPs(userId int) error
	VerifyOTP(email string, otpHash string) (int, error)
	MarkVerified(userId int) error
}
```

- [ ] **Step 4: Create `cooldown.go`**

`internal/domain/emailverification/cooldown.go`:

```go
package emailverification

import (
	"math"
	"time"
)

// RemainingCooldown returns the number of whole seconds left before the resend
// cooldown expires, rounded up, or 0 if the cooldown has already elapsed.
func RemainingCooldown(lastSentAt, now time.Time, cooldown time.Duration) int64 {
	if lastSentAt.IsZero() {
		return 0
	}
	remaining := cooldown - now.Sub(lastSentAt)
	if remaining <= 0 {
		return 0
	}
	return int64(math.Ceil(remaining.Seconds()))
}
```

- [ ] **Step 5: Create `repository.go`**

`internal/domain/emailverification/repository.go`:

```go
package emailverification

import (
	"database/sql"
	"errors"
	"time"
)

var ErrInvalidOTP = errors.New("invalid or expired verification code")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveOTP(userId int, email string, otpHash string, exp time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO email_verification_otps (userId, email, otpHash, expiresAt, createdAt)
		VALUES (?, ?, ?, ?, ?)
	`, userId, email, otpHash, exp, time.Now())
	return err
}

func (r *Repository) LastOTPSentAt(userId int) (time.Time, error) {
	var createdAt time.Time
	err := r.db.QueryRow(`
		SELECT createdAt
		FROM email_verification_otps
		WHERE userId = ?
		ORDER BY createdAt DESC
		LIMIT 1
	`, userId).Scan(&createdAt)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	return createdAt, err
}

func (r *Repository) DeleteOTPs(userId int) error {
	_, err := r.db.Exec("DELETE FROM email_verification_otps WHERE userId = ?", userId)
	return err
}

func (r *Repository) VerifyOTP(email string, otpHash string) (int, error) {
	var userId int
	var expiresAt time.Time
	var usedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT userId, expiresAt, usedAt
		FROM email_verification_otps
		WHERE email = ? AND otpHash = ?
		LIMIT 1
	`, email, otpHash).Scan(&userId, &expiresAt, &usedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrInvalidOTP
		}
		return 0, err
	}

	if usedAt.Valid {
		return 0, ErrInvalidOTP
	}

	if time.Now().After(expiresAt) {
		return 0, ErrInvalidOTP
	}

	return userId, nil
}

func (r *Repository) MarkVerified(userId int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		UPDATE users
		SET emailVerifiedAt = ?
		WHERE id = ?
	`, time.Now(), userId); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		DELETE FROM email_verification_otps
		WHERE userId = ?
	`, userId); err != nil {
		return err
	}

	return tx.Commit()
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && go test ./internal/domain/emailverification/...`
Expected: PASS (`ok megome/internal/domain/emailverification`).

- [ ] **Step 7: Commit**

```bash
cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && git add internal/domain/emailverification && git commit -m "feat(domain): add email verification repository and cooldown helper"
```

---

### Task 5: User domain — `EmailVerifiedAt`, payloads, `MarkEmailVerified`

**Files:**
- Modify: `internal/domain/user/model.go`
- Modify: `internal/domain/user/repository.go`

**Interfaces:**
- Consumes: nothing.
- Produces: `User.EmailVerifiedAt *time.Time` (json `emailVerifiedAt`); payloads `VerifyEmailPayload{Email, OTP}` and `ResendOTPPayload{Email}`; `(*user.Repository).MarkEmailVerified(id int) error` (single `UPDATE users SET emailVerifiedAt = NOW()`); `UserStore` interface gains `MarkEmailVerified`.

- [ ] **Step 1: Update the model**

In `internal/domain/user/model.go`, add the field to the `User` struct and the two payloads plus the interface method. Replace:

```go
type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByEmailOrUsername(input string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(User) (*User, error)
	GetOAuthAccount(provider string, providerUserID string) (*OAuthAccount, error)
	CreateOAuthAccount(account OAuthAccount) error
}

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
}
```

with:

```go
type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByEmailOrUsername(input string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(User) (*User, error)
	GetOAuthAccount(provider string, providerUserID string) (*OAuthAccount, error)
	CreateOAuthAccount(account OAuthAccount) error
	MarkEmailVerified(id int) error
}

type User struct {
	ID              int        `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Password        string     `json:"password"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}
```

Then add the payloads after `ForgotPassPayload`:

```go
type VerifyEmailPayload struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

type ResendOTPPayload struct {
	Email string `json:"email" validate:"required,email"`
}
```

- [ ] **Step 2: Update the repository**

In `internal/domain/user/repository.go`:

1. Add `"time"` to the imports.
2. In `GetUserByEmail`, replace the column list with one that includes `emailVerifiedAt`:

```go
	rows, err := s.db.Query("SELECT id, username, email, password, emailVerifiedAt, createdAt FROM users WHERE email = ?", email)
```

3. Update `scanRowIntoUser` to scan the new column (the `SELECT *` queries in `GetUserByEmailOrUsername` and `GetUserByID` already include the column):

```go
	err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
	)
```

4. Add `MarkEmailVerified` (place it next to `CreateOAuthAccount`):

```go
func (s *Repository) MarkEmailVerified(id int) error {
	_, err := s.db.Exec("UPDATE users SET emailVerifiedAt = ? WHERE id = ?", time.Now(), id)
	return err
}
```

- [ ] **Step 3: Verify build and tests**

Run: `cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && go build ./... && go vet ./... && go test ./...`
Expected: build/vet clean; all tests PASS.

- [ ] **Step 4: Commit**

```bash
cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && git add internal/domain/user/model.go internal/domain/user/repository.go && git commit -m "feat(domain): add emailVerifiedAt to user model and mark-verified repository method"
```

---

### Task 6: Handlers — register sends OTP, verify-email, resend-otp, login gating, Google auto-verify

**Files:**
- Modify: `internal/api/handler/user.go`
- Modify: `internal/api/router.go`

**Interfaces:**
- Consumes: `auth.GenerateOTP` (Task 2), `mailer.Service.SendVerifyEmail` (Task 3), `emailverification.Repository` (Task 4), `user.VerifyEmailPayload` / `user.ResendOTPPayload` / `User.EmailVerifiedAt` / `user.Repository.MarkEmailVerified` (Task 5).
- Produces: `NewUserHandler` accepts an additional `*emailverification.Repository`; routes `POST /auth/verify-email` and `POST /auth/resend-otp`; register responds `201 {success, message, email}` with NO tokens; login returns `403 {error, email}` for unverified; Google-created users are marked verified.

- [ ] **Step 1: Update imports and constructor**

In `internal/api/handler/user.go`:
1. Add imports `"crypto/sha256"`, `"encoding/hex"`, and `"megome/internal/domain/emailverification"`.
2. Add the field to the struct and the parameter to the constructor. Replace:

```go
type UserHandler struct {
	userStore      *user.Repository
	profileStore   *profile.Repository
	refreshStore   *refreshtoken.Repository
	emailService   *mailer.Service
	passwordForgot *passwordforgot.Repository
}

func NewUserHandler(userStore *user.Repository, profileStore *profile.Repository, refreshStore *refreshtoken.Repository, emailService *mailer.Service, passwordForgot *passwordforgot.Repository) *UserHandler {
	return &UserHandler{userStore: userStore, profileStore: profileStore, refreshStore: refreshStore, emailService: emailService, passwordForgot: passwordForgot}
}
```

with:

```go
type UserHandler struct {
	userStore         *user.Repository
	profileStore      *profile.Repository
	refreshStore      *refreshtoken.Repository
	emailService      *mailer.Service
	passwordForgot    *passwordforgot.Repository
	emailVerification *emailverification.Repository
}

func NewUserHandler(userStore *user.Repository, profileStore *profile.Repository, refreshStore *refreshtoken.Repository, emailService *mailer.Service, passwordForgot *passwordforgot.Repository, emailVerification *emailverification.Repository) *UserHandler {
	return &UserHandler{userStore: userStore, profileStore: profileStore, refreshStore: refreshStore, emailService: emailService, passwordForgot: passwordForgot, emailVerification: emailVerification}
}
```

- [ ] **Step 2: Register the new routes**

In `RegisterRoutes`, after the `change-forgot-pass` line add:

```go
	router.HandleFunc("/auth/verify-email", h.handleVerifyEmail).Methods("POST")
	router.HandleFunc("/auth/resend-otp", h.handleResendOTP).Methods("POST")
```

- [ ] **Step 3: Modify `handleRegister`**

Replace the token-issuing block at the end of `handleRegister` (from `at, rt, err := h.getTokens(u.ID)` through the final `httputil.WriteJSON(...)`):

```go
	otp, err := auth.GenerateOTP()
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	hash := sha256.Sum256([]byte(otp))
	otpHash := hex.EncodeToString(hash[:])

	err = h.emailVerification.SaveOTP(u.ID, u.Email, otpHash, time.Now().Add(emailverification.OTPExpiry))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.emailService.SendVerifyEmail(u.Email, otp)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "verification code sent to your email",
		"email":   u.Email,
	})
```

- [ ] **Step 4: Add `handleVerifyEmail`**

Add this method to `user.go` (after `handleRegister`):

```go
func (h *UserHandler) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var payload user.VerifyEmailPayload

	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	hash := sha256.Sum256([]byte(payload.OTP))
	otpHash := hex.EncodeToString(hash[:])

	userId, err := h.emailVerification.VerifyOTP(payload.Email, otpHash)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.emailVerification.MarkVerified(userId); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	at, rt, err := h.getTokens(userId)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := refreshtoken.AuthResponse{
		Success:      true,
		Message:      "Email verified successfully",
		AccessToken:  at,
		RefreshToken: rt,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 5: Add `handleResendOTP`**

Add this method to `user.go` (after `handleVerifyEmail`):

```go
func (h *UserHandler) handleResendOTP(w http.ResponseWriter, r *http.Request) {
	var payload user.ResendOTPPayload

	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	u, err := h.userStore.GetUserByEmail(payload.Email)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}

	if u.EmailVerifiedAt != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("email already verified"))
		return
	}

	lastSent, err := h.emailVerification.LastOTPSentAt(u.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if remaining := emailverification.RemainingCooldown(lastSent, time.Now(), emailverification.ResendCooldown); remaining > 0 {
		httputil.WriteJSON(w, http.StatusConflict, map[string]any{
			"error":             "please wait before requesting another code",
			"retryAfterSeconds": remaining,
		})
		return
	}

	otp, err := auth.GenerateOTP()
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	hash := sha256.Sum256([]byte(otp))
	otpHash := hex.EncodeToString(hash[:])

	if err := h.emailVerification.DeleteOTPs(u.ID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.emailVerification.SaveOTP(u.ID, u.Email, otpHash, time.Now().Add(emailverification.OTPExpiry)); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.emailService.SendVerifyEmail(u.Email, otp); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "verification code sent to your email",
		"email":   u.Email,
	})
}
```

- [ ] **Step 6: Gate `handleLogin` on verification**

In `handleLogin`, immediately after the password comparison block (the `if !auth.ComparePasswords(...)` block), add:

```go
	if u.EmailVerifiedAt == nil {
		httputil.WriteJSON(w, http.StatusForbidden, map[string]any{
			"error": "email not verified",
			"email": u.Email,
		})
		return
	}
```

- [ ] **Step 7: Auto-verify Google-created users**

In `handleGoogleCallback`, inside the `else` branch where a new user is created via `h.userStore.CreateUser(...)`, add immediately after that `CreateUser` call and its error check:

```go
		err = h.userStore.MarkEmailVerified(u.ID)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err)
			return
		}
```

- [ ] **Step 8: Wire the repository in the router**

In `internal/api/router.go`:
1. Add import `"megome/internal/domain/emailverification"`.
2. After `passwordForgotRepo := passwordforgot.NewRepository(s.db)` add:

```go
	emailVerificationRepo := emailverification.NewRepository(s.db)
```

3. Replace the `NewUserHandler` call:

```go
	handler.NewUserHandler(userRepo, profileRepo, refreshRepo, emailService, passwordForgotRepo, emailVerificationRepo).RegisterRoutes(internal)
```

- [ ] **Step 9: Verify build and tests**

Run: `cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && go build ./... && go vet ./... && go test ./...`
Expected: build/vet clean; all tests PASS.

- [ ] **Step 10: Commit**

```bash
cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && git add internal/api/handler/user.go internal/api/router.go && git commit -m "feat(api): add OTP email verification endpoints and gate login on verification"
```

---

### Task 7: Frontend proxy routes — register, login, verify-email, resend-otp

**Files:**
- Modify: `app/api/auth/register/route.ts`
- Modify: `app/api/auth/login/route.ts`
- Create: `app/api/auth/verify-email/route.ts`
- Create: `app/api/auth/resend-otp/route.ts`

**Interfaces:**
- Consumes: backend endpoints `/api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/auth/verify-email`, `/api/v1/auth/resend-otp`; `cookies` from `next/headers`.
- Produces: `POST /api/auth/register` returns `{success, message, email}` (no cookies); `POST /api/auth/login` passes through `email` on error; `POST /api/auth/verify-email` sets `access_token`/`refresh_token` httpOnly cookies and returns `{success, user}`; `POST /api/auth/resend-otp` passes through the backend response.
- All work in `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification`.

- [ ] **Step 1: Rewrite `register/route.ts` (stop setting cookies, pass `email`)**

Replace the entire file:

```ts
import { NextResponse } from "next/server";

const BACKEND_URL = process.env.NEXT_PUBLIC_API_URL!;

export async function POST(req: Request) {
  try {
    const body = await req.json();

    const response = await fetch(
      `${BACKEND_URL}/api/v1/auth/register`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      }
    );

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(
        {
          success: false,
          message: data.error || "Register failed",
        },
        {
          status: response.status,
        }
      );
    }

    return NextResponse.json({
      success: true,
      message: data.message || "Verification code sent to your email",
      email: data.email,
    });
  } catch (_) {
    return NextResponse.json(
      {
        success: false,
        message: "Internal server error",
      },
      {
        status: 500,
      }
    );
  }
}
```

- [ ] **Step 2: Add `email` passthrough to `login/route.ts`**

Replace the error block in `login/route.ts`:

```ts
    if (!response.ok) {
      return NextResponse.json(
        {
          success: false,
          message: data.error || "Login failed",
          ...(data.email ? { email: data.email } : {}),
        },
        {
          status: response.status,
        }
      );
    }
```

- [ ] **Step 3: Create `verify-email/route.ts`**

```ts
import { NextResponse } from "next/server";
import { cookies } from "next/headers";

const BACKEND_URL = process.env.NEXT_PUBLIC_API_URL!;

export async function POST(req: Request) {
  try {
    const body = await req.json();

    const response = await fetch(
      `${BACKEND_URL}/api/v1/auth/verify-email`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      }
    );

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(
        {
          success: false,
          message: data.error || "Verification failed",
        },
        {
          status: response.status,
        }
      );
    }

    const cookieStore = await cookies();

    cookieStore.set("access_token", data.accessToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
    });

    cookieStore.set("refresh_token", data.refreshToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
    });

    return NextResponse.json({
      success: true,
      user: data.user,
    });
  } catch (_) {
    return NextResponse.json(
      {
        success: false,
        message: "Internal server error",
      },
      {
        status: 500,
      }
    );
  }
}
```

- [ ] **Step 4: Create `resend-otp/route.ts`**

```ts
import { NextResponse } from "next/server";

const BACKEND_URL = process.env.NEXT_PUBLIC_API_URL!;

export async function POST(req: Request) {
  try {
    const body = await req.json();

    return await fetch(`${BACKEND_URL}/api/v1/auth/resend-otp`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });
  } catch (_) {
    return NextResponse.json(
      { message: "Internal server error" },
      { status: 500 }
    );
  }
}
```

- [ ] **Step 5: Verify lint**

Run: `cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && npm run lint`
Expected: lint clean.

- [ ] **Step 6: Commit**

```bash
cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && git add app/api/auth/register/route.ts app/api/auth/login/route.ts app/api/auth/verify-email/route.ts app/api/auth/resend-otp/route.ts && git commit -m "feat(auth): add verify-email and resend-otp proxy routes; register no longer sets cookies"
```

---

### Task 8: Frontend client functions + schema

**Files:**
- Modify: `lib/api/client/auth.ts`
- Modify: `features/auth/schema.ts`

**Interfaces:**
- Consumes: `handleResponse` from `@/utils/api/handleResponse`.
- Produces: `verifyEmailClient(email: string, otp: string)` and `resendOtpClient(email: string)`; `verifyEmailSchema` (zod: `email` valid email, `otp` exactly 6 digits).
- All work in `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification`.

- [ ] **Step 1: Add the client functions**

Append to `lib/api/client/auth.ts` (after `registerClient`):

```ts
export const verifyEmailClient = async (email: string, otp: string) => {
  const res = await fetch(
    "/api/auth/verify-email",
    {
      method: "POST",
      body: JSON.stringify({ email, otp }),
    },
  )
  return handleResponse<Response>(res)
}

export const resendOtpClient = async (email: string) => {
  const res = await fetch(
    "/api/auth/resend-otp",
    {
      method: "POST",
      body: JSON.stringify({ email }),
    },
  )
  return handleResponse<Response>(res)
}
```

- [ ] **Step 2: Add the schema**

Append to `features/auth/schema.ts`:

```ts
export const verifyEmailSchema = z.object({
  email: z.string().email("invalid email address"),
  otp: z.string().regex(/^\d{6}$/, "enter the 6-digit code"),
});
```

- [ ] **Step 3: Verify lint**

Run: `cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && npm run lint`
Expected: lint clean.

- [ ] **Step 4: Commit**

```bash
cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && git add lib/api/client/auth.ts features/auth/schema.ts && git commit -m "feat(auth): add verify-email and resend-otp clients and schema"
```

---

### Task 9: AuthForm — register redirects to verify, login routes unverified users to verify

**Files:**
- Modify: `features/auth/components/AuthForm.tsx`

**Interfaces:**
- Consumes: `registerClient`, `loginClient` (Task 8 unchanged), `getInitClient`, `useRouter`, `useToast` (all existing).
- Produces: `handleRegister` shows a success toast and pushes `/auth/verify-email?email=<email>`; `handleLogin` catches a `403` response carrying `data.email` and pushes `/auth/verify-email?email=<email>`.

- [ ] **Step 1: Update `handleLogin`**

Replace `handleLogin` in `features/auth/components/AuthForm.tsx`:

```tsx
  const handleLogin = async () => {
    try {
      const res = await loginClient(emailOrUsername, password);

      const initData = await getInitClient();

      if (!initData.profile) {
        router.push("/profile-setup");
        return;
      }

      router.push("/dashboard");
    } catch (err: any) {
      if (err?.data?.email) {
        router.push(
          `/auth/verify-email?email=${encodeURIComponent(err.data.email)}`
        );
        return;
      }
      throw err;
    }
  };
```

- [ ] **Step 2: Update `handleRegister`**

Replace `handleRegister`:

```tsx
  const handleRegister = async () => {
    const res = await registerClient(username, email, password);

    showToast(res.message, "success");

    router.push(`/auth/verify-email?email=${encodeURIComponent(email)}`);
  };
```

- [ ] **Step 3: Verify lint and build**

Run: `cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && npm run lint && npm run build`
Expected: lint clean, build succeeds.

- [ ] **Step 4: Commit**

```bash
cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && git add features/auth/components/AuthForm.tsx && git commit -m "feat(auth): route sign-up and unverified sign-in to email verification"
```

---

### Task 10: Verify-email page

**Files:**
- Create: `app/(auth)/auth/verify-email/page.tsx`

**Interfaces:**
- Consumes: `verifyEmailClient`, `resendOtpClient` (Task 8), `getInitClient`, `withRequest` from `@/utils/api/withRequest`, `useToast` from `@/components/ui/toast/useToast`, `Card` from `@/components/ui/Card`.
- Produces: page at `/auth/verify-email` with email (prefilled from `?email=`, editable), 6-digit OTP input, verify button, resend button with a 60s countdown, and "Back to Sign In" link. On success it redirects to `/profile-setup` (no profile) or `/dashboard`.
- All work in `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification`.

- [ ] **Step 1: Create the page**

`app/(auth)/auth/verify-email/page.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { ArrowRightIcon } from "@heroicons/react/16/solid";
import { Card } from "@/components/ui/Card";
import { useToast } from "@/components/ui/toast/useToast";
import { resendOtpClient, verifyEmailClient } from "@/lib/api/client/auth";
import { getInitClient } from "@/lib/api/client/init";
import { withRequest } from "@/utils/api/withRequest";

const RESEND_COOLDOWN = 60;

export default function VerifyEmailPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { showToast } = useToast();

  const [email, setEmail] = useState(searchParams?.get("email") || "");
  const [otp, setOtp] = useState("");
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);
  const [cooldown, setCooldown] = useState(0);

  useEffect(() => {
    if (cooldown <= 0) return;

    const timer = setTimeout(() => setCooldown((c) => c - 1), 1000);
    return () => clearTimeout(timer);
  }, [cooldown]);

  const redirectAfterLogin = async () => {
    const initData = await getInitClient();

    if (!initData.profile) {
      router.push("/profile-setup");
      return;
    }

    router.push("/dashboard");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setLoading(true);

      const res = await withRequest(
        () => verifyEmailClient(email, otp),
        showToast
      );

      if (res && (res as any).success) {
        showToast("Email verified successfully", "success");
        await redirectAfterLogin();
      }
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    try {
      setResending(true);

      await withRequest(() => resendOtpClient(email), showToast);

      setCooldown(RESEND_COOLDOWN);
    } finally {
      setResending(false);
    }
  };

  return (
    <main className="min-h-screen bg-base-200 flex items-center justify-center px-4 py-10">
      <div className="w-full max-w-md">
        <Card className="p-6 sm:p-8 shadow-lg">

          {/* Header */}
          <div className="text-center space-y-2 mb-6">
            <h2 className="text-2xl sm:text-3xl font-bold text-primary">
              Verify Your Email
            </h2>

            <p className="text-sm sm:text-base text-base-content/70 leading-relaxed">
              We sent a 6-digit code to your email. Enter it below to activate
              your account.
            </p>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-5">
            <fieldset className="fieldset">
              <legend className="fieldset-legend">Email</legend>

              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="input w-full"
                required
              />
            </fieldset>

            <fieldset className="fieldset">
              <legend className="fieldset-legend">Verification code</legend>

              <input
                type="text"
                inputMode="numeric"
                maxLength={6}
                value={otp}
                onChange={(e) => setOtp(e.target.value.replace(/\D/g, ""))}
                className="input w-full text-center tracking-[8px] font-bold"
                placeholder="000000"
                required
              />
            </fieldset>

            {/* Actions */}
            <div className="flex flex-col-reverse sm:flex-row sm:items-center sm:justify-between gap-3 pt-2">

              <Link
                href="/auth"
                className="text-sm text-accent hover:opacity-80 transition"
              >
                Back to Sign In
              </Link>

              <button
                type="submit"
                disabled={loading}
                className="
                  flex items-center justify-center gap-2
                  bg-primary text-primary-content
                  px-4 py-2 rounded-md font-bold
                  w-full sm:w-auto
                  disabled:opacity-60
                "
              >
                {loading ? "Verifying..." : "Verify Email"}
                <ArrowRightIcon className="h-5 w-5" />
              </button>
            </div>
          </form>

          {/* Resend */}
          <div className="mt-6 pt-4 border-t border-base-300 text-center">
            <p className="text-sm text-base-content/70 mb-2">
              Did not receive the code?
            </p>

            <button
              type="button"
              onClick={handleResend}
              disabled={resending || cooldown > 0 || !email}
              className="text-sm text-accent hover:opacity-80 transition disabled:opacity-60"
            >
              {cooldown > 0
                ? `Resend code in ${cooldown}s`
                : resending
                ? "Sending..."
                : "Resend code"}
            </button>
          </div>

        </Card>
      </div>
    </main>
  );
}
```

- [ ] **Step 2: Verify lint and build**

Run: `cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && npm run lint && npm run build`
Expected: lint clean, build succeeds.

- [ ] **Step 3: Commit**

```bash
cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && git add "app/(auth)/auth/verify-email/page.tsx" && git commit -m "feat(auth): add email verification page with resend cooldown"
```

---

### Task 11: Fix all pre-existing frontend lint errors

**Files:**
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/app/(auth)/auth/reset-password/page.tsx`
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/components/ui/modal/Modal.tsx`
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/components/ui/rich-editor/RichEditor.tsx`
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/components/ui/Sidebar.tsx`
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/components/ui/ThemeToggle.tsx`
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/features/ai/components/AiAssistModal.tsx`
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/features/ai/components/AiStatusBanner.tsx`
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/features/auth/components/AuthForm.tsx`
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/features/profile/components/TopProfile.tsx`
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/features/project/components/ProjectWizard.tsx`
- `/home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification/utils/api/withRequest.ts`

**Context:**
- Baseline `npm run lint` reports 15 errors + 70 warnings. Fix the 15 ERRORS so `npm run lint` exits 0. Warnings may remain.
- Error inventory (from `npx eslint .` at commit `35a5a1e`; regenerated via path-aware scan):
  - `app/(auth)/auth/reset-password/page.tsx`: 1× `no-explicit-any` (56).
  - `components/ui/modal/Modal.tsx`: 1× react-hooks set-state-in-effect (37).
  - `components/ui/rich-editor/RichEditor.tsx`: 1× react-hooks refs-during-render (40).
  - `components/ui/Sidebar.tsx`: 1× `no-explicit-any` (22).
  - `components/ui/ThemeToggle.tsx`: 1× react-hooks set-state-in-effect (14).
  - `features/ai/components/AiAssistModal.tsx`: 1× react-hooks set-state-in-effect (29).
  - `features/ai/components/AiStatusBanner.tsx`: 1× react-hooks set-state-in-effect (22).
  - `features/auth/components/AuthForm.tsx`: 1× `no-explicit-any` (96, in `handleAction`'s catch — note Task 9 may already have touched this file; fix whatever `any` remains).
  - `features/profile/components/TopProfile.tsx`: 2× `react/no-unescaped-entities` (150) — escape `"` as `&quot;` in JSX text.
  - `features/project/components/ProjectWizard.tsx`: 2× `no-explicit-any` (159, 195).
  - `utils/api/withRequest.ts`: 3× `no-explicit-any` (7, 8, 11).
- Fix via React Compiler-safe patterns (wrap `setState` in event handlers; gate with layout effect or derive state; hoist ref reads into effects/events; replace `any` with specific types; escape `"` as `&quot;` in JSX text).
- Keep runtime behavior identical. No `eslint-disable` comments.

- [ ] **Step 1: Fix the errors**

Fix each `no-explicit-any` with a proper type (`unknown`, `Record<string, unknown>`, or a specific interface). For react-hooks errors, restructure so state updates happen in event handlers (or via an effect reading a changed dependency) rather than synchronously during render; read refs only inside effects/handlers. Re-escape the quotes on TopProfile.tsx:150.

- [ ] **Step 2: Verify lint passes**

Run: `cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && npm run lint`
Expected: exit code 0 (no errors; warnings tolerated).

- [ ] **Step 3: Verify build still passes**

Run: `cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && npm run build`
Expected: build succeeds.

- [ ] **Step 4: Commit**

`cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && git add "app/(auth)/auth/reset-password/page.tsx" components/ui/modal/Modal.tsx components/ui/rich-editor/RichEditor.tsx components/ui/Sidebar.tsx components/ui/ThemeToggle.tsx features/ai/components/AiAssistModal.tsx features/ai/components/AiStatusBanner.tsx features/auth/components/AuthForm.tsx features/profile/components/TopProfile.tsx features/project/components/ProjectWizard.tsx utils/api/withRequest.ts && git commit -m "fix(lint): resolve pre-existing eslint errors across UI and ai components"`

### Task 12: End-to-end verification

**Files:**
- None (verification only).

- [ ] **Step 1: Run all backend checks**

Run: `cd /home/alexanderudag/dev/megome/megome/.worktrees/otp-email-verification && go build ./... && go vet ./... && go test ./...`
Expected: build/vet clean, all tests PASS (including `internal/pkg/auth`, `internal/pkg/mailer`, `internal/domain/emailverification`).

- [ ] **Step 2: Run all frontend checks**

Run: `cd /home/alexanderudag/dev/megome/megome-front/.worktrees/otp-email-verification && npm run lint && npm run build`
Expected: lint clean, build succeeds.

- [ ] **Step 3: Manual QA checklist**

With both servers running and SMTP configured in the backend `.env`:

1. Register a new account → receive an email with a 6-digit code; you are NOT logged in.
2. Visit the verify page with the wrong code → clear error, still unverified.
3. Enter the correct code → toast "Email verified successfully", auto-login, redirected to `/profile-setup` or `/dashboard`.
4. Log out, then attempt to log in with the now-verified account → succeeds.
5. Register a second account, close the verify page, try to log in → `403` and the frontend redirects to the verify page.
6. On the verify page, click resend twice within 60s → second attempt shows "please wait before requesting another code".
7. Confirm Google OAuth sign-in still works end to end (email considered verified).
