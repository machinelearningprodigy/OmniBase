# OmniBase — Feature Status

This document maps what is **built and usable** today so you can self-host and use OmniBase end-to-end like Supabase, Firebase, or Appwrite.

---

## ✅ Implemented (ready to use)

### Core infrastructure
- **API Gateway** — Go + Fiber, routing, rate limiting, CORS, request logging, health (`/health`), readiness (`/ready` with auth check)
- **Docker Compose** — Postgres 17, PostgREST, MinIO, Valkey, NATS, all services wired

### Database
- **Auto REST API** — PostgREST at `/rest/v1/*` (path strip so PostgREST receives `/table`)
- **Auto GraphQL** — Database service `pg_graphql` at `/graphql/v1`
- **Schema & meta** — Create tables, list tables/schemas/functions/policies, run SQL via `/pg/*`
- **Migrations** — Tracked in `omnibase.migrations`; run via SQL Editor or CLI `db migrate`
- **RLS** — Row Level Security in Postgres; dashboard RLS page to enable/disable per table

### Auth
- **Email + password** — Sign up, sign in, sign out, password reset (recover + reset)
- **JWT** — Access + refresh tokens, anon and service_role keys
- **OAuth2** — Google, GitHub (configure via env); authorize + callback
- **Sessions** — Stored in DB; revoke refresh token / all sessions
- **Admin** — List users, get/update/delete/ban, invite, generate magic links

### Storage
- **Buckets** — Create, list, delete (MinIO backend)
- **Objects** — Upload, list, delete; public URL for public buckets
- **Signed URLs** — JWT-based signed URLs for private objects

### Realtime
- **WAL consumer** — Postgres logical replication → change events
- **WebSocket** — Subscribe by schema/table/event (INSERT, UPDATE, DELETE)
- **Presence & broadcast** — Channel presence and ephemeral broadcast messages

### Functions
- **Runtimes** — `static-json`, `webhook`, `javascript` (Node), `python`
- **CRUD** — List, get, create/update, delete functions via API
- **Invoke** — `POST /functions/v1/:slug/invoke` with optional JWT

### Dashboard (SvelteKit)
- **Overview** — Stats, service health, realtime feed, quick links
- **Table Editor** — List tables, create table, browse rows (REST), insert
- **SQL Editor** — Run SQL, optional migration name
- **GraphQL** — Explorer for `/graphql/v1`
- **RLS** — List tables/policies, toggle RLS per table
- **Migrations** — List migration history
- **Auth** — Sign up, sign in, sign out; Users list (admin); Auth Providers (OAuth status)
- **Storage** — Buckets, upload, list objects, signed URL, delete
- **Functions** — List, create, edit, delete, invoke
- **API Logs** — Recent requests (method, path, status, latency)
- **Settings** — Project URL (editable), anon/service keys (when signed in)

### SDK (TypeScript/JavaScript)
- **OmniBaseClient** — `createClient(url, anonKey)`
- **Auth** — signUp, signIn, signOut, getUser, session persistence
- **Database** — `from(table).select().insert().update().delete().order().eq()` (PostgREST-style)
- **Storage** — `from(bucket).upload().getPublicUrl()` etc.
- **Realtime** — `channel(name).on('INSERT', cb).subscribe()`

### CLI
- **omnibase init** — Scaffold `.omnibase/`, `.env.example`
- **omnibase gen types** — Fetch tables from API, write TypeScript types
- **omnibase db migrate** — Run SQL files from `.omnibase/migrations` via API
- **omnibase version** — Print version

---

## 🔜 Roadmap (not yet built)

- **Phase 2+** — WASM functions, cron triggers, DB event triggers, edge deploy
- **Phase 3** — Push notifications (APNs, FCM, Web Push), crash reporting, remote config
- **Phase 4** — Vector search (pgvector), hybrid search, AI query builder
- **Phase 5** — Web hosting, git-push deploy, preview URLs, product analytics
- **Phase 6** — DB branching, scale-to-zero, SAML SSO, audit logs, single binary mode

---

## Parity at a glance

| Feature              | Supabase | Firebase | Appwrite | OmniBase (current) |
|----------------------|----------|----------|----------|--------------------|
| REST API             | ✅       | —        | ✅       | ✅                 |
| GraphQL              | —        | —        | —        | ✅                 |
| Auth (email + OAuth) | ✅       | ✅       | ✅       | ✅                 |
| Storage              | ✅       | ✅       | ✅       | ✅                 |
| Realtime             | ✅       | ✅       | Limited | ✅                 |
| Functions            | TS       | JS/Py    | Multi   | Static/Webhook/JS/Py |
| Dashboard            | ✅       | ✅       | ✅       | ✅                 |
| CLI                  | ✅       | —        | —       | ✅                 |
| Self-host full       | Partial  | No       | ✅       | ✅                 |

You can run OmniBase today for database, auth, storage, realtime, and functions end-to-end, with dashboard and CLI, fully self-hosted.
