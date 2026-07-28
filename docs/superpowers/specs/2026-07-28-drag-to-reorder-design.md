# Drag-to-Reorder Feature Design

## Overview

Add a dedicated `/reorder` page where users can drag-and-drop to reorder items across 4 portfolio sections: Experience, Certificates, Education, and Projects. Uses `@dnd-kit` for drag-and-drop, auto-saves on drop, and persists order via a new `displayOrder` column on each table.

## Database Changes

### New migrations (3 tables)

```sql
ALTER TABLE experiences ADD COLUMN displayOrder INT UNSIGNED NOT NULL DEFAULT 0;
ALTER TABLE certifications ADD COLUMN displayOrder INT UNSIGNED NOT NULL DEFAULT 0;
ALTER TABLE education ADD COLUMN displayOrder INT UNSIGNED NOT NULL DEFAULT 0;
```

### Projects table

The `displayOrder` column already exists in the `projects` table schema but is dead code. Wire it into the Go model, repository, and handlers.

### Default behavior

All existing items get `displayOrder = 0`. After first reorder, values become sequential (1, 2, 3...). `ORDER BY` clauses use `displayOrder ASC, id ASC` so items with the same value maintain insertion order.

## Backend API

### New endpoint

`PATCH /api/v1/{resource}/reorder` for each of: `experiences`, `certifications`, `education`, `projects`

**Request:**
```json
{
  "items": [
    { "id": 1, "displayOrder": 1 },
    { "id": 3, "displayOrder": 2 },
    { "id": 2, "displayOrder": 3 }
  ]
}
```

**Response:** `200 OK`

### Implementation

- Each domain gets a `Reorder(ctx, userID int, items []ReorderItem) error` repository method
- Runs in a transaction: updates `displayOrder` for each item where `id` and `userId` match
- Handler validates: all IDs belong to authenticated user, no duplicate IDs
- All existing `GetAll` queries updated to include `ORDER BY displayOrder ASC, id ASC`

### Go model changes

- Add `DisplayOrder int` field to `Experience`, `Certification`, `Education` structs
- Wire `displayOrder` into `Project` struct (currently missing)
- Include `displayOrder` in JSON responses for all 4 resources

## Frontend

### New page: `/reorder`

**Route:** `/reorder` (add to sidebar navigation)

**Layout:**
- DaisyUI tabbed interface: Experience | Certificates | Education | Projects
- Each tab renders a compact sortable list with item title + subtitle
- Drag handle (grip icon) on each item's left side

### Components

| Component | Purpose |
|---|---|
| `ReorderPage` | Page component. Fetches all 4 resource lists. Renders tabbed interface. |
| `ReorderTab` | Generic sortable list. Accepts items + resource type. Uses `@dnd-kit` `SortableContext`. |
| `SortableItem` | Individual draggable row with grip handle, title, subtitle. |

### Behavior

- On drop → immediately calls `PATCH /api/v1/{resource}/reorder` with new order array
- Optimistic UI update (list reorders instantly)
- Toast notification on success/error
- Loading state while initial data fetches

### Dependencies

- `@dnd-kit/core` and `@dnd-kit/sortable` (npm install)

### API client

New `reorder(resource: string, items: {id: number, displayOrder: number}[])` function in the API client layer.

## Files to modify

### Backend (megome/)
- `internal/domain/experience/model.go` — add DisplayOrder field
- `internal/domain/experience/repository.go` — add Reorder method, update GetAll query
- `internal/domain/experience/handler.go` — add ReorderHandler
- `internal/domain/certification/model.go` — add DisplayOrder field
- `internal/domain/certification/repository.go` — add Reorder method, update GetAll query
- `internal/domain/certification/handler.go` — add ReorderHandler
- `internal/domain/education/model.go` — add DisplayOrder field
- `internal/domain/education/repository.go` — add Reorder method, update GetAll query
- `internal/domain/education/handler.go` — add ReorderHandler
- `internal/domain/project/model.go` — add DisplayOrder field
- `internal/domain/project/repository.go` — add Reorder method, update GetAll query
- `internal/domain/project/handler.go` — add ReorderHandler
- Router registration for new PATCH endpoints
- 3 new migration files (experiences, certifications, education)

### Frontend (megome-front/)
- New page: `app/(app)/reorder/page.tsx`
- New components: `features/reorder/components/ReorderPage.tsx`, `ReorderTab.tsx`, `SortableItem.tsx`
- New API client: `lib/api/client/reorder.ts`
- Update sidebar navigation to include /reorder link
