# Getting Started with OmniBase

This guide gets you from zero to a fully working, self-hosted backend in one place — similar to Supabase, Firebase, or Appwrite, but 100% open source and run by you.

---

## What you get

| Feature | Like | OmniBase |
|--------|------|----------|
| **Database** | Supabase PostgREST | ✅ Auto REST + GraphQL from Postgres |
| **Auth** | Firebase Auth / Supabase Auth | ✅ Email/password, OAuth (Google, GitHub), JWT |
| **Storage** | Supabase Storage / Appwrite | ✅ Buckets, upload, signed URLs |
| **Realtime** | Supabase Realtime / Firebase | ✅ WAL-based subscriptions, presence |
| **Functions** | Supabase Edge / Firebase Functions | ✅ Static JSON, webhook, JavaScript, Python |
| **Dashboard** | Supabase Dashboard | ✅ Tables, SQL editor, RLS, users, storage, logs |
| **CLI** | Supabase CLI | ✅ `omnibase init`, `gen types`, `db migrate` |

---

## Prerequisites

- **Docker** and **Docker Compose** v2 ([install](https://docs.docker.com/get-docker/))
- (Optional) **Node.js** 18+ for the dashboard dev server or SDK
- (Optional) **Go** 1.24+ to build the CLI

---

## 1. Clone and start the stack

```bash
git clone https://github.com/machinelearningprodigy/OmniBase
cd OmniBase
cp .env.example .env
# Edit .env if needed (defaults work for local dev)
docker compose up -d
```

Wait for all services to be healthy (about 30–60 seconds). Then:

| What | URL |
|------|-----|
| **Dashboard** | http://localhost:3001 |
| **API** | http://localhost:8000 |
| **MinIO console** | http://localhost:9001 |

---

## 2. Create an account and get your API key

1. Open **http://localhost:3001**
2. Click **Sign in** (top right) or go to **http://localhost:3001/auth/login**
3. Click **Create one** to go to **http://localhost:3001/auth/signup**
4. Sign up with email and password (min 8 characters)
5. Sign in with the same credentials
6. Go to **Settings** (or **http://localhost:3001/settings**)
7. Copy your **Project URL** and **anon** key (and **service_role** if you need admin operations)

You’ll use the **anon** key in your app; use **service_role** only on a secure server or in the CLI for migrations.

---

## 3. Create a table and call the API

### From the dashboard

1. Go to **Table Editor** (sidebar)
2. Click **+ New**, name the table (e.g. `posts`)
3. Add columns (e.g. `id` uuid PK, `title` text, `created_at` timestamptz)
4. Click **Create Table**
5. Select the table and use **Insert Row** or the SQL Editor to add data

### From your app (SDK)

```bash
npm install @omnibase/omnibase-js
```

```typescript
import { createClient } from '@omnibase/omnibase-js'

const omni = createClient('http://localhost:8000', 'YOUR_ANON_KEY')

// Insert
const { data, error } = await omni.from('posts').insert({
  title: 'Hello OmniBase',
  created_at: new Date().toISOString(),
})

// Query
const { data: posts } = await omni.from('posts').select('*').order('created_at', { ascending: false })
```

### REST and GraphQL

- **REST:** `GET/POST/PATCH/DELETE` on `http://localhost:8000/rest/v1/<table>`
  - Headers: `Authorization: Bearer YOUR_ANON_KEY`, `Content-Type: application/json`
- **GraphQL:** `POST http://localhost:8000/graphql/v1` with `{ "query": "..." }` and the same `Authorization` header

---

## 4. Auth in your app

```typescript
// Sign up
const { data, error } = await omni.auth.signUp({ email: 'u@example.com', password: 'secure' })

// Sign in
const { data } = await omni.auth.signIn({ email: 'u@example.com', password: 'secure' })
// data.access_token, data.refresh_token, data.user

// Use the session: later requests automatically use the stored token
const { data: me } = await omni.auth.getUser()
```

---

## 5. Storage (files)

1. In the dashboard go to **Storage**, create a bucket (e.g. `avatars`)
2. Upload files from the UI or via SDK:

```typescript
const { data, error } = await omni.storage.from('avatars').upload('me.jpg', file)
// Get public URL (if bucket is public) or signed URL
const { data: url } = await omni.storage.from('avatars').getPublicUrl('me.jpg')
```

---

## 6. Realtime (live data)

```typescript
omni.channel('posts').on('INSERT', (payload) => {
  console.log('New row:', payload)
}).subscribe()
```

---

## 7. Edge functions

1. In the dashboard go to **Edge Functions**
2. Create a function (e.g. slug `hello`, runtime **static-json** or **javascript**)
3. Invoke: `POST http://localhost:8000/functions/v1/hello` with `Authorization: Bearer YOUR_ANON_KEY`

---

## 8. CLI (types and migrations)

Build the CLI (from repo root):

```bash
cd cli
go build -o omnibase .
# Or: go install .   then use `omnibase` from PATH
```

Then:

```bash
# Scaffold project
./omnibase init

# Generate TypeScript types (uses anon or service key)
export OMNIBASE_ANON_KEY="your-anon-key"
./omnibase gen types -o src/omnibase-types.ts

# Run migrations (use service role key)
export OMNIBASE_ANON_KEY="your-service-role-key"
./omnibase db migrate --path .omnibase/migrations
```

Put SQL files in `.omnibase/migrations/` (e.g. `001_initial.sql`). The CLI calls the same `/pg/query` API the SQL Editor uses.

---

## 9. Row Level Security (RLS)

1. Go to **RLS Policies** in the dashboard
2. Enable RLS on a table and add policies (e.g. “Users can only read their own rows”) via the SQL Editor:  
   `CREATE POLICY ... ON your_table FOR SELECT USING (auth.uid() = user_id);`
3. Use the **anon** key in the client; PostgREST applies RLS per request using the JWT.

---

## 10. Comparison at a glance

| Task | Supabase | Firebase | Appwrite | OmniBase (self-hosted) |
|------|----------|----------|----------|------------------------|
| Start backend | Cloud signup | Cloud signup | Docker / Cloud | `docker compose up -d` |
| Database API | REST (PostgREST) | Firestore SDK | REST | REST + GraphQL |
| Auth | Email + OAuth | Email + OAuth | Email + OAuth | Email + OAuth |
| Storage | Yes | Yes | Yes | Yes (MinIO) |
| Realtime | Yes | Yes | Limited | Yes (WAL) |
| Functions | Edge (TS) | Cloud Functions | Many runtimes | Static, webhook, JS, Python |
| Hosting | No | No | Yes | Planned |
| Full self-host | Partial | No | Yes | Yes |

---

## Troubleshooting

- **Dashboard or API not loading**  
  Run `docker compose ps` and `docker compose logs gateway auth`. Ensure nothing is already using ports 3001, 8000, 5432.

- **“Invalid token” or 401**  
  Use the key from **Settings** (anon or service_role). For REST/GraphQL use header:  
  `Authorization: Bearer YOUR_ANON_KEY`

- **Tables not found in REST**  
  Create tables in the **Table Editor**. PostgREST serves the `public` schema by default.

- **Functions 500**  
  For **javascript** / **python** runtimes, Node.js or Python must be installed in the container (or use **static-json** / **webhook** only).

- **CLI “API key required”**  
  Set `OMNIBASE_ANON_KEY` or pass `--key YOUR_KEY`. Use service role key for `db migrate`.

---

## Next steps

- **Docs:** See [README.md](README.md) and [OmniBase_Roadmap.md](OmniBase_Roadmap.md)
- **API reference:** Use the dashboard **SQL Editor**, **GraphQL** page, and **API Logs** to inspect requests
- **Production:** Use a strong `JWT_SECRET` and proper Postgres/Redis/MinIO backups; put the stack behind TLS (e.g. Caddy/Traefik)

You now have a single, self-hosted backend that covers database, auth, storage, realtime, and functions end to end — similar to Supabase, Firebase, and Appwrite, all in your own environment.
