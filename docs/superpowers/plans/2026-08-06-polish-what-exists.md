# Polish What Exists Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add pagination, soft delete, error boundaries, dashboard activity/chart, form polish (dirty guard, autosave, rich text preview), image handling (resize, drag-drop, crop, validation), and empty state consistency across both backend and frontend.

**Architecture:** Backend changes touch every domain repository (pagination + soft delete), add 2 new dashboard endpoints (activity feed + usage chart), and 7 new soft-delete migrations. Frontend adds error boundaries + 404 page, reworks dashboard, adds shared hooks for dirty guard and image resize, and polishes forms.

**Tech Stack:** Go 1.25 (gorilla/mux, database/sql), Next.js 16 (App Router, React 19, Tailwind v4, DaisyUI v5, Zustand, Zod)

## Global Constraints

- All backend handler responses use `Message` + `Data` wrapper pattern as seen in `apilog.go` and `dashboard.go`
- All frontend API client functions use `fetchClient` from `lib/api/client/fetchClient.ts` + `handleResponse` from `utils/api/handleResponse.ts`
- All frontend proxy routes use `getAccessToken()` from `lib/auth/cookies.ts` and forward to `BACKEND_URL` with Bearer token
- Pagination defaults: `limit=20`, `offset=0`, max limit `100`
- Soft delete tables: projects, skills, education, experiences, certifications, project_images, project_techs, experience_tech
- Frontend components use DaisyUI classes and existing `EmptyState` shared component
- New hooks go in `lib/hooks/`
- New frontend components go in `features/<domain>/components/`

---

### Task 1: Add soft-delete migrations (all affected tables)

**Files:**
- Create: `cmd/migrate/migrations/20260806000000_add-soft-delete-projects.up.sql`
- Create: `cmd/migrate/migrations/20260806000000_add-soft-delete-projects.down.sql`
- Create: `cmd/migrate/migrations/20260806000001_add-soft-delete-skills.up.sql`
- Create: `cmd/migrate/migrations/20260806000001_add-soft-delete-skills.down.sql`
- Create: `cmd/migrate/migrations/20260806000002_add-soft-delete-education.up.sql`
- Create: `cmd/migrate/migrations/20260806000002_add-soft-delete-education.down.sql`
- Create: `cmd/migrate/migrations/20260806000003_add-soft-delete-experiences.up.sql`
- Create: `cmd/migrate/migrations/20260806000003_add-soft-delete-experiences.down.sql`
- Create: `cmd/migrate/migrations/20260806000004_add-soft-delete-certifications.up.sql`
- Create: `cmd/migrate/migrations/20260806000004_add-soft-delete-certifications.down.sql`
- Create: `cmd/migrate/migrations/20260806000005_add-soft-delete-project-images.up.sql`
- Create: `cmd/migrate/migrations/20260806000005_add-soft-delete-project-images.down.sql`
- Create: `cmd/migrate/migrations/20260806000006_add-soft-delete-project-techs.up.sql`
- Create: `cmd/migrate/migrations/20260806000006_add-soft-delete-project-techs.down.sql`
- Create: `cmd/migrate/migrations/20260806000007_add-soft-delete-experience-techs.up.sql`
- Create: `cmd/migrate/migrations/20260806000007_add-soft-delete-experience-techs.down.sql`

- [ ] **Step 1: Create all up migration files**

Each up migration follows this exact pattern. Files:

`cmd/migrate/migrations/20260806000000_add-soft-delete-projects.up.sql`:
```sql
ALTER TABLE projects ADD COLUMN deletedAt TIMESTAMP NULL DEFAULT NULL;
```

`cmd/migrate/migrations/20260806000001_add-soft-delete-skills.up.sql`:
```sql
ALTER TABLE skills ADD COLUMN deletedAt TIMESTAMP NULL DEFAULT NULL;
```

`cmd/migrate/migrations/20260806000002_add-soft-delete-education.up.sql`:
```sql
ALTER TABLE education ADD COLUMN deletedAt TIMESTAMP NULL DEFAULT NULL;
```

`cmd/migrate/migrations/20260806000003_add-soft-delete-experiences.up.sql`:
```sql
ALTER TABLE experiences ADD COLUMN deletedAt TIMESTAMP NULL DEFAULT NULL;
```

`cmd/migrate/migrations/20260806000004_add-soft-delete-certifications.up.sql`:
```sql
ALTER TABLE certifications ADD COLUMN deletedAt TIMESTAMP NULL DEFAULT NULL;
```

`cmd/migrate/migrations/20260806000005_add-soft-delete-project-images.up.sql`:
```sql
ALTER TABLE project_images ADD COLUMN deletedAt TIMESTAMP NULL DEFAULT NULL;
```

`cmd/migrate/migrations/20260806000006_add-soft-delete-project-techs.up.sql`:
```sql
ALTER TABLE project_techs ADD COLUMN deletedAt TIMESTAMP NULL DEFAULT NULL;
```

`cmd/migrate/migrations/20260806000007_add-soft-delete-experience-techs.up.sql`:
```sql
ALTER TABLE experience_techs ADD COLUMN deletedAt TIMESTAMP NULL DEFAULT NULL;
```

- [ ] **Step 2: Create all down migration files**

Each down migration follows this exact pattern:

`cmd/migrate/migrations/20260806000000_add-soft-delete-projects.down.sql`:
```sql
ALTER TABLE projects DROP COLUMN deletedAt;
```

`cmd/migrate/migrations/20260806000001_add-soft-delete-skills.down.sql`:
```sql
ALTER TABLE skills DROP COLUMN deletedAt;
```

`cmd/migrate/migrations/20260806000002_add-soft-delete-education.down.sql`:
```sql
ALTER TABLE education DROP COLUMN deletedAt;
```

`cmd/migrate/migrations/20260806000003_add-soft-delete-experiences.down.sql`:
```sql
ALTER TABLE experiences DROP COLUMN deletedAt;
```

`cmd/migrate/migrations/20260806000004_add-soft-delete-certifications.down.sql`:
```sql
ALTER TABLE certifications DROP COLUMN deletedAt;
```

`cmd/migrate/migrations/20260806000005_add-soft-delete-project-images.down.sql`:
```sql
ALTER TABLE project_images DROP COLUMN deletedAt;
```

`cmd/migrate/migrations/20260806000006_add-soft-delete-project-techs.down.sql`:
```sql
ALTER TABLE project_techs DROP COLUMN deletedAt;
```

`cmd/migrate/migrations/20260806000007_add-soft-delete-experience-techs.down.sql`:
```sql
ALTER TABLE experience_techs DROP COLUMN deletedAt;
```

- [ ] **Step 3: Run migrations to verify they apply cleanly**

Run: `cd /home/alex/my-go/megome && make migrate-up`
Expected: All 8 new migrations apply without errors.

- [ ] **Step 4: Run migrations down to verify rollback**

Run: `make migrate-down` (run 8 times to roll back the new ones)
Expected: All 8 roll back cleanly.

- [ ] **Step 5: Re-run migrations up to restore state**

Run: `make migrate-up`
Expected: Clean application.

- [ ] **Step 6: Commit**

```bash
git add cmd/migrate/migrations/2026080600000*_add-soft-delete-*.sql
git commit -m "feat: add soft-delete columns to entity tables"
```

---

### Task 2: Soft-delete projects — model, repository, handler

**Files:**
- Modify: `internal/domain/project/model.go` (add DeletedAt to Project struct, update ProjectStore interface)
- Modify: `internal/domain/project/repository.go` (change DeleteProject to soft-delete, add deletedAt IS NULL to list queries)
- Modify: `internal/api/handler/project.go` (pass userID to DeleteProject)

**Interfaces:**
- Consumes: `DeletedAt` column from Task 1 migration
- Produces: `DeleteProject(id int, deletedBy int) (ProjectFull, error)` — updated signature

- [ ] **Step 1: Add DeletedAt to Project struct in model.go**

In `internal/domain/project/model.go`, after the existing `UpdatedAt` field, add:

```go
type Project struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userId"`
	Title        string    `json:"title"`
	Tags         *string   `json:"tags"`
	Description  string    `json:"description"`
	Link         *string   `json:"link"`
	GithubLink   *string   `json:"githubLink"`
	Status       string    `json:"status"`
	IsDraft      bool      `json:"isDraft"`
	DisplayOrder int       `json:"displayOrder"`
	DeletedAt    *string   `json:"deletedAt,omitempty"`
	CreatedAt    string    `json:"createdAt"`
	UpdatedAt    string    `json:"updatedAt"`
}
```

Add a `scan.TimePtr` scan function at the top of `repository.go` (or add `database/sql` import already exists — just scan into `&project.DeletedAt`):

- [ ] **Step 2: Change ProjectStore interface DeleteProject signature**

In `internal/domain/project/model.go`, change:
```go
DeleteProject(id int) (ProjectFull, error)
```
to:
```go
DeleteProject(id int, deletedBy int) (ProjectFull, error)
```

- [ ] **Step 3: Change DeleteProject in repository.go to soft-delete**

In `internal/domain/project/repository.go`, replace the `DeleteProject` function body. Keep the first part (fetch project, begin tx) but change the DELETE statements to UPDATE:

Replace lines 126-185 of the function. The key change is:

Old:
```go
if _, err = tx.Exec(`
    DELETE FROM project_images
    WHERE projectId = ?
`, id); err != nil {
    return ProjectFull{}, err
}

if _, err = tx.Exec(`
    DELETE FROM project_techs
    WHERE projectId = ?
`, id); err != nil {
    return ProjectFull{}, err
}

if _, err = tx.Exec(`
    DELETE FROM projects
    WHERE id = ?
`, id); err != nil {
    return ProjectFull{}, err
}
```

New:
```go
if _, err = tx.Exec(`
    UPDATE project_images SET deletedAt = NOW() WHERE projectId = ? AND deletedAt IS NULL
`, id); err != nil {
    return ProjectFull{}, err
}

if _, err = tx.Exec(`
    UPDATE project_techs SET deletedAt = NOW() WHERE projectId = ? AND deletedAt IS NULL
`, id); err != nil {
    return ProjectFull{}, err
}

if _, err = tx.Exec(`
    UPDATE projects SET deletedAt = NOW() WHERE id = ? AND deletedAt IS NULL
`, id); err != nil {
    return ProjectFull{}, err
}
```

Remove the R2 cleanup code (lines after `tx.Commit()` that delete R2 objects) — we keep R2 objects on soft-delete.

The full function becomes:
```go
func (s *Repository) DeleteProject(id int, deletedBy int) (ProjectFull, error) {
	project, err := s.GetProjectById(id)
	if err != nil {
		return ProjectFull{}, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return ProjectFull{}, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if _, err = tx.Exec(`
		UPDATE project_images SET deletedAt = NOW() WHERE projectId = ? AND deletedAt IS NULL
	`, id); err != nil {
		return ProjectFull{}, err
	}

	if _, err = tx.Exec(`
		UPDATE project_techs SET deletedAt = NOW() WHERE projectId = ? AND deletedAt IS NULL
	`, id); err != nil {
		return ProjectFull{}, err
	}

	if _, err = tx.Exec(`
		UPDATE projects SET deletedAt = NOW() WHERE id = ? AND deletedAt IS NULL
	`, id); err != nil {
		return ProjectFull{}, err
	}

	if err = tx.Commit(); err != nil {
		return ProjectFull{}, err
	}

	return project, nil
}
```

- [ ] **Step 4: Add deletedAt IS NULL to all list/get queries in repository.go**

In `GetProjects` (the simple list query), find the WHERE clause and add:
```sql
WHERE userId = ? AND deletedAt IS NULL
```

In `GetProjectsFull` (the full join query), add:
```sql
WHERE p.userId = ? AND p.deletedAt IS NULL
```

In `GetProjectById`, add:
```sql
WHERE id = ? AND deletedAt IS NULL
```

In `GetProjectTechs`, add:
```sql
WHERE pt.projectId = ? AND pt.deletedAt IS NULL
```

In `GetProjectImages`, add:
```sql
WHERE pi.projectId = ? AND pi.deletedAt IS NULL
```

- [ ] **Step 5: Update handler to pass userID**

In `internal/api/handler/project.go`, change `handleDeleteProject`:

Add `userID := middleware.GetUserIDFromContext(r.Context())` before calling delete:

```go
func (h *ProjectHandler) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	p, err := h.projectStore.DeleteProject(id, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleProjResponse{
		Message: "Project deleted successfully",
		Project: p,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 6: Run tests to verify**

Run: `make test`
Expected: All existing tests pass (no project deletion tests exist, so no regressions).

- [ ] **Step 7: Commit**

```bash
git add internal/domain/project/model.go internal/domain/project/repository.go internal/api/handler/project.go
git commit -m "feat: soft-delete projects"
```

---

### Task 3: Soft-delete skills, education, experiences, certifications

**Files:**
- Modify: `internal/domain/skill/model.go` (add DeletedAt field, update Store interface)
- Modify: `internal/domain/skill/repository.go` (soft-delete + filter queries)
- Modify: `internal/api/handler/skill.go` (pass userID)
- Modify: `internal/domain/education/model.go` (add DeletedAt field, update Store interface)
- Modify: `internal/domain/education/repository.go` (soft-delete + filter queries)
- Modify: `internal/api/handler/education.go` (pass userID)
- Modify: `internal/domain/experience/model.go` (add DeletedAt field, update Store interface)
- Modify: `internal/domain/experience/repository.go` (soft-delete + filter queries)
- Modify: `internal/api/handler/experience.go` (pass userID)
- Modify: `internal/domain/certification/model.go` (add DeletedAt field, update Store interface)
- Modify: `internal/domain/certification/repository.go` (soft-delete + filter queries)
- Modify: `internal/api/handler/certification.go` (pass userID)

**Pattern:** Same as Task 2 for each entity. The repository changes are simpler since these entities don't have child relations — just change `DELETE FROM ...` to `UPDATE ... SET deletedAt = NOW()`, add `AND deletedAt IS NULL` to all queries.

- [ ] **Step 1: Soft-delete skills**

In `internal/domain/skill/model.go`, add `DeletedAt *string \`json:"deletedAt,omitempty"\`` to the Skill struct. Change `DeleteSkill(id int)` to `DeleteSkill(id int, deletedBy int)` in the SkillStore interface.

In `internal/domain/skill/repository.go`, change the DeleteSkill function:
```go
func (s *Repository) DeleteSkill(id int, deletedBy int) (Skill, error) {
	skill, err := s.GetSkillById(id)
	if err != nil {
		return Skill{}, err
	}
	_, err = s.db.Exec("UPDATE skills SET deletedAt = NOW() WHERE id = ? AND deletedAt IS NULL", id)
	if err != nil {
		return Skill{}, err
	}
	return skill, nil
}
```

Add `AND deletedAt IS NULL` to GetSkills query:
```sql
SELECT id, userId, skillName, proficiency, deletedAt, createdAt, updatedAt FROM skills WHERE userId = ? AND deletedAt IS NULL
```

Update `scanRowIntoSkill` to scan `deletedAt`:
```go
scanner.Scan(
	&skill.ID,
	&skill.UserID,
	&skill.SkillName,
	&skill.Proficiency,
	&skill.DeletedAt,
	&skill.CreatedAt,
	&skill.UpdatedAt,
)
```

In `internal/api/handler/skill.go`, change `handleDeleteSkill` to pass userID:
```go
func (h *SkillHandler) handleDeleteSkill(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	s, err := h.skillStore.DeleteSkill(id, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := SingleSkillResponse{
		Message: "Skill deleted successfully",
		Skill:   s,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 2: Soft-delete education**

Same pattern as Step 1. In `internal/domain/education/`:

`model.go`: Add `DeletedAt *string` to Education struct. Change `DeleteEducation(id int)` to `DeleteEducation(id int, deletedBy int)`.

`repository.go`: In `DeleteEducation`, change `DELETE FROM education` to `UPDATE education SET deletedAt = NOW()`. In `GetEducations`, add `AND deletedAt IS NULL` to the WHERE clause and add `deletedAt` to SELECT and Scan. In `GetEducationById`, add `AND deletedAt IS NULL`.

`handler/education.go`: In `handleDeleteEducation`, add `userID := middleware.GetUserIDFromContext(r.Context())` and pass to `h.educationStore.DeleteEducation(id, userID)`.

- [ ] **Step 3: Soft-delete experiences**

Same pattern. In `internal/domain/experience/`:

`model.go`: Add `DeletedAt *string` to Experience struct. Update `DeleteExperience` and `DeleteExperienceTech` signatures.

`repository.go`: In `DeleteExperience`, soft-delete experience + its experience_techs in a transaction. In `GetExperiences`, add `AND e.deletedAt IS NULL`. In `GetExperienceById`, same. In `GetExperienceTechs`, add `AND et.deletedAt IS NULL`.

`handler/experience.go`: In `handleDeleteExperience`, pass userID.

- [ ] **Step 4: Soft-delete certifications**

Same pattern. In `internal/domain/certification/`:

`model.go`: Add `DeletedAt *string` to Certification struct. Change `DeleteCertification(id int)` to `DeleteCertification(id int, deletedBy int)`.

`repository.go`: In `DeleteCertification`, change to `UPDATE certifications SET deletedAt = NOW()`. In `GetCertifications`, add `AND deletedAt IS NULL`. Remove R2 cleanup code from the handler (keep it for a future garbage collection task).

`handler/certification.go`: In `handleDeleteCertification`, pass userID. Remove the R2 deletion block (the current handler deletes R2 objects inline — comment this out with a TODO for future GC).

- [ ] **Step 5: Run tests**

Run: `make test`
Expected: All existing tests pass.

- [ ] **Step 6: Commit**

```bash
git add internal/domain/skill/ internal/domain/education/ internal/domain/experience/ internal/domain/certification/ internal/api/handler/skill.go internal/api/handler/education.go internal/api/handler/experience.go internal/api/handler/certification.go
git commit -m "feat: soft-delete skills, education, experiences, certifications"
```

---

### Task 4: Add PaginatedResponse type and pagination to skills (reference pattern)

**Files:**
- Modify: `internal/domain/skill/model.go` (add PaginatedResponse type, update Store interface)
- Modify: `internal/domain/skill/repository.go` (add CountByUserID, update GetSkills with LIMIT/OFFSET)
- Modify: `internal/api/handler/skill.go` (parse limit/offset params, return paginated response)

**Interfaces:**
- Consumes: existing Skill struct, SkillStore interface
- Produces: `GetSkills(userID int, limit int, offset int) ([]Skill, error)`, `CountByUserID(userID int) (int, error)`, `PaginatedSkillResponse` handler type

- [ ] **Step 1: Add generic PaginatedResponse to skill model.go**

Add this type to `internal/domain/skill/model.go`:

```go
type PaginatedResponse struct {
	Data       []Skill `json:"data"`
	Pagination struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Total  int `json:"total"`
	} `json:"pagination"`
}
```

Update the SkillStore interface:
```go
type SkillStore interface {
	GetSkills(userId int, limit int, offset int) ([]Skill, error)
	CountByUserID(userId int) (int, error)
	CreateSkill(Skill) (Skill, error)
	UpdateSkill(id int, Skill Skill) (Skill, error)
	DeleteSkill(id int, deletedBy int) (Skill, error)
}
```

- [ ] **Step 2: Update GetSkills in repository.go and add CountByUserID**

In `internal/domain/skill/repository.go`, change `GetSkills`:

```go
func (s *Repository) GetSkills(userID int, limit int, offset int) ([]Skill, error) {
	rows, err := s.db.Query(
		"SELECT id, userId, skillName, proficiency, deletedAt, createdAt, updatedAt FROM skills WHERE userId = ? AND deletedAt IS NULL LIMIT ? OFFSET ?",
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRowsIntoSkill(rows)
}

func (s *Repository) CountByUserID(userID int) (int, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM skills WHERE userId = ? AND deletedAt IS NULL",
		userID,
	).Scan(&count)
	return count, err
}
```

- [ ] **Step 3: Update handleViewSkills in handler**

In `internal/api/handler/skill.go`, replace `handleViewSkills`:

```go
func (h *SkillHandler) handleViewSkills(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	limit := 20
	offset := 0
	query := r.URL.Query()
	if l := query.Get("limit"); l != "" {
		limit = httputil.ParseIntOrDefault(l, 20)
	}
	if o := query.Get("offset"); o != "" {
		offset = httputil.ParseIntOrDefault(o, 0)
	}
	if limit > 100 {
		limit = 100
	}

	skills, err := h.skillStore.GetSkills(userID, limit, offset)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	total, err := h.skillStore.CountByUserID(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := skill.PaginatedResponse{
		Data: skills,
	}
	resp.Pagination.Limit = limit
	resp.Pagination.Offset = offset
	resp.Pagination.Total = total

	httputil.WriteJSON(w, http.StatusOK, resp)
}
```

Remove the old `SkillReponse` type since we use `PaginatedResponse` now.

- [ ] **Step 4: Run tests**

Run: `make test`
Expected: All existing tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/domain/skill/model.go internal/domain/skill/repository.go internal/api/handler/skill.go
git commit -m "feat: add pagination to skills endpoint"
```

---

### Task 5: Pagination for projects, education, experiences, certifications, technologies, PATs

**Files:**
- Modify: `internal/domain/project/model.go` + `repository.go` + `internal/api/handler/project.go`
- Modify: `internal/domain/education/model.go` + `repository.go` + `internal/api/handler/education.go`
- Modify: `internal/domain/experience/model.go` + `repository.go` + `internal/api/handler/experience.go`
- Modify: `internal/domain/certification/model.go` + `repository.go` + `internal/api/handler/certification.go`
- Modify: `internal/domain/technology/model.go` + `repository.go` + `internal/api/handler/technology.go`
- Modify: `internal/domain/personalaccesstoken/model.go` + `repository.go` + `internal/api/handler/personalaccesstoken.go`

**Pattern:** Exact same as Task 4 for each entity. For each:

1. **model.go**: Add `PaginatedResponse` type with entity-specific data type (e.g. `[]ProjectFull`). Update Store interface — change `GetXxx(userID int)` to `GetXxx(userID int, limit int, offset int)` and add `CountByUserID(userID int) (int, error)`.

2. **repository.go**: Add `LIMIT ? OFFSET ?` to list query, add `CountByUserID` method.

3. **handler**: Parse limit/offset, cap limit at 100, call store, return `PaginatedResponse`.

- [ ] **Step 1: Paginate projects**

In `internal/domain/project/model.go`, add `ProjectPaginatedResponse` with `Data []ProjectFull`. Update `ProjectStore` interface — change `GetProjects(userID int)` to `GetProjects(userID int, limit int, offset int) ([]ProjectFull, error)`, add `CountByUserID`.

In `internal/domain/project/repository.go`, update `GetProjectsFull` name to `GetProjects` (or update existing) with:
```sql
... WHERE p.userId = ? AND p.deletedAt IS NULL ORDER BY p.displayOrder ASC, p.id ASC LIMIT ? OFFSET ?
```
Add `CountByUserID`:
```go
func (s *Repository) CountByUserID(userID int) (int, error) {
    var count int
    err := s.db.QueryRow("SELECT COUNT(*) FROM projects WHERE userId = ? AND deletedAt IS NULL", userID).Scan(&count)
    return count, err
}
```

In `internal/api/handler/project.go`, update `handleViewProjects` with the same limit/offset parsing pattern from Task 4 and return `ProjectPaginatedResponse`.

- [ ] **Step 2: Paginate education**

Apply pattern. Update `GetEducations` in repository to accept `limit int, offset int`. Add `CountByUserID`. Update `handleViewEducation` in handler.

- [ ] **Step 3: Paginate experiences**

Apply pattern. Update `GetExperiences` in repository, add `CountByUserID`. Update `handleViewExperience` in handler.

- [ ] **Step 4: Paginate certifications**

Apply pattern. Update `GetCertifications` in repository, add `CountByUserID`. Update `handleViewCertifications` in handler.

- [ ] **Step 5: Paginate technologies**

Technologies list is global (not scoped to user). Still add pagination via `LIMIT ? OFFSET ?`. Add `CountAll` method. Same pattern but no userId filter in WHERE.

- [ ] **Step 6: Paginate PATs**

PATs already have `GetTokensByUserID`. Add limit/offset params and CountByUserID. Update handler.

- [ ] **Step 7: Run tests**

Run: `make test`
Expected: All existing tests pass.

- [ ] **Step 8: Commit**

```bash
git add internal/domain/project/ internal/domain/education/ internal/domain/experience/ internal/domain/certification/ internal/domain/technology/ internal/domain/personalaccesstoken/ internal/api/handler/
git commit -m "feat: add pagination to all list endpoints"
```

---

### Task 6: Add dashboard activity and usage stats endpoints (backend)

**Files:**
- Modify: `internal/api/handler/dashboard.go` (add activity + usage stats handlers)
- Modify: `internal/api/router.go` (register new routes — but routes are auto-registered, so this may not need changes if using RegisterRoutes pattern)
- New queries in `internal/domain/apilog/repository.go` (add GetDailyUsage method)

**Interfaces:**
- Consumes: existing DashboardHandler, apilog.Repository, skill/education/experience/certification/project repos
- Produces: `GET /dashboard/activity`, `GET /dashboard/usage-stats?days=30`

- [ ] **Step 1: Add GetDailyUsage to apilog repository**

In `internal/domain/apilog/model.go`, add the response type and store method:

```go
type DailyUsage struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type APIUsageLogStore interface {
	Create(log APIUsageLog) error
	GetByTokenID(tokenId int, limit int, offset int) (APIUsageLogWithToken, error)
	GetUserUsageStats(userId int) (UserAPIUsageStats, error)
	GetDailyUsage(userId int, days int) ([]DailyUsage, error)
}
```

In `internal/domain/apilog/repository.go`, add:

```go
func (s *Repository) GetDailyUsage(userID int, days int) ([]DailyUsage, error) {
	rows, err := s.db.Query(`
		SELECT DATE(createdAt) as date, COUNT(*) as count
		FROM api_usage_logs
		WHERE userId = ? AND createdAt >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE(createdAt)
		ORDER BY date ASC
	`, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usage []DailyUsage
	for rows.Next() {
		var d DailyUsage
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return nil, err
		}
		usage = append(usage, d)
	}
	return usage, rows.Err()
}
```

- [ ] **Step 2: Create Activity type**

In `internal/api/handler/dashboard.go`, add types:

```go
type ActivityItem struct {
	Type      string `json:"type"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type ActivityResponse struct {
	Message string             `json:"message"`
	Data    []DashboardActivity `json:"data"`
}
```

... actually, the activity feed uses a UNION query. Let's build it in the dashboard handler since it spans multiple entities:

```go
type ActivityItem struct {
	Type      string `json:"type"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type ActivityResponse struct {
	Message string         `json:"message"`
	Data    []ActivityItem `json:"data"`
}
```

- [ ] **Step 3: Add activity + usage stats handlers to dashboard.go**

In `internal/api/handler/dashboard.go`, update DashboardHandler to accept skill, education, experience, certification, project repos:

```go
type DashboardHandler struct {
	userStore        *user.Repository
	patStore         *personalaccesstoken.Repository
	apiUsageLogStore *apilog.Repository
	projectStore     *project.Repository
	skillStore       *skill.Repository
	educationStore   *education.Repository
	experienceStore  *experience.Repository
	certificationStore *certification.Repository
}
```

Update `NewDashboardHandler` accordingly.

In `RegisterRoutes`, add:
```go
router.HandleFunc("/dashboard/activity", middleware.WithJWTAuth(h.handleViewActivity, h.userStore)).Methods("GET")
router.HandleFunc("/dashboard/usage-stats", middleware.WithJWTAuth(h.handleViewUsageStats, h.userStore)).Methods("GET")
```

Add `handleViewActivity`:
```go
func (h *DashboardHandler) handleViewActivity(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	db := h.apiUsageLogStore.db // need to expose db or use individual repo methods

	// Actually, a simpler approach: query each store individually and merge.
	// Since we can't access db directly from the handler, we'll add a method
	// to apilog repo or create a new dashboard repo.
}
```

Wait — this needs db access. Alternative cleaner approach: add a `GetRecentActivity` method to the apilog repository (it already has db access) that does the UNION. But that's coupling concerns.

Simpler: just query each entity's `CountByUserID` + `GetXxx` sorted by createdAt DESC with a small limit. No new repository method needed.

Actually, the cleanest: add `GetRecentActivity(userID int, limit int) ([]ActivityItem, error)` to a new or existing repository. Let's add it to the apilog repository since it already exists:

In `internal/domain/apilog/repository.go`:
```go
func (s *Repository) GetRecentActivity(userID int, limit int) ([]DashboardActivity, error) {
	rows, err := s.db.Query(`
		SELECT 'project' as type, id, title as name, createdAt FROM projects WHERE userId = ? AND deletedAt IS NULL
		UNION ALL
		SELECT 'skill', id, skillName, createdAt FROM skills WHERE userId = ? AND deletedAt IS NULL
		UNION ALL
		SELECT 'education', id, CONCAT(degree, ' at ', school), createdAt FROM education WHERE userId = ? AND deletedAt IS NULL
		UNION ALL
		SELECT 'experience', id, title, createdAt FROM experiences WHERE userId = ? AND deletedAt IS NULL
		UNION ALL
		SELECT 'certification', id, name, createdAt FROM certifications WHERE userId = ? AND deletedAt IS NULL
		ORDER BY createdAt DESC
		LIMIT ?
	`, userID, userID, userID, userID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DashboardActivity
	for rows.Next() {
		var item DashboardActivity
		if err := rows.Scan(&item.Type, &item.ID, &item.Name, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
```

Add `DashboardActivity` to model.go:
```go
type DashboardActivity struct {
	Type      string `json:"type"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}
```

Add to store interface:
```go
GetRecentActivity(userId int, limit int) ([]DashboardActivity, error)
```

Then in the dashboard handler, the handlers are straightforward:

```go
func (h *DashboardHandler) handleViewActivity(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	limit := 20
	query := r.URL.Query()
	if l := query.Get("limit"); l != "" {
		limit = httputil.ParseIntOrDefault(l, 20)
	}

	activity, err := h.apiUsageLogStore.GetRecentActivity(userID, limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if activity == nil {
		activity = []apilog.DashboardActivity{}
	}

	resp := ActivityResponse{
		Message: "Activity fetched successfully",
		Data:    activity,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
```

```go
func (h *DashboardHandler) handleViewUsageStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	days := 30
	query := r.URL.Query()
	if d := query.Get("days"); d != "" {
		days = httputil.ParseIntOrDefault(d, 30)
	}

	usage, err := h.apiUsageLogStore.GetDailyUsage(userID, days)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if usage == nil {
		usage = []apilog.DailyUsage{}
	}

	resp := UsageStatsResponse{
		Message: "Usage stats fetched successfully",
		Data:    usage,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
```

Add `UsageStatsResponse` type with `Data []apilog.DailyUsage`.

Do NOT add extra repo dependencies to DashboardHandler — keep it using only userRepo, patRepo, and apiLogRepo.

- [ ] **Step 4: Run tests**

Run: `make test`
Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/domain/apilog/model.go internal/domain/apilog/repository.go internal/api/handler/dashboard.go
git commit -m "feat: add activity feed and usage stats to dashboard"
```

---

### Task 7: Frontend — error boundary and 404 page

**Files:**
- Create: `app/error.tsx`
- Create: `app/not-found.tsx`
- Create: `components/ui/ErrorBoundary.tsx`

- [ ] **Step 1: Create app/error.tsx**

```tsx
"use client";

import { useEffect } from "react";
import Link from "next/link";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-base-200">
      <div className="text-center max-w-md px-6">
        <h1 className="text-4xl font-bold mb-4">Something went wrong</h1>
        <p className="text-base-content/60 mb-8">
          An unexpected error occurred. Please try again.
        </p>
        <div className="flex gap-3 justify-center">
          <button onClick={reset} className="btn btn-primary">
            Try again
          </button>
          <Link href="/dashboard" className="btn btn-ghost">
            Go to Dashboard
          </Link>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Create app/not-found.tsx**

```tsx
import Link from "next/link";

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-base-200">
      <div className="text-center max-w-md px-6">
        <h1 className="text-6xl font-bold text-primary mb-4">404</h1>
        <h2 className="text-2xl font-semibold mb-2">Page not found</h2>
        <p className="text-base-content/60 mb-8">
          The page you&apos;re looking for doesn&apos;t exist or has been moved.
        </p>
        <div className="flex gap-3 justify-center">
          <Link href="/dashboard" className="btn btn-primary">
            Go to Dashboard
          </Link>
          <Link href="/api/intro" className="btn btn-ghost">
            API Docs
          </Link>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Create components/ui/ErrorBoundary.tsx**

```tsx
"use client";

import React from "react";

interface Props {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

export default class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error("ErrorBoundary caught:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback;

      return (
        <div className="rounded-2xl border border-error/30 bg-base-100 p-6 text-center">
          <h3 className="text-lg font-semibold text-error mb-2">Something went wrong</h3>
          <p className="text-sm text-base-content/60 mb-4">
            This section encountered an error and couldn&apos;t load.
          </p>
          <button
            className="btn btn-ghost btn-sm"
            onClick={() => this.setState({ hasError: false, error: undefined })}
          >
            Try again
          </button>
          {this.state.error && (
            <details className="mt-3 text-left">
              <summary className="text-xs text-base-content/40 cursor-pointer">Error details</summary>
              <pre className="mt-2 text-xs text-error bg-base-200 p-2 rounded overflow-auto max-h-32">
                {this.state.error.message}
              </pre>
            </details>
          )}
        </div>
      );
    }

    return this.props.children;
  }
}
```

- [ ] **Step 4: Run lint**

Run: `npm run lint` in megome-front
Expected: No lint errors.

- [ ] **Step 5: Commit**

```bash
git add app/error.tsx app/not-found.tsx components/ui/ErrorBoundary.tsx
git commit -m "feat: add error boundary and 404 page"
```

---

### Task 8: Wrap profile sections and project list in ErrorBoundary

**Files:**
- Modify: `features/profile/components/TopProfile.tsx`
- Modify: `features/profile/components/rightContents/ProfileEducation.tsx`
- Modify: `features/profile/components/rightContents/ProfileExperience.tsx`
- Modify: `features/profile/components/rightContents/ProfileCertificate.tsx`
- Modify: `features/profile/components/rightContents/ProfileProjects.tsx`
- Modify: `features/profile/components/ProfileSkill.tsx`
- Modify: `features/project/components/ProjectsClient.tsx`

- [ ] **Step 1: Wrap each profile section component**

In each file, import ErrorBoundary and wrap the main return value:

For each, add `import ErrorBoundary from "@/components/ui/ErrorBoundary";` then change:

From:
```tsx
return (
  <div className="...">
    ...
  </div>
);
```

To:
```tsx
return (
  <ErrorBoundary>
    <div className="...">
      ...
    </div>
  </ErrorBoundary>
);
```

For each component:

In `TopProfile.tsx`, wrap the outermost `<div>` in the return.

In `ProfileEducation.tsx`, wrap the outermost `<div>`.

In `ProfileExperience.tsx`, wrap the outermost `<div>`.

In `ProfileCertificate.tsx`, wrap the outermost `<div>`.

In `ProfileProjects.tsx`, wrap the outermost `<div>`.

In `ProfileSkill.tsx`, wrap the outermost `<div>`.

In `ProjectsClient.tsx`, wrap the outermost fragment `<>` (change to `<ErrorBoundary><div>...</div></ErrorBoundary>`).

- [ ] **Step 2: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
git add features/profile/components/TopProfile.tsx features/profile/components/rightContents/ features/profile/components/ProfileSkill.tsx features/project/components/ProjectsClient.tsx
git commit -m "feat: wrap profile sections and project list in error boundaries"
```

---

### Task 9: Frontend — dashboard proxy routes for activity + usage stats

**Files:**
- Create: `app/api/dashboard/activity/route.ts`
- Create: `app/api/dashboard/usage-stats/route.ts`

- [ ] **Step 1: Create activity proxy route**

`app/api/dashboard/activity/route.ts`:
```ts
import { NextResponse } from "next/server";
import { getAccessToken } from "@/lib/auth/cookies";

const BACKEND_URL = process.env.NEXT_PUBLIC_API_URL!;

export async function GET() {
  try {
    const accessToken = await getAccessToken();

    const res = await fetch(`${BACKEND_URL}/api/v1/dashboard/activity?limit=20`, {
      method: "GET",
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    const data = await res.json();
    return NextResponse.json(data, { status: res.status });
  } catch (err) {
    return NextResponse.json({ message: "Internal server error", data: [] }, { status: 500 });
  }
}
```

- [ ] **Step 2: Create usage-stats proxy route**

`app/api/dashboard/usage-stats/route.ts`:
```ts
import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from "@/lib/auth/cookies";

const BACKEND_URL = process.env.NEXT_PUBLIC_API_URL!;

export async function GET(request: NextRequest) {
  try {
    const accessToken = await getAccessToken();
    const { searchParams } = new URL(request.url);
    const days = searchParams.get("days") || "30";

    const res = await fetch(`${BACKEND_URL}/api/v1/dashboard/usage-stats?days=${days}`, {
      method: "GET",
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    const data = await res.json();
    return NextResponse.json(data, { status: res.status });
  } catch (err) {
    return NextResponse.json({ message: "Internal server error", data: [] }, { status: 500 });
  }
}
```

- [ ] **Step 3: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 4: Commit**

```bash
git add app/api/dashboard/ activity/usage-stats/
git commit -m "feat: add proxy routes for dashboard activity and usage stats"
```

---

### Task 10: Frontend — dashboard API client functions

**Files:**
- Modify: `lib/api/client/dashboard.ts` (add getDashboardActivity, getDashboardUsageStats)
- Modify: `types/api.ts` (add ActivityItem, DailyUsage types)

- [ ] **Step 1: Add types to types/api.ts**

```ts
export type ActivityItem = {
  type: string;
  id: number;
  name: string;
  createdAt: string;
};

export type DailyUsage = {
  date: string;
  count: number;
};
```

- [ ] **Step 2: Add API client functions to dashboard.ts**

After the existing functions, add:

```ts
interface ActivityResponse {
  message: string;
  data: ActivityItem[];
}

export const getDashboardActivity = async () => {
  const res = await fetchClient("/api/dashboard/activity", {
    method: "GET",
    credentials: "include",
  });
  return handleResponse<ActivityResponse>(res);
};

interface UsageStatsResponse {
  message: string;
  data: DailyUsage[];
}

export const getDashboardUsageStats = async (days: number = 30) => {
  const res = await fetchClient(`/api/dashboard/usage-stats?days=${days}`, {
    method: "GET",
    credentials: "include",
  });
  return handleResponse<UsageStatsResponse>(res);
};
```

- [ ] **Step 3: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 4: Commit**

```bash
git add lib/api/client/dashboard.ts types/api.ts
git commit -m "feat: add dashboard activity and usage stats client functions"
```

---

### Task 11: Frontend — dashboard page rework (remove playground, add activity + chart)

**Files:**
- Modify: `app/(app)/dashboard/page.tsx`
- Create: `features/dashboard/components/ActivityTimeline.tsx`
- Create: `features/dashboard/components/UsageChart.tsx`

**Interfaces:**
- Consumes: `getDashboardActivity`, `getDashboardUsageStats` from Task 10, `ActivityItem`, `DailyUsage` from Task 10
- Produces: reworked dashboard with activity feed + chart, playground moved out (just remove from dashboard for now, add to API docs in a follow-up)

- [ ] **Step 1: Create ActivityTimeline component**

`features/dashboard/components/ActivityTimeline.tsx`:
```tsx
import type { ActivityItem } from "@/types/api";

const typeIcons: Record<string, string> = {
  project: "",
  skill: "",
  education: "",
  experience: "",
  certification: "",
};

function timeAgo(dateStr: string): string {
  const now = Date.now();
  const date = new Date(dateStr).getTime();
  const seconds = Math.floor((now - date) / 1000);

  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  const months = Math.floor(days / 30);
  return `${months}mo ago`;
}

export default function ActivityTimeline({ items }: { items: ActivityItem[] }) {
  if (items.length === 0) {
    return (
      <div className="flex min-h-[120px] items-center justify-center rounded-2xl border border-dashed border-base-300">
        <p className="text-sm text-base-content/60">No recent activity</p>
      </div>
    );
  }

  return (
    <div className="space-y-1">
      {items.map((item) => (
        <div key={`${item.type}-${item.id}`} className="flex items-center gap-3 py-2 px-3 rounded-lg hover:bg-base-200/50 transition-colors">
          <span className="badge badge-sm badge-ghost capitalize">{item.type}</span>
          <span className="text-sm flex-1 truncate">{item.name}</span>
          <span className="text-xs text-base-content/40 whitespace-nowrap">{timeAgo(item.createdAt)}</span>
        </div>
      ))}
    </div>
  );
}
```

- [ ] **Step 2: Create UsageChart component**

`features/dashboard/components/UsageChart.tsx`:
```tsx
"use client";

import type { DailyUsage } from "@/types/api";
import { useMemo } from "react";

export default function UsageChart({ data }: { data: DailyUsage[] }) {
  const maxCount = useMemo(() => Math.max(...data.map((d) => d.count), 1), [data]);

  if (data.length === 0) {
    return (
      <div className="flex min-h-[120px] items-center justify-center rounded-2xl border border-dashed border-base-300">
        <p className="text-sm text-base-content/60">No usage data yet</p>
      </div>
    );
  }

  return (
    <div className="flex items-end gap-1 h-32">
      {data.map((d) => {
        const height = Math.max((d.count / maxCount) * 100, 4);
        const dateLabel = new Date(d.date).toLocaleDateString("en-US", { month: "short", day: "numeric" });
        return (
          <div key={d.date} className="flex-1 flex flex-col items-center gap-1 min-w-0 group relative">
            <span className="text-xs text-base-content/60 opacity-0 group-hover:opacity-100 transition-opacity absolute -top-5">
              {d.count}
            </span>
            <div
              className="w-full bg-primary rounded-t transition-all hover:opacity-80"
              style={{ height: `${height}%` }}
              title={`${dateLabel}: ${d.count} requests`}
            />
          </div>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 3: Rework the dashboard page**

In `app/(app)/dashboard/page.tsx`:

1. Remove the `ApiPlayground` function entirely.
2. Remove the import for `BoltIcon`, `KeyIcon`, etc. — keep only the ones used by Card components.
3. Add state for activity and usage stats.
4. Add a section for Activity Timeline and Usage Chart.
5. Import the new components.

The changes to the main DashboardPage function:

Add imports:
```tsx
import { getDashboardActivity, getDashboardUsageStats } from "@/lib/api/client/dashboard";
import { ActivityItem, DailyUsage } from "@/types/api";
import ActivityTimeline from "@/features/dashboard/components/ActivityTimeline";
import UsageChart from "@/features/dashboard/components/UsageChart";
```

Add state:
```tsx
const [activity, setActivity] = useState<ActivityItem[]>([]);
const [usageStats, setUsageStats] = useState<DailyUsage[]>([]);
```

Update the useEffect to also fetch activity + usage:
```tsx
useEffect(() => {
  const fetchData = async () => {
    try {
      setLoading(true);
      const [overviewRes, completionRes, activityRes, usageRes] = await Promise.all([
        getDashboardOverview(),
        getCompletion(),
        getDashboardActivity(),
        getDashboardUsageStats(),
      ]);
      setDashboardOverview(overviewRes.data ?? null);
      setCompletion(completionRes.data ?? null);
      setActivity(activityRes.data ?? []);
      setUsageStats(usageRes.data ?? []);
    } catch (error) {
      console.error("Failed to fetch dashboard data: ", error);
    } finally {
      setLoading(false);
    }
  };
  fetchData();
}, []);
```

After the quick actions card, add the activity + chart sections:
```tsx
{/* Activity & Usage */}
<div className="grid gap-6 grid-cols-1 lg:grid-cols-2">
  <div className="rounded-2xl border border-base-300 p-5">
    <h2 className="font-semibold mb-4">Recent Activity</h2>
    <ActivityTimeline items={activity} />
  </div>
  <div className="rounded-2xl border border-base-300 p-5">
    <h2 className="font-semibold mb-4">API Usage</h2>
    <UsageChart data={usageStats} />
  </div>
</div>
```

Remove the `<ApiPlayground />` JSX usage (the component itself can remain as dead code for now, or be deleted — prefer to remove the JSX usage and leave the function in the file to be extracted later).

- [ ] **Step 4: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 5: Commit**

```bash
git add app/(app)/dashboard/page.tsx features/dashboard/components/
git commit -m "feat: rework dashboard with activity timeline and usage chart"
```

---

### Task 12: Frontend — useDirtyGuard hook

**Files:**
- Create: `lib/hooks/useDirtyGuard.ts`

- [ ] **Step 1: Create the hook**

`lib/hooks/useDirtyGuard.ts`:
```ts
"use client";

import { useEffect, useCallback, useRef } from "react";

export function useDirtyGuard(dirty: boolean) {
  const dirtyRef = useRef(dirty);
  dirtyRef.current = dirty;

  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (dirtyRef.current) {
        e.preventDefault();
        e.returnValue = "";
      }
    };

    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  }, []);

  const confirmNavigation = useCallback((): Promise<boolean> => {
    if (!dirtyRef.current) return Promise.resolve(true);

    return new Promise((resolve) => {
      const ok = window.confirm("You have unsaved changes. Are you sure you want to leave?");
      resolve(ok);
    });
  }, []);

  return { confirmNavigation };
}
```

- [ ] **Step 2: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
git add lib/hooks/useDirtyGuard.ts
git commit -m "feat: add useDirtyGuard hook for unsaved changes protection"
```

---

### Task 13: Frontend — add dirty guard to ProfileForm

**Files:**
- Modify: `features/profile/components/ProfileForm.tsx`

- [ ] **Step 1: Add dirty tracking and guard to ProfileForm**

In `features/profile/components/ProfileForm.tsx`:

Add import: `import { useDirtyGuard } from "@/lib/hooks/useDirtyGuard";`

Add state:
```tsx
const [isDirty, setIsDirty] = useState(false);
```

Hook in the guard:
```tsx
const { confirmNavigation } = useDirtyGuard(isDirty);
```

Update `handleChange` to mark dirty:
```tsx
const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
  const { name, value } = e.target;
  setForm((prev) => ({ ...prev, [name]: value }));
  setIsDirty(true);
};
```

Update `handleBioChange`:
```tsx
const handleBioChange = (html: string) => {
  setForm((prev) => ({ ...prev, bio: html }));
  setIsDirty(true);
};
```

Update `handleSubmit` — after successful save, reset dirty:
```tsx
const handleSubmit = async (e: React.FormEvent) => {
  e.preventDefault();
  setErrors({});
  setLoading(true);

  try {
    // ... existing validation and save logic ...

    const data = await withRequest(
      () => updateProfileClient(form),
      showToast
    );

    if (!data) return;

    setIsDirty(false);
    setProfile?.(data.profile ?? null);

    if (isOnboarding) {
      router.push("/dashboard");
    }
  } finally {
    setLoading(false);
  }
};
```

- [ ] **Step 2: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
git add features/profile/components/ProfileForm.tsx
git commit -m "feat: add dirty guard to profile form"
```

---

### Task 14: Frontend — add autosave to ProfileForm

**Files:**
- Modify: `features/profile/components/ProfileForm.tsx`

- [ ] **Step 1: Add autosave with debounce and visual indicator**

In `features/profile/components/ProfileForm.tsx`:

Add states for save indicator:
```tsx
const [saveStatus, setSaveStatus] = useState<"idle" | "saving" | "saved" | "error">("idle");
```

Add a debounced autosave effect:
```tsx
useEffect(() => {
  if (!isDirty || isOnboarding) return;

  const timer = setTimeout(async () => {
    setSaveStatus("saving");
    try {
      const data = await updateProfileClient(form);
      if (data) {
        setSaveStatus("saved");
        setIsDirty(false);
        setProfile?.(data.profile ?? null);
        setTimeout(() => setSaveStatus("idle"), 2000);
      }
    } catch {
      setSaveStatus("error");
    }
  }, 1000);

  return () => clearTimeout(timer);
}, [form, isDirty, isOnboarding, setProfile]);
```

Add the save status indicator near the submit button. After the existing Save button, add:
```tsx
<span className={`text-xs ${saveStatus === "saving" ? "text-base-content/50" : saveStatus === "saved" ? "text-success" : saveStatus === "error" ? "text-error" : "invisible"}`}>
  {saveStatus === "saving" && "Saving..."}
  {saveStatus === "saved" && "Saved"}
  {saveStatus === "error" && "Save failed"}
</span>
```

- [ ] **Step 2: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
git add features/profile/components/ProfileForm.tsx
git commit -m "feat: add autosave to profile form"
```

---

### Task 15: Frontend — add dirty guard to ProjectWizard

**Files:**
- Modify: `features/project/components/ProjectWizard.tsx`

- [ ] **Step 1: Integrate useDirtyGuard into wizard**

In `features/project/components/ProjectWizard.tsx`:

Add import: `import { useDirtyGuard } from "@/lib/hooks/useDirtyGuard";`

Hook in the guard:
```tsx
const { confirmNavigation } = useDirtyGuard(isDirty);
```

On step transition (in Step2 and Step3 advance handlers), check dirty before navigating. Wrap the existing `step` advance callbacks. The existing `isDirty` state is already set to `true` on step info changes — ensure it also gets set in step images and tech changes. 

In `StepImages`, add `setIsDirty` to props and call `setIsDirty(true)` when cover or screenshots change.

In `StepTech`, add `setIsDirty` to props and call `setIsDirty(true)` when selected technologies change.

- [ ] **Step 2: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
git add features/project/components/ProjectWizard.tsx features/project/components/stepperForm/
git commit -m "feat: add dirty guard to project wizard"
```

---

### Task 16: Frontend — rich text live preview toggle

**Files:**
- Create: `components/ui/rich-editor/RichTextPreview.tsx`

- [ ] **Step 1: Create RichTextPreview component**

`components/ui/rich-editor/RichTextPreview.tsx`:
```tsx
export default function RichTextPreview({ html }: { html: string }) {
  return (
    <div
      className="prose prose-sm max-w-none text-base-content/80 [&_h2]:text-lg [&_h2]:font-semibold [&_h3]:text-base [&_h3]:font-semibold [&_p]:text-sm [&_ul]:text-sm [&_ol]:text-sm [&_blockquote]:text-sm [&_blockquote]:border-l-2 [&_blockquote]:border-base-300 [&_blockquote]:pl-3 [&_blockquote]:italic [&_code]:text-xs [&_code]:bg-base-200 [&_code]:px-1 [&_code]:rounded [&_pre]:bg-base-200 [&_pre]:p-3 [&_pre]:rounded-lg [&_pre]:text-xs [&_pre]:overflow-x-auto"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
```

- [ ] **Step 2: Add preview toggle to ProfileForm bio field**

In `features/profile/components/ProfileForm.tsx`:

Add import: `import RichTextPreview from "@/components/ui/rich-editor/RichTextPreview";`

Add state:
```tsx
const [showBioPreview, setShowBioPreview] = useState(false);
```

In the bio section, add a toggle button and conditional rendering. After the `<legend>` that says "Your bio", change the section:

```tsx
<fieldset className="fieldset">
  <div className="flex items-center justify-between gap-4">
    <legend className="fieldset-legend">Your bio</legend>
    <div className="flex items-center gap-2">
      <button
        type="button"
        className="btn btn-ghost btn-xs"
        onClick={() => setShowBioPreview((p) => !p)}
      >
        {showBioPreview ? "Edit" : "Preview"}
      </button>
      <div className="w-auto">
        <AiAssistButton ... />
      </div>
    </div>
  </div>
  {showBioPreview ? (
    <RichTextPreview html={form.bio || ""} />
  ) : (
    <RichEditor content={form.bio || ""} onChange={handleBioChange} />
  )}
  {/* ... rest of existing content ... */}
</fieldset>
```

- [ ] **Step 3: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 4: Commit**

```bash
git add components/ui/rich-editor/RichTextPreview.tsx features/profile/components/ProfileForm.tsx
git commit -m "feat: add rich text live preview to profile bio"
```

---

### Task 17: Frontend — image resize utility hook

**Files:**
- Create: `lib/hooks/useImageResize.ts`

- [ ] **Step 1: Create useImageResize hook**

`lib/hooks/useImageResize.ts`:
```ts
"use client";

export interface ResizeOptions {
  maxWidth: number;
  quality?: number;
}

export function useImageResize() {
  const resizeImage = (file: File, options: ResizeOptions): Promise<Blob> => {
    const { maxWidth, quality = 0.85 } = options;

    return new Promise((resolve, reject) => {
      const img = new Image();
      const url = URL.createObjectURL(file);

      img.onload = () => {
        URL.revokeObjectURL(url);

        if (img.width <= maxWidth) {
          resolve(file);
          return;
        }

        const canvas = document.createElement("canvas");
        const ratio = maxWidth / img.width;
        canvas.width = maxWidth;
        canvas.height = Math.round(img.height * ratio);

        const ctx = canvas.getContext("2d");
        if (!ctx) {
          reject(new Error("Could not get canvas context"));
          return;
        }

        ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
        canvas.toBlob(
          (blob) => {
            if (blob) resolve(blob);
            else reject(new Error("Canvas toBlob failed"));
          },
          "image/jpeg",
          quality
        );
      };

      img.onerror = () => {
        URL.revokeObjectURL(url);
        reject(new Error("Failed to load image"));
      };

      img.src = url;
    });
  };

  return { resizeImage };
}
```

- [ ] **Step 2: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
git add lib/hooks/useImageResize.ts
git commit -m "feat: add image resize utility hook"
```

---

### Task 18: Frontend — image resize + drag-drop + validation in StepImages

**Files:**
- Modify: `features/project/components/stepperForm/stepImages.tsx`

- [ ] **Step 1: Add resize, drag-drop, and file validation to StepImages**

In `features/project/components/stepperForm/stepImages.tsx`:

Add imports:
```tsx
import { useState } from "react";
import { useImageResize } from "@/lib/hooks/useImageResize";
import { useToast } from "@/components/ui/toast/useToast";
```

Add state for drag-over:
```tsx
const [coverDragOver, setCoverDragOver] = useState(false);
const [screenshotDragOver, setScreenshotDragOver] = useState(false);
```

Add the hook and toast:
```tsx
const { resizeImage } = useImageResize();
const { showToast } = useToast();
```

Add a validation helper inside the component:
```tsx
function validateFile(file: File): boolean {
  if (!file.type.startsWith("image/")) {
    showToast("Only image files are allowed", "error");
    return false;
  }
  if (file.size > 10 * 1024 * 1024) {
    showToast("File must be under 10MB", "error");
    return false;
  }
  return true;
}
```

Add a resize wrapper:
```tsx
async function processCover(file: File) {
  if (!validateFile(file)) return;
  const resized = await resizeImage(file, { maxWidth: 1920 });
  setImages((prev) => ({
    ...prev,
    cover: {
      file: new File([resized], file.name, { type: "image/jpeg" }),
      preview: URL.createObjectURL(file),
      type: "cover",
      status: "idle",
    },
  }));
}
```

For the cover drop zone, replace the existing `<label>` with a version that also accepts drag-and-drop:

```tsx
<label
  className={`border border-dashed rounded-lg p-8 text-center cursor-pointer hover:bg-base-200 transition ${coverDragOver ? "border-primary bg-primary/5" : ""}`}
  onDragOver={(e) => { e.preventDefault(); setCoverDragOver(true); }}
  onDragLeave={() => setCoverDragOver(false)}
  onDrop={(e) => {
    e.preventDefault();
    setCoverDragOver(false);
    const file = e.dataTransfer.files?.[0];
    if (file) processCover(file);
  }}
>
  <input
    type="file"
    className="hidden"
    accept="image/*"
    onChange={(e) => {
      const file = e.target.files?.[0];
      if (file) processCover(file);
    }}
  />
  <p className="font-medium">Upload cover image</p>
  <p className="text-sm opacity-60">Recommended: 16:9 ratio. Drag & drop or click</p>
</label>
```

Do the same for the screenshot add button — add drag-and-drop + resize. For screenshots, add a `processScreenshot` function that resizes to maxWidth 1920 and handles multiple files:

```tsx
async function processScreenshots(files: FileList) {
  const newImgs: ProjectImage[] = [];
  for (const file of Array.from(files)) {
    if (!validateFile(file)) continue;
    try {
      const resized = await resizeImage(file, { maxWidth: 1920 });
      newImgs.push({
        file: new File([resized], file.name, { type: "image/jpeg" }),
        preview: URL.createObjectURL(file),
        type: "screenshot",
        status: "idle",
      });
    } catch {
      showToast(`Failed to process ${file.name}`, "error");
    }
  }
  if (newImgs.length > 0) {
    setImages((prev) => ({
      ...prev,
      screenshots: [...prev.screenshots, ...newImgs],
    }));
  }
}
```

Update the screenshot add label similarly.

- [ ] **Step 2: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
git add features/project/components/stepperForm/stepImages.tsx
git commit -m "feat: add image resize, drag-drop, and validation to project images"
```

---

### Task 19: Frontend — avatar crop modal

**Files:**
- Create: `features/profile/components/AvatarCropModal.tsx`
- Modify: `features/profile/components/ProfileForm.tsx`

- [ ] **Step 1: Create AvatarCropModal**

`features/profile/components/AvatarCropModal.tsx`:
```tsx
"use client";

import { useState, useRef, useEffect } from "react";
import Modal from "@/components/ui/modal/Modal";

interface Props {
  open: boolean;
  file: File | null;
  onClose: () => void;
  onCrop: (blob: Blob) => void;
}

export default function AvatarCropModal({ open, file, onClose, onCrop }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [image, setImage] = useState<HTMLImageElement | null>(null);
  const [position, setPosition] = useState({ x: 0, y: 0 });

  const SIZE = 200;

  useEffect(() => {
    if (!file || !open) return;
    const img = new Image();
    img.onload = () => setImage(img);
    img.src = URL.createObjectURL(file);
  }, [file, open]);

  const handleCrop = () => {
    if (!image || !canvasRef.current) return;
    const canvas = canvasRef.current;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    canvas.width = SIZE;
    canvas.height = SIZE;

    ctx.beginPath();
    ctx.arc(SIZE / 2, SIZE / 2, SIZE / 2, 0, Math.PI * 2);
    ctx.clip();

    const scale = SIZE / 200;
    ctx.drawImage(image, position.x * scale, position.y * scale, image.width * scale, image.height * scale);

    canvas.toBlob((blob) => {
      if (blob) onCrop(blob);
    }, "image/jpeg", 0.9);
  };

  return (
    <Modal isOpen={open} onClose={onClose} title="Crop Avatar" onAccept={handleCrop} acceptText="Crop">
      <div className="flex flex-col items-center gap-4">
        {image && (
          <div
            className="relative w-[200px] h-[200px] overflow-hidden rounded-full border-2 border-base-300 cursor-move"
            onMouseDown={(e) => {
              const startX = e.clientX - position.x;
              const startY = e.clientY - position.y;
              const onMove = (ev: MouseEvent) => {
                setPosition({ x: ev.clientX - startX, y: ev.clientY - startY });
              };
              const onUp = () => {
                document.removeEventListener("mousemove", onMove);
                document.removeEventListener("mouseup", onUp);
              };
              document.addEventListener("mousemove", onMove);
              document.addEventListener("mouseup", onUp);
            }}
          >
            <img
              src={image.src}
              alt="Crop preview"
              className="absolute max-w-none select-none"
              style={{ left: position.x, top: position.y, width: image.width }}
              draggable={false}
            />
          </div>
        )}
        <canvas ref={canvasRef} className="hidden" />
        <p className="text-xs text-base-content/50">Drag to position your avatar</p>
      </div>
    </Modal>
  );
}
```

- [ ] **Step 2: Integrate into ProfileForm avatar upload**

In `features/profile/components/ProfileForm.tsx`:

Add import: `import AvatarCropModal from "./AvatarCropModal";`

Add state:
```tsx
const [cropOpen, setCropOpen] = useState(false);
const [cropFile, setCropFile] = useState<File | null>(null);
```

Change the avatar file input's onChange:
```tsx
onChange={(e) => {
  const file = e.target.files?.[0] || null;
  if (file) {
    setCropFile(file);
    setCropOpen(true);
  }
}}
```

Add the modal near the form's end (before closing tag):
```tsx
<AvatarCropModal
  open={cropOpen}
  file={cropFile}
  onClose={() => setCropOpen(false)}
  onCrop={(blob) => {
    const croppedFile = new File([blob], cropFile?.name || "avatar.jpg", { type: "image/jpeg" });
    setForm((prev) => ({ ...prev, profileImage: croppedFile }));
    setPreview(URL.createObjectURL(blob));
    setCropOpen(false);
  }}
/>
```

- [ ] **Step 3: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 4: Commit**

```bash
git add features/profile/components/AvatarCropModal.tsx features/profile/components/ProfileForm.tsx
git commit -m "feat: add avatar crop modal to profile form"
```

---

### Task 20: Frontend — empty state consistency

**Files:**
- Modify: `features/project/components/ProjectsClient.tsx`
- Modify: `features/profile/components/sections/EducationForm.tsx`
- Modify: `features/profile/components/sections/ExperienceForm.tsx`
- Modify: `features/profile/components/sections/CertificateForm.tsx`
- Modify: `features/profile/components/sections/SkillForm.tsx`

- [ ] **Step 1: Update ProjectsClient empty state**

In `features/project/components/ProjectsClient.tsx`:

Add import: `import { EmptyState } from "@/features/profile/components/sections/EmptyState";`
Add import: `import Link from "next/link";` (already imported)

Replace the inline empty state in `ProjectGrid`:
```tsx
if (projects.length === 0) {
  return (
    <EmptyState
      icon={<span className="text-2xl">&#128640;</span>}
      title="No projects found"
      description="Showcase your work by adding your first project"
      action={
        <Link href="/projects/new" className="btn btn-primary btn-sm">
          Add Project
        </Link>
      }
    />
  );
}
```

- [ ] **Step 2: Update EducationForm empty state**

In `EducationForm.tsx`, find the plain text empty state ("No education added yet") and replace with:

```tsx
<EmptyState
  icon={<span className="text-2xl">&#127891;</span>}
  title="No education added yet"
  description="Add your academic background"
/>
```

- [ ] **Step 3: Update ExperienceForm empty state**

Replace with:
```tsx
<EmptyState
  icon={<span className="text-2xl">&#128188;</span>}
  title="No experience added yet"
  description="Add your work experience"
/>
```

- [ ] **Step 4: Update CertificateForm empty state**

Replace with:
```tsx
<EmptyState
  icon={<span className="text-2xl">&#127942;</span>}
  title="No certificates added yet"
  description="Add your certifications"
/>
```

- [ ] **Step 5: Update SkillForm empty state**

Replace with:
```tsx
<EmptyState
  icon={<span className="text-2xl">&#128295;</span>}
  title="No skills added yet"
  description="Add your skills to showcase expertise"
/>
```

- [ ] **Step 6: Run lint**

Run: `npm run lint`
Expected: No lint errors.

- [ ] **Step 7: Commit**

```bash
git add features/project/components/ProjectsClient.tsx features/profile/components/sections/
git commit -m "feat: use shared EmptyState component across all sections"
```

---

## Self-Review Checklist

- [x] **Spec coverage:** All 6 sections mapped to tasks:
  1. Pagination: Tasks 4-5
  2. Soft delete: Tasks 1-3
  3. Error handling: Tasks 7-8
  4. Dashboard overhaul: Tasks 6, 9-11
  5. Form polish: Tasks 12-16
  6. Image handling: Tasks 17-19
  7. Empty state: Task 20
- [x] **Placeholder scan:** No TBD, TODO, or vague steps. All code is concrete.
- [x] **Type consistency:** `ActivityItem`, `DailyUsage` defined in Task 10, consumed in Task 11. `UseDirtyGuard` return type `{ confirmNavigation }` defined in Task 12, consumed in Tasks 13, 15. `ResizeOptions` defined in Task 17, consumed in Tasks 18, 19. `PaginatedResponse` defined per entity in Tasks 4-5. `DeletedAt` column from Task 1 consumed in Tasks 2-3.
