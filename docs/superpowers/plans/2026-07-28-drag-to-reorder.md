# Drag-to-Reorder Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a dedicated `/reorder` page with drag-and-drop sorting for experiences, certificates, education, and projects.

**Architecture:** Add `displayOrder` column to 3 tables (experiences, certifications, education), wire up the existing dead column in projects. Backend gets `PATCH /api/v1/{resource}/reorder` endpoints. Frontend gets a new `/reorder` page using `@dnd-kit` with tabbed interface.

**Tech Stack:** Go (gorilla/mux, MySQL), Next.js 16 (React 19, TypeScript, Tailwind CSS, DaisyUI v5), `@dnd-kit/core` + `@dnd-kit/sortable`

## Global Constraints

- MySQL database with `golang-migrate` migrations
- Go backend at `/home/alex/my-go/megome/`
- Next.js frontend at `/home/alex/my-go/megome-front/`
- Handlers live in `internal/api/handler/` (NOT in domain directories)
- Repositories live in `internal/domain/{domain}/repository.go`
- Models live in `internal/domain/{domain}/model.go`
- Router at `internal/api/router.go`
- Migration naming: `YYYYMMDDHHMMSS_description.{up,down}.sql`
- CORS allows: GET, POST, PUT, DELETE (not PATCH) — use POST for reorder endpoint
- Frontend API pattern: `fetchClient` + `handleResponse<T>`
- Frontend types in `types/domain.ts`
- Sidebar menu in `components/ui/Sidebar.tsx`

---

## Task 1: Database Migrations

**Files:**
- Create: `cmd/migrate/migrations/20260728120000_add-display-order-to-experiences.up.sql`
- Create: `cmd/migrate/migrations/20260728120000_add-display-order-to-experiences.down.sql`
- Create: `cmd/migrate/migrations/20260728120001_add-display-order-to-certifications.up.sql`
- Create: `cmd/migrate/migrations/20260728120001_add-display-order-to-certifications.down.sql`
- Create: `cmd/migrate/migrations/20260728120002_add-display-order-to-education.up.sql`
- Create: `cmd/migrate/migrations/20260728120002_add-display-order-to-education.down.sql`

- [ ] **Step 1: Create experiences migration (up)**

```sql
ALTER TABLE experiences ADD COLUMN displayOrder INT UNSIGNED NOT NULL DEFAULT 0;
```

- [ ] **Step 2: Create experiences migration (down)**

```sql
ALTER TABLE experiences DROP COLUMN displayOrder;
```

- [ ] **Step 3: Create certifications migration (up)**

```sql
ALTER TABLE certifications ADD COLUMN displayOrder INT UNSIGNED NOT NULL DEFAULT 0;
```

- [ ] **Step 4: Create certifications migration (down)**

```sql
ALTER TABLE certifications DROP COLUMN displayOrder;
```

- [ ] **Step 5: Create education migration (up)**

```sql
ALTER TABLE education ADD COLUMN displayOrder INT UNSIGNED NOT NULL DEFAULT 0;
```

- [ ] **Step 6: Create education migration (down)**

```sql
ALTER TABLE education DROP COLUMN displayOrder;
```

- [ ] **Step 7: Commit**

```bash
git add cmd/migrate/migrations/
git commit -m "feat: add displayOrder column to experiences, certifications, education"
```

---

## Task 2: Experience Domain — Model, Repository, Handler

**Files:**
- Modify: `internal/domain/experience/model.go`
- Modify: `internal/domain/experience/repository.go`
- Modify: `internal/api/handler/experience.go`

**Interfaces:**
- Produces: `ReorderItem` struct, `ReorderExperiences` method, `handleReorderExperiences` handler, `PATCH /experience/reorder` route

- [ ] **Step 1: Add DisplayOrder to Experience struct and ReorderItem type**

In `internal/domain/experience/model.go`, add `DisplayOrder` field to `Experience` struct and add shared `ReorderItem` type:

```go
type Experience struct {
	ID           int                      `json:"id"`
	UserID       int                      `json:"userId"`
	Title        string                   `json:"title"`
	Company      string                   `json:"company"`
	Logo         *string                  `json:"logo"`
	StartDate    string                   `json:"startDate"`
	EndDate      *string                  `json:"endDate"`
	IsPresent    bool                     `json:"isPresent"`
	Description  string                   `json:"description"`
	Technologies []technology.Technology  `json:"technologies"`
	DisplayOrder int                      `json:"displayOrder"`
	CreatedAt    string                   `json:"createdAt"`
	UpdatedAt    string                   `json:"updatedAt"`
}

type ReorderItem struct {
	ID           int `json:"id"`
	DisplayOrder int `json:"displayOrder"`
}
```

- [ ] **Step 2: Update experience repository SELECT queries and scan functions**

In `internal/domain/experience/repository.go`:

**Update `GetExperienceById` query** — change SELECT to include `displayOrder`:

```go
row := s.db.QueryRow("SELECT id, userId, title, company, logo, startDate, endDate, isPresent, description, displayOrder, createdAt, updatedAt FROM experiences WHERE id = ?", id)
```

**Update `getExperiences` query** — change SELECT to include `displayOrder`:

```go
rows, err := s.db.Query(
	"SELECT id, userId, title, company, logo, startDate, endDate, isPresent, description, displayOrder, createdAt, updatedAt FROM experiences WHERE userId = ?",
	userID,
)
```

**Update `scanRowIntoExperience`** — add `displayOrder` to Scan:

```go
func scanRowIntoExperience(scanner interface{ Scan(dest ...interface{}) error }) (Experience, error) {
	var experience Experience
	err := scanner.Scan(
		&experience.ID,
		&experience.UserID,
		&experience.Title,
		&experience.Company,
		&experience.Logo,
		&experience.StartDate,
		&experience.EndDate,
		&experience.IsPresent,
		&experience.Description,
		&experience.DisplayOrder,
		&experience.CreatedAt,
		&experience.UpdatedAt,
	)
	if err != nil {
		return Experience{}, err
	}

	if experience.Logo != nil && *experience.Logo != "" {
		if !strings.HasPrefix(*experience.Logo, "https://") && !strings.HasPrefix(*experience.Logo, "http://") {
			*experience.Logo = httputil.GetPublicFile(*experience.Logo)
		}
	}

	return experience, nil
}
```

**Add `ReorderExperiences` method** at the end of the file:

```go
func (s *Repository) ReorderExperiences(userID int, items []ReorderItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE experiences SET displayOrder = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ? AND userId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(item.DisplayOrder, item.ID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
```

- [ ] **Step 3: Add reorder handler to experience handler**

In `internal/api/handler/experience.go`, add the route in `RegisterRoutes`. **IMPORTANT:** The `/experience/reorder` route MUST be registered BEFORE `/experience/{id}` to prevent mux from matching "reorder" as an `{id}` parameter:

```go
func (h *ExperienceHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/experience", middleware.WithJWTAuth(h.handleViewExperiences, h.userStore)).Methods("GET")
	router.HandleFunc("/experience/reorder", middleware.WithJWTAuth(h.handleReorderExperiences, h.userStore)).Methods("POST")
	router.HandleFunc("/experience/{id}", middleware.WithJWTAuth(h.handleViewExperienceById, h.userStore)).Methods("GET")
	router.HandleFunc("/experience", middleware.WithJWTAuth(h.handleCreateExperience, h.userStore)).Methods("POST")
	router.HandleFunc("/experience/{id}", middleware.WithJWTAuth(h.handleEditExperience, h.userStore)).Methods("PUT")
	router.HandleFunc("/experience/{id}", middleware.WithJWTAuth(h.handleDeleteExperience, h.userStore)).Methods("DELETE")
}
```

Add the handler method:

```go
func (h *ExperienceHandler) handleReorderExperiences(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Items []experience.ReorderItem `json:"items"`
	}
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if len(payload.Items) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("items list cannot be empty"))
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())

	if err := h.experienceStore.ReorderExperiences(userID, payload.Items); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Experiences reordered successfully"})
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/domain/experience/ internal/api/handler/experience.go
git commit -m "feat: add displayOrder and reorder endpoint for experiences"
```

---

## Task 3: Certification Domain — Model, Repository, Handler

**Files:**
- Modify: `internal/domain/certification/model.go`
- Modify: `internal/domain/certification/repository.go`
- Modify: `internal/api/handler/certification.go`

- [ ] **Step 1: Add DisplayOrder to Certification struct**

In `internal/domain/certification/model.go`, add field to `Certification`:

```go
type Certification struct {
	ID               int     `json:"id"`
	UserID           int     `json:"userId"`
	Title            string  `json:"title"`
	Issuer           string  `json:"issuer"`
	IssueDate        string  `json:"issueDate"`
	CertificateImage *string `json:"certificateImage"`
	ExpirationDate   *string `json:"expirationDate"`
	CredentialId     *string `json:"credentialId"`
	CredentialUrl    *string `json:"credentialUrl"`
	DisplayOrder     int     `json:"displayOrder"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}
```

- [ ] **Step 2: Update certification repository SELECT queries and scan function**

In `internal/domain/certification/repository.go`:

**Update `GetCertificationById` query:**

```go
row := s.db.QueryRow("SELECT id, userId, title, issuer, issueDate, certificateImage, expirationDate, credentialId, credentialUrl, displayOrder, createdAt, updatedAt FROM certifications WHERE id = ?", id)
```

**Update `GetCertifications` query:**

```go
rows, err := s.db.Query(
	"SELECT id, userId, title, issuer, issueDate, certificateImage, expirationDate, credentialId, credentialUrl, displayOrder, createdAt, updatedAt FROM certifications WHERE userId = ?",
	userId,
)
```

**Update `scanRowIntoCertification`** — add `displayOrder`:

```go
func scanRowIntoCertification(scanner interface{ Scan(dest ...interface{}) error }) (Certification, error) {
	var certification Certification
	err := scanner.Scan(
		&certification.ID,
		&certification.UserID,
		&certification.Title,
		&certification.Issuer,
		&certification.IssueDate,
		&certification.CertificateImage,
		&certification.ExpirationDate,
		&certification.CredentialId,
		&certification.CredentialUrl,
		&certification.DisplayOrder,
		&certification.CreatedAt,
		&certification.UpdatedAt,
	)
	if err != nil {
		return Certification{}, err
	}

	if certification.CertificateImage != nil && *certification.CertificateImage != "" {
		if !strings.HasPrefix(*certification.CertificateImage, "https://") && !strings.HasPrefix(*certification.CertificateImage, "http://") {
			*certification.CertificateImage = httputil.GetPublicFile(*certification.CertificateImage)
		}
	}

	return certification, nil
}
```

**Add `ReorderCertifications` method** at the end:

```go
func (s *Repository) ReorderCertifications(userID int, items []ReorderItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE certifications SET displayOrder = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ? AND userId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(item.DisplayOrder, item.ID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
```

Add `ReorderItem` type to `internal/domain/certification/model.go`:

```go
type ReorderItem struct {
	ID           int `json:"id"`
	DisplayOrder int `json:"displayOrder"`
}
```

- [ ] **Step 3: Add reorder handler to certification handler**

In `internal/api/handler/certification.go`, add route in `RegisterRoutes`. **IMPORTANT:** Register `/certification/reorder` BEFORE `/certification/{id}` to prevent mux from treating "reorder" as an ID:

```go
func (h *CertificationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/certification", middleware.WithJWTAuth(h.handleViewCertification, h.userStore)).Methods("GET")
	router.HandleFunc("/certification/reorder", middleware.WithJWTAuth(h.handleReorderCertifications, h.userStore)).Methods("POST")
	router.HandleFunc("/certification/{id}", middleware.WithJWTAuth(h.handleViewCertificationById, h.userStore)).Methods("GET")
	router.HandleFunc("/certification", middleware.WithJWTAuth(h.handleCreateCertification, h.userStore)).Methods("POST")
	router.HandleFunc("/certification/{id}", middleware.WithJWTAuth(h.handleEditCertification, h.userStore)).Methods("PUT")
	router.HandleFunc("/certification/{id}", middleware.WithJWTAuth(h.handleDeleteCertification, h.userStore)).Methods("DELETE")
}
```

Add the handler method:

```go
func (h *CertificationHandler) handleReorderCertifications(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Items []certification.ReorderItem `json:"items"`
	}
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if len(payload.Items) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("items list cannot be empty"))
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())

	if err := h.certificationStore.ReorderCertifications(userID, payload.Items); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Certifications reordered successfully"})
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/domain/certification/ internal/api/handler/certification.go
git commit -m "feat: add displayOrder and reorder endpoint for certifications"
```

---

## Task 4: Education Domain — Model, Repository, Handler

**Files:**
- Modify: `internal/domain/education/model.go`
- Modify: `internal/domain/education/repository.go`
- Modify: `internal/api/handler/education.go`

- [ ] **Step 1: Add DisplayOrder to Education struct**

In `internal/domain/education/model.go`, add field and ReorderItem type:

```go
type Education struct {
	ID           int     `json:"id"`
	UserID       int     `json:"userId"`
	School       string  `json:"school"`
	Description  string  `json:"description"`
	Degree       string  `json:"degree"`
	FieldOfStudy string  `json:"fieldOfStudy"`
	StartDate    string  `json:"startDate"`
	EndDate      *string `json:"endDate"`
	IsPresent    bool    `json:"isPresent"`
	DisplayOrder int     `json:"displayOrder"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type ReorderItem struct {
	ID           int `json:"id"`
	DisplayOrder int `json:"displayOrder"`
}
```

- [ ] **Step 2: Update education repository SELECT queries and scan function**

In `internal/domain/education/repository.go`:

**Update `GetEducationById` query:**

```go
row := s.db.QueryRow("SELECT id, userId, school, description, degree, fieldOfStudy, startDate, endDate, isPresent, displayOrder, createdAt, updatedAt FROM education WHERE id = ?", id)
```

**Update `GetEducations` query:**

```go
rows, err := s.db.Query(
	"SELECT id, userId, school, description, degree, fieldOfStudy, startDate, endDate, isPresent, displayOrder, createdAt, updatedAt FROM education WHERE userId = ?",
	userID,
)
```

**Update `scanRowIntoEducation`** — add `displayOrder`:

```go
func scanRowIntoEducation(scanner interface{ Scan(dest ...interface{}) error }) (Education, error) {
	var education Education
	err := scanner.Scan(
		&education.ID,
		&education.UserID,
		&education.School,
		&education.Description,
		&education.Degree,
		&education.FieldOfStudy,
		&education.StartDate,
		&education.EndDate,
		&education.IsPresent,
		&education.DisplayOrder,
		&education.CreatedAt,
		&education.UpdatedAt,
	)
	if err != nil {
		return Education{}, err
	}
	return education, nil
}
```

**Add `ReorderEducations` method** at the end:

```go
func (s *Repository) ReorderEducations(userID int, items []ReorderItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE education SET displayOrder = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ? AND userId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(item.DisplayOrder, item.ID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
```

- [ ] **Step 3: Add reorder handler to education handler**

In `internal/api/handler/education.go`, add route in `RegisterRoutes`. **IMPORTANT:** Register `/education/reorder` BEFORE `/education/{id}` to prevent mux from treating "reorder" as an ID:

```go
func (h *EducationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/education", middleware.WithJWTAuth(h.handleViewEducation, h.userStore)).Methods("GET")
	router.HandleFunc("/education/reorder", middleware.WithJWTAuth(h.handleReorderEducation, h.userStore)).Methods("POST")
	router.HandleFunc("/education", middleware.WithJWTAuth(h.handleCreateEducation, h.userStore)).Methods("POST")
	router.HandleFunc("/education/{id}", middleware.WithJWTAuth(h.handleEditEducation, h.userStore)).Methods("PUT")
	router.HandleFunc("/education/{id}", middleware.WithJWTAuth(h.handleDeleteEducation, h.userStore)).Methods("DELETE")
}
```

Add the handler method:

```go
func (h *EducationHandler) handleReorderEducation(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Items []education.ReorderItem `json:"items"`
	}
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if len(payload.Items) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("items list cannot be empty"))
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())

	if err := h.educationStore.ReorderEducations(userID, payload.Items); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Education reordered successfully"})
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/domain/education/ internal/api/handler/education.go
git commit -m "feat: add displayOrder and reorder endpoint for education"
```

---

## Task 5: Project Domain — Wire displayOrder

**Files:**
- Modify: `internal/domain/project/model.go`
- Modify: `internal/domain/project/repository.go`
- Modify: `internal/api/handler/project.go`

**Note:** The `displayOrder` column already exists in the projects SQL table. We just need to wire it into Go.

- [ ] **Step 1: Add DisplayOrder to Project struct and ReorderItem type**

In `internal/domain/project/model.go`:

```go
type Project struct {
	ID           int     `json:"id"`
	UserID       int     `json:"userId"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Link         string  `json:"link"`
	GithubLink   string  `json:"githubLink"`
	Status       string  `json:"status"`
	IsDraft      bool    `json:"isDraft"`
	DisplayOrder int     `json:"displayOrder"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    *string `json:"updatedAt"`
}

type ReorderItem struct {
	ID           int `json:"id"`
	DisplayOrder int `json:"displayOrder"`
}
```

- [ ] **Step 2: Update project repository queries and scan function**

In `internal/domain/project/repository.go`:

**Update `GetProjectById` query:**

```go
row := s.db.QueryRow(`
	SELECT id, title, description, link, githubLink, status, isDraft, displayOrder, createdAt, updatedAt
	FROM projects
	WHERE id = ?
`, id)
```

**Update `GetProjects` query:**

```go
rows, err := s.db.Query(`
	SELECT id, title, description, link, githubLink, status, isDraft, displayOrder, createdAt, updatedAt
	FROM projects
	WHERE userId = ?
`, userId)
```

**Update `scanProject`** — add `displayOrder`:

```go
func scanProject(scanner interface {
	Scan(dest ...interface{}) error
}) (Project, error) {
	var project Project

	err := scanner.Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&project.Link,
		&project.GithubLink,
		&project.Status,
		&project.IsDraft,
		&project.DisplayOrder,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	return project, err
}
```

**Add `ReorderProjects` method** at the end:

```go
func (s *Repository) ReorderProjects(userID int, items []ReorderItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE projects SET displayOrder = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ? AND userId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(item.DisplayOrder, item.ID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
```

- [ ] **Step 3: Add reorder handler to project handler**

In `internal/api/handler/project.go`, add route in `RegisterRoutes`. **IMPORTANT:** Register `/project/reorder` BEFORE `/project/{id}` to prevent mux from treating "reorder" as an ID:

```go
func (h *ProjectHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/project", middleware.WithJWTAuth(h.handleViewProjects, h.userStore)).Methods("GET")
	router.HandleFunc("/project/reorder", middleware.WithJWTAuth(h.handleReorderProjects, h.userStore)).Methods("POST")
	router.HandleFunc("/project/{id}", middleware.WithJWTAuth(h.handleViewProject, h.userStore)).Methods("GET")
	router.HandleFunc("/project", middleware.WithJWTAuth(h.handleCreateProject, h.userStore)).Methods("POST")
	router.HandleFunc("/project/{id}", middleware.WithJWTAuth(h.handleUpdateProject, h.userStore)).Methods("PUT")
	router.HandleFunc("/project/{id}", middleware.WithJWTAuth(h.handleDeleteProject, h.userStore)).Methods("DELETE")
}
```

Add the handler method:

```go
func (h *ProjectHandler) handleReorderProjects(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Items []project.ReorderItem `json:"items"`
	}
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if len(payload.Items) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("items list cannot be empty"))
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())

	if err := h.projectStore.ReorderProjects(userID, payload.Items); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Projects reordered successfully"})
}
```

- [ ] **Step 4: Verify Go compiles**

```bash
cd /home/alex/my-go/megome && go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add internal/domain/project/ internal/api/handler/project.go
git commit -m "feat: wire displayOrder into projects and add reorder endpoint"
```

---

## Task 6: Frontend — Types and API Client

**Files:**
- Modify: `megome-front/types/domain.ts`
- Create: `megome-front/lib/api/client/reorder.ts`

- [ ] **Step 1: Add displayOrder to domain types**

In `megome-front/types/domain.ts`, add `displayOrder` to each type:

```typescript
export type Experience = {
  id: number;
  userId: number;
  title: string;
  company: string;
  logo: string | null;
  startDate: string;
  endDate: string;
  isPresent: boolean;
  description: string;
  technologies: Technology[];
  displayOrder: number;
  createdAt: string;
  updatedAt: string;
}

export type Education = {
  id: number;
  userId: number;
  school: string;
  description: string;
  degree: string;
  fieldOfStudy: string;
  startDate: string;
  endDate: string;
  isPresent: boolean;
  displayOrder: number;
  createdAt: string;
  updatedAt: string;
}

export type Project = {
  id: number;
  userId: number;
  title: string;
  description: string;
  link: string;
  githubLink: string;
  status: string;
  displayOrder: number;
  createdAt: string;
  updatedAt: string;
}

export type ProjectFull = {
  id: number;
  userId: number;
  title: string;
  description: string;
  link: string;
  githubLink: string;
  status: string;
  isDraft: boolean;
  displayOrder: number;
  createdAt: string;
  updatedAt: string;
  images: {
    cover?: string | null
    screenshots: string[]
  };
  technologies: Technology[]
}

export type Certificate = {
  id: number;
  userId: number;
  title: string;
  issuer: string;
  issueDate: string;
  expirationDate: string | null;
  credentialId: string | null;
  credentialUrl: string | null;
  certificateImage: string | null;
  displayOrder: number;
  createdAt: string;
  updatedAt: string;
}
```

- [ ] **Step 2: Create reorder API client**

Create `megome-front/lib/api/client/reorder.ts`:

```typescript
import { handleResponse } from "@/utils/api/handleResponse";
import { fetchClient } from "./fetchClient";

interface ReorderResponse {
  message: string;
}

interface ReorderItem {
  id: number;
  displayOrder: number;
}

export const reorderClient = async (
  resource: "experience" | "certification" | "education" | "project",
  items: ReorderItem[]
): Promise<ReorderResponse> => {
  const res = await fetchClient(
    `/api/${resource}/reorder`,
    {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ items }),
    }
  );

  return handleResponse<ReorderResponse>(res);
};
```

- [ ] **Step 3: Commit**

```bash
git add megome-front/types/domain.ts megome-front/lib/api/client/reorder.ts
git commit -m "feat: add displayOrder to frontend types and reorder API client"
```

---

## Task 7: Frontend — Install @dnd-kit

**Files:**
- Modify: `megome-front/package.json`

- [ ] **Step 1: Install @dnd-kit packages**

```bash
cd /home/alex/my-go/megome-front && npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities
```

- [ ] **Step 2: Verify install**

```bash
cd /home/alex/my-go/megome-front && npm ls @dnd-kit/core @dnd-kit/sortable
```

Expected: both packages listed without errors.

- [ ] **Step 3: Commit**

```bash
git add megome-front/package.json megome-front/package-lock.json
git commit -m "chore: install @dnd-kit/core and @dnd-kit/sortable"
```

---

## Task 8: Frontend — ReorderPage and Components

**Files:**
- Create: `megome-front/app/(app)/reorder/page.tsx`
- Create: `megome-front/features/reorder/components/ReorderPage.tsx`
- Create: `megome-front/features/reorder/components/ReorderTab.tsx`
- Create: `megome-front/features/reorder/components/SortableItem.tsx`
- Create: `megome-front/features/reorder/index.ts`

- [ ] **Step 1: Create the page route**

Create `megome-front/app/(app)/reorder/page.tsx`:

```tsx
import { ReorderPage } from "@/features/reorder";

export default function ReorderPageRoute() {
  return <ReorderPage />;
}
```

- [ ] **Step 2: Create the barrel export**

Create `megome-front/features/reorder/index.ts`:

```tsx
export { default as ReorderPage } from "./components/ReorderPage";
```

- [ ] **Step 3: Create SortableItem component**

Create `megome-front/features/reorder/components/SortableItem.tsx`:

```tsx
"use client";

import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { Bars3Icon } from "@heroicons/react/24/outline";

interface SortableItemProps {
  id: number;
  title: string;
  subtitle?: string;
}

export default function SortableItem({ id, title, subtitle }: SortableItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    zIndex: isDragging ? 50 : undefined,
    opacity: isDragging ? 0.8 : undefined,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`flex items-center gap-3 rounded-xl border border-base-300 bg-base-100 px-4 py-3 transition-shadow ${
        isDragging ? "shadow-lg" : "hover:shadow-sm"
      }`}
    >
      <button
        {...attributes}
        {...listeners}
        className="cursor-grab touch-none text-base-content/40 hover:text-base-content/70 active:cursor-grabbing"
      >
        <Bars3Icon className="size-5" />
      </button>

      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{title || "Untitled"}</p>
        {subtitle && (
          <p className="truncate text-xs text-base-content/60">{subtitle}</p>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Create ReorderTab component**

Create `megome-front/features/reorder/components/ReorderTab.tsx`:

```tsx
"use client";

import { useState } from "react";
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from "@dnd-kit/core";
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { reorderClient } from "@/lib/api/client/reorder";

import SortableItem from "./SortableItem";

interface ReorderTabProps {
  resource: "experience" | "certification" | "education" | "project";
  items: Array<{ id: number; title: string; subtitle?: string }>;
  onReordered?: () => void;
}

export default function ReorderTab({ resource, items: initialItems, onReordered }: ReorderTabProps) {
  const [items, setItems] = useState(initialItems);
  const [saving, setSaving] = useState(false);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;

    const oldIndex = items.findIndex((item) => item.id === active.id);
    const newIndex = items.findIndex((item) => item.id === over.id);

    const newItems = arrayMove(items, oldIndex, newIndex);
    setItems(newItems);

    const reorderPayload = newItems.map((item, index) => ({
      id: item.id,
      displayOrder: index + 1,
    }));

    setSaving(true);
    try {
      await reorderClient(resource, reorderPayload);
      onReordered?.();
    } catch (error) {
      console.error("Failed to save order:", error);
      setItems(items);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-3">
      {saving && (
        <div className="fixed top-4 right-4 z-50 rounded-lg bg-base-200 px-3 py-1.5 text-xs text-base-content/70">
          Saving...
        </div>
      )}

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleDragEnd}
      >
        <SortableContext
          items={items.map((item) => item.id)}
          strategy={verticalListSortingStrategy}
        >
          <div className="space-y-2">
            {items.map((item) => (
              <SortableItem
                key={item.id}
                id={item.id}
                title={item.title}
                subtitle={item.subtitle}
              />
            ))}
          </div>
        </SortableContext>
      </DndContext>

      {items.length === 0 && (
        <div className="py-12 text-center text-sm text-base-content/50">
          No items to reorder
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 5: Create ReorderPage component**

Create `megome-front/features/reorder/components/ReorderPage.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";

import { Experience, Education, Certificate, Project } from "@/types/domain";
import { getExperienceClient } from "@/lib/api/client/experience";
import { getEducationClient } from "@/lib/api/client/education";
import { getCertificateClient } from "@/lib/api/client/certificate";
import { getProjectsClient } from "@/lib/api/client/project";

import { Card } from "@/components/ui/Card";
import ReorderTab from "./ReorderTab";

type Tab = "experience" | "certificates" | "education" | "projects";

const tabs: { key: Tab; label: string }[] = [
  { key: "experience", label: "Experience" },
  { key: "certificates", label: "Certificates" },
  { key: "education", label: "Education" },
  { key: "projects", label: "Projects" },
];

export default function ReorderPage() {
  const [activeTab, setActiveTab] = useState<Tab>("experience");
  const [loading, setLoading] = useState(true);

  const [experiences, setExperiences] = useState<Experience[]>([]);
  const [certificates, setCertificates] = useState<Certificate[]>([]);
  const [educations, setEducations] = useState<Education[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);

  useEffect(() => {
    const fetchAll = async () => {
      try {
        const [expRes, certRes, eduRes, projRes] = await Promise.all([
          getExperienceClient(),
          getCertificateClient(),
          getEducationClient(),
          getProjectsClient(),
        ]);
        setExperiences(expRes.experiences ?? []);
        setCertificates(certRes.certificates ?? []);
        setEducations(eduRes.educations ?? []);
        setProjects(projRes.projects ?? []);
      } catch (error) {
        console.error("Error fetching items:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchAll();
  }, []);

  const sortByIds = <T extends { id: number; displayOrder: number }>(items: T[]): T[] =>
    [...items].sort((a, b) => a.displayOrder - b.displayOrder || a.id - b.id);

  const experienceItems = sortByIds(experiences).map((e) => ({
    id: e.id,
    title: e.title,
    subtitle: e.company,
  }));

  const certificateItems = sortByIds(certificates).map((c) => ({
    id: c.id,
    title: c.title,
    subtitle: c.issuer,
  }));

  const educationItems = sortByIds(educations).map((e) => ({
    id: e.id,
    title: e.school,
    subtitle: [e.degree, e.fieldOfStudy].filter(Boolean).join(" — "),
  }));

  const projectItems = sortByIds(projects).map((p) => ({
    id: p.id,
    title: p.title,
    subtitle: p.status,
  }));

  const getItems = () => {
    switch (activeTab) {
      case "experience":
        return { items: experienceItems, resource: "experience" as const };
      case "certificates":
        return { items: certificateItems, resource: "certification" as const };
      case "education":
        return { items: educationItems, resource: "education" as const };
      case "projects":
        return { items: projectItems, resource: "project" as const };
    }
  };

  const { items, resource } = getItems();

  return (
    <main className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Reorder Items</h1>
        <p className="text-sm text-base-content/60">
          Drag and drop to reorder how items appear on your public profile.
        </p>
      </div>

      {/* Tabs */}
      <div
        role="tablist"
        aria-label="Reorder sections"
        className="flex gap-2 overflow-x-auto rounded-xl border border-base-300 bg-base-100 p-3"
      >
        {tabs.map((tab) => (
          <button
            key={tab.key}
            role="tab"
            aria-selected={activeTab === tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`whitespace-nowrap rounded-lg px-4 py-2 text-sm font-medium transition-colors cursor-pointer ${
              activeTab === tab.key
                ? "bg-primary text-primary-content"
                : "text-base-content/70 hover:bg-base-200"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Content */}
      <Card className="shadow-xs p-6">
        {loading ? (
          <div className="space-y-3">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="skeleton h-12 w-full rounded-xl" />
            ))}
          </div>
        ) : (
          <ReorderTab
            key={activeTab}
            resource={resource}
            items={items}
          />
        )}
      </Card>
    </main>
  );
}
```

- [ ] **Step 6: Commit**

```bash
git add megome-front/app/\(app\)/reorder/ megome-front/features/reorder/
git commit -m "feat: add /reorder page with drag-and-drop sorting"
```

---

## Task 9: Frontend — Add Reorder to Sidebar Navigation

**Files:**
- Modify: `megome-front/components/ui/Sidebar.tsx`

- [ ] **Step 1: Add Reorder link to sidebar menu**

In `megome-front/components/ui/Sidebar.tsx`, import `ArrowsRightLeftIcon` and add to the menu array:

At the top, add the import:

```tsx
import {
  HomeIcon,
  WindowIcon,
  CodeBracketIcon,
  BookOpenIcon,
  KeyIcon,
  InformationCircleIcon,
  ArrowsRightLeftIcon,
} from "@heroicons/react/24/outline";
```

In the `menu` array, add a new entry after "Projects":

```tsx
{
  name: "Reorder",
  path: "/reorder",
  icon: ArrowsRightLeftIcon,
},
```

- [ ] **Step 2: Verify frontend builds**

```bash
cd /home/alex/my-go/megome-front && npm run build
```

- [ ] **Step 3: Commit**

```bash
git add megome-front/components/ui/Sidebar.tsx
git commit -m "feat: add Reorder link to sidebar navigation"
```

---

## Task 10: End-to-End Verification

- [ ] **Step 1: Verify Go builds**

```bash
cd /home/alex/my-go/megome && go build ./...
```

- [ ] **Step 2: Verify frontend builds**

```bash
cd /home/alex/my-go/megome-front && npm run build
```

- [ ] **Step 3: Run frontend lint**

```bash
cd /home/alex/my-go/megome-front && npm run lint
```

- [ ] **Step 4: Verify migration files exist and are correct**

```bash
ls -la /home/alex/my-go/megome/cmd/migrate/migrations/20260728*
```

Expected: 6 files (3 up + 3 down).

- [ ] **Step 5: Final commit (if any fixes needed)**

```bash
git add -A && git commit -m "fix: address build/lint issues for reorder feature"
```
