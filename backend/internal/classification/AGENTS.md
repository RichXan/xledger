# Classification Package

Categories and tags for transaction organization.

## Files

| File | Purpose |
|------|---------|
| `handler.go` | HTTP handlers for categories/tags |
| `category_service.go` | Category CRUD, tree structure |
| `tag_service.go` | Tag CRUD, deduplication |
| `template_service.go` | Default category templates |
| `repository*.go` | In-memory + Postgres storage |

## Key Rules

- Categories support parent-child hierarchy
- Tags are flat, unique per user (case-insensitive)
- Default categories seeded from `default_categories` table
- `usage_count` tracks category popularity

## Patterns

- Category tree built via `parent_id` FK
- Tag dedup via `lower(name)` unique index
- Templates copied to user on first login

## Don't

- Don't allow circular category references
- Don't delete categories with children
- Don't skip tag normalization (lowercase)
