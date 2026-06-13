# Tasks — a focused, keyboard-driven task manager

A full-stack task management application: a **Go + PostgreSQL** REST API and a
**Next.js** frontend. Built to feel like a real product — fast, dense, and
keyboard-first (think Linear/Raycast) — rather than a CRUD demo.

> **Live demo:** _add deployed frontend + API URLs here_
>
> **Demo login:** create an account in seconds, or use the seeded one if provided.

---

## Highlights

- **Real Go backend** — idiomatic `chi` + `pgx`, not API routes. Migrations run
  on startup; the whole thing is one static binary.
- **Secure auth** — signup/login with bcrypt-hashed passwords and a JWT stored
  in an **httpOnly cookie**, so a page refresh keeps you signed in and the token
  is never exposed to JavaScript.
- **Ownership enforced at the query layer** — a user can only ever read or
  mutate their own tasks; this is covered by an integration test.
- **Filter + search + sort that compose** — status filter, title search, and
  sort by due/priority/created all work together, with pagination.
- **Optimistic UI** — create / complete / delete update instantly and roll back
  automatically if the server rejects the change.
- **Real-time** — task changes stream over Server-Sent Events (e.g. across tabs).
- **Activity log** — every task keeps a per-field change history (created,
  status, priority, due date, …) shown as a timeline.
- **⌘K command palette**, **dark/light mode** (persisted), responsive layout,
  and graceful loading / empty / error states throughout.
- **Tests + CI** — Go unit + HTTP integration tests, frontend Vitest suite, all
  wired into GitHub Actions.

---

## Tech stack

| Layer    | Choice                                                                 |
| -------- | ---------------------------------------------------------------------- |
| Frontend | Next.js 16 (App Router), React 19, TypeScript, Tailwind v4             |
| Data/UI  | TanStack Query, react-hook-form + zod, framer-motion, cmdk, next-themes |
| Backend  | Go 1.26, chi (router), pgx (Postgres), goose (migrations)              |
| Auth     | JWT (`golang-jwt`) in an httpOnly cookie, bcrypt password hashing      |
| Database | PostgreSQL 17                                                          |
| Infra    | Docker / docker-compose, GitHub Actions                               |

---

## Quick start

### Option A — one command (Docker)

Requires Docker. Brings up Postgres, the API, and the frontend together:

```bash
docker compose up --build
```

- Frontend → http://localhost:3000
- API → http://localhost:8080

### Option B — run each service locally

**Prerequisites:** Go 1.26+, Node 20+, a running PostgreSQL.

**1. Database** — create a database (or use the compose one):

```bash
createdb tasks
```

**2. Backend**

```bash
cd backend
cp .env.example .env          # then edit DATABASE_URL + JWT_SECRET
go run ./cmd/server           # migrations run automatically on startup
# API listening on :8080
```

**3. Frontend**

```bash
cd frontend
cp .env.example .env.local    # set NEXT_PUBLIC_API_URL=http://localhost:8080
npm install
npm run dev
# App on http://localhost:3000
```

---

## Environment variables

See [`backend/.env.example`](backend/.env.example) and
[`frontend/.env.example`](frontend/.env.example). Summary:

**Backend**

| Variable          | Required | Description                                            |
| ----------------- | -------- | ------------------------------------------------------ |
| `DATABASE_URL`    | yes      | Postgres connection string                             |
| `JWT_SECRET`      | yes      | Token signing secret (**≥ 32 chars**)                  |
| `JWT_TTL_HOURS`   | no       | Session lifetime in hours (default 168 = 7 days)       |
| `FRONTEND_ORIGIN` | no       | Allowed CORS origin (default `http://localhost:3000`)  |
| `PORT`            | no       | API port (default `8080`)                              |
| `APP_ENV`         | no       | `production` sets `Secure` + `SameSite=None` on cookie |

**Frontend**

| Variable              | Required | Description                  |
| --------------------- | -------- | ---------------------------- |
| `NEXT_PUBLIC_API_URL` | yes      | Base URL of the Go API       |

---

## API reference

All task routes require authentication (the auth cookie is sent automatically).
Errors share one envelope: `{ "error": { "code", "message", "fields?" } }`.

### Auth

| Method | Path           | Description                          |
| ------ | -------------- | ------------------------------------ |
| POST   | `/auth/signup` | Create account, returns user + cookie |
| POST   | `/auth/login`  | Log in, returns user + cookie         |
| POST   | `/auth/logout` | Clear the session cookie              |
| GET    | `/auth/me`     | Current user (used to rehydrate)      |

### Tasks

| Method | Path            | Description                                    |
| ------ | --------------- | ---------------------------------------------- |
| POST   | `/tasks`        | Create a task                                  |
| GET    | `/tasks`        | List (filter/search/sort/paginate — see below) |
| GET    | `/tasks/{id}`   | Fetch one                                      |
| GET    | `/tasks/{id}/activity` | Change history for a task               |
| PATCH  | `/tasks/{id}`   | Partial update                                 |
| DELETE | `/tasks/{id}`   | Delete                                         |
| GET    | `/tasks/stream` | Server-Sent Events stream of changes           |
| GET    | `/admin/tasks`  | **Admin only** — list every user's tasks       |

**List query parameters** (all optional, all combine):

| Param      | Values                                  |
| ---------- | --------------------------------------- |
| `status`   | `todo` \| `in_progress` \| `done`       |
| `search`   | case-insensitive match on title         |
| `sortBy`   | `created_at` \| `due_date` \| `priority` |
| `sortDir`  | `asc` \| `desc`                         |
| `page`     | 1-based page number                     |
| `pageSize` | items per page (max 100)                |

Example:

```
GET /tasks?status=todo&search=report&sortBy=due_date&sortDir=asc&page=1&pageSize=10
```

---

## Admin role

The admin bonus lets an admin view all users' tasks (toggle in the UI header).
Roles aren't self-assignable; promote a user directly in the database:

```sql
UPDATE users SET role = 'admin' WHERE email = 'you@example.com';
```

Sign out and back in to refresh the token's role claim.

---

## Tests

**Backend** (unit tests always run; integration tests run when a DB is provided):

```bash
cd backend
go test ./...                                   # unit tests
TEST_DATABASE_URL=postgres://… go test ./...    # + HTTP integration tests
```

- `internal/auth` — password hashing, JWT round-trip, tamper/expiry rejection
- `internal/httpx` — PATCH three-state semantics (absent vs null vs value)
- `internal/server` — full HTTP flow: **ownership isolation** (user B can't touch
  user A's task), **filter + search + sort composing** correctly, and the
  **activity log** recording created/status changes

**Frontend**

```bash
cd frontend
npm test
```

- `toQueryString` composes filter/search/sort/pagination
- `TaskCard` rendering, complete toggle, delete action, overdue handling

CI ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)) runs both suites,
gofmt/vet, lint, and a production build on every push.

---

## Architecture notes

```
backend/
  cmd/server         entrypoint (config, db, graceful shutdown)
  internal/
    auth             bcrypt + JWT issue/verify
    config           env loading with fail-fast validation
    db               pgx pool + embedded goose migrations
    events           in-process pub/sub broker for SSE
    httpx            JSON helpers, error envelope, validation, Optional[T]
    middleware       auth (cookie/bearer) + admin guard
    server           router wiring (testable)
    task / user      models, repositories, HTTP handlers
frontend/
  src/lib            api client, auth context, React Query hooks
  src/components/ui  small design-system kit (button, input, modal, …)
  src/components     header, command palette, task views
  src/app            App Router routes (login, signup, tasks)
```

---

## Assumptions & trade-offs

- **Auth via httpOnly cookie** rather than a token in `localStorage`. It's
  XSS-safe and survives refresh without extra client code. The cost is
  cross-site cookie config: in production the API sets `Secure` + `SameSite=None`
  (enabled by `APP_ENV=production`) and CORS is locked to `FRONTEND_ORIGIN`.
  The middleware also accepts a `Bearer` token, so non-browser clients still work.
- **SSE, not WebSockets.** Task updates are one-directional (server → client), so
  SSE is the simpler, proxy-friendly fit. The broker is **in-memory**, which is
  perfect for a single instance; horizontal scaling would swap it for Redis
  pub/sub or Postgres `LISTEN/NOTIFY` — isolated to `internal/events`.
- **Repository pattern with hand-written SQL** (via pgx) over an ORM — explicit,
  fast, and easy for a reviewer to read. Ownership is enforced in the queries
  themselves (`WHERE user_id = …`) rather than only in handlers.
- **Priority sorting** uses a Postgres enum declared `low < medium < high`, so
  ordering is correct without a `CASE` expression.
- **Optimistic UI** updates the visible page immediately and rolls back on error;
  it then revalidates against the server so canonical ordering/pagination win.
- **Migrations run on startup** for zero-friction setup. For a larger team I'd
  gate this behind an explicit migrate step in the deploy pipeline.
- The admin "all tasks" view shows a short owner id rather than joining user
  details, to keep the admin path a thin addition to the existing list query.

---

## Deployment

- **Frontend** → Vercel. Set `NEXT_PUBLIC_API_URL` to the deployed API URL.
- **Backend** → any container host (Railway / Render / Fly). It's a single Docker
  image; provide `DATABASE_URL`, `JWT_SECRET`, `FRONTEND_ORIGIN`, and
  `APP_ENV=production`.
- **Database** → managed Postgres (e.g. Neon). Migrations apply automatically on
  the first boot.
