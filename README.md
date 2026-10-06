# IncidentFlow

Realtime incident management platform with a Go backend, React frontend, PostgreSQL, and Playwright-based QA automation.

## What is in this repository

- `apps/backend` - Go API with auth, profile, middleware, repository and DB migrations
- `apps/frontend` - React application placeholder, not started yet
- `tests` - API, E2E, factories and fixtures
- `docs` - implementation notes and setup documentation

## Stack

- Go
- Gin
- PostgreSQL
- React
- TypeScript
- Playwright
- Docker
- GitHub Actions

## Current status

- Backend auth and JWT are implemented
- Database migrations are working
- Backend tests are passing
- Frontend has not been started yet

## How to run locally

### 1. Start the database and backend

From the repository root:

```powershell
docker compose up -d
```

### 2. Run backend tests

```powershell
cd apps/backend
go test ./tests -v
```

### 3. Stop and reset everything if needed

```powershell
docker compose down -v
```

## Database migrations

The backend uses versioned SQL migrations in `apps/backend/migrations`.

- `001_create_users.sql` creates the base `users` table
- `002_add_user_names.sql` adds `first_name` and `last_name`

This keeps schema changes incremental and safer to evolve.

## Next recommended sprint

1. Finish backend foundation and CI
2. Add incident model and CRUD
3. Build the frontend shell
4. Connect auth flow end to end
