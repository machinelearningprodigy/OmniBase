# OmniBase

> **The Ultimate Open-Source Backend-as-a-Service Platform**  
> Combining the best of Supabase, Firebase, Appwrite, Convex, and PocketBase — into one.

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-blue.svg)](LICENSE)
[![Phase](https://img.shields.io/badge/Phase-1%20Foundation-brightgreen)](https://github.com/machinelearningprodigy/OmniBase)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![SvelteKit](https://img.shields.io/badge/Dashboard-SvelteKit-FF3E00?logo=svelte)](https://kit.svelte.dev)

---

## What is OmniBase?

OmniBase is a next-generation, 100% open-source Backend-as-a-Service (BaaS) platform designed to be the **last backend infrastructure a developer will ever need**. It eliminates fragmentation by combining:

| Feature | OmniBase | vs. Others |
|---------|----------|-----------|
| **PostgreSQL + Auto-REST + Auto-GraphQL** | ✅ | Supabase only gives REST |
| **Realtime subscriptions (WAL-based)** | ✅ | Supabase ✅, Firebase ✅ |
| **Built-in Object Storage** | ✅ | Appwrite ✅, Supabase ✅ |
| **Multi-runtime Serverless Functions** | 🔜 Phase 2 | Supabase TS-only |
| **Push Notifications (no Firebase dep)** | 🔜 Phase 3 | Firebase only |
| **Vector Search + AI-native** | 🔜 Phase 4 | Supabase (limited) |
| **Single Binary Mode (like PocketBase)** | 🔜 Phase 6 | PocketBase only |
| **100% Open Source, Apache 2.0** | ✅ | Supabase: mixed |

---

## 🚀 Quick Start (5 minutes)

### Prerequisites
- [Docker](https://docker.com) with Docker Compose v2

### 1. Clone and configure
```bash
git clone https://github.com/machinelearningprodigy/OmniBase
cd omnibase
cp .env.example .env          # Edit with your settings
```

### 2. Start the stack
```bash
docker compose up -d
```

That's it. OmniBase is now running:

| Service | URL |
|---------|-----|
| **Admin Dashboard** | [http://localhost:3001](http://localhost:3001) |
| **API Gateway** | [http://localhost:8000](http://localhost:8000) |  
| **REST API** | [http://localhost:8000/rest/v1/](http://localhost:8000/rest/v1/) |
| **Auth API** | [http://localhost:8000/auth/v1/](http://localhost:8000/auth/v1/) |
| **Storage API** | [http://localhost:8000/storage/v1/](http://localhost:8000/storage/v1/) |
| **MinIO Console** | [http://localhost:9001](http://localhost:9001) |

### 3. Connect with the SDK

```bash
npm install @omnibase/omnibase-js
```

```typescript
import { createClient } from '@omnibase/omnibase-js'

const omni = createClient('http://localhost:8000', 'your-anon-key')

// Auth — works just like Supabase
const { data, error } = await omni.auth.signUp({
  email: 'user@example.com',
  password: 'super-secure-password',
})

// Database — fully typed PostgREST queries
const { data: posts } = await omni
  .from('posts')
  .select('id, title, user:users(email)')
  .eq('published', true)
  .order('created_at', { ascending: false })
  .limit(10)

// Realtime — subscribe to live changes
omni.channel('posts')
  .on('INSERT', (payload) => console.log('New post!', payload.new))
  .subscribe()

// Storage
const { data: file } = await omni.storage
  .from('avatars')
  .upload('user-123.jpg', imageFile)
```

---

## Architecture

OmniBase is built as a **Go microservices monorepo** with a **SvelteKit** admin dashboard:

```
omnibase/
├── services/
│   ├── gateway/      # API Gateway — Go + Fiber (Port 8000)
│   ├── auth/         # Authentication — Go + pgx (Port 9001)  
│   ├── storage/      # File Storage — Go + MinIO (Port 9003)
│   └── realtime/     # WebSocket + WAL consumer (Port 9004)
├── shared/           # Shared Go packages (config, logger, jwt)
├── dashboard/        # Admin UI — SvelteKit (Port 3001)
├── sdk/js/           # JavaScript/TypeScript SDK
└── docker-compose.yml
```

**Infrastructure**: Postgres 17 + PostgREST + MinIO + Valkey (Redis fork) + NATS JetStream

---

## Roadmap

| Phase | Focus | Status |
|-------|-------|--------|
| **Phase 1** | Auth, Database (REST+GraphQL+RLS), Storage, Realtime, Admin UI | 🔨 In Progress |
| **Phase 2** | Serverless Functions (10+ runtimes, WASM isolation) | 📅 Planned |
| **Phase 3** | Push Notifications, Crash Reporting, Remote Config | 📅 Planned |
| **Phase 4** | Vector Search, Hybrid Search, AI Query Builder | 📅 Planned |
| **Phase 5** | Web Hosting, Git-push Deploy, Analytics | 📅 Planned |
| **Phase 6** | DB Branching, Scale-to-Zero, SAML SSO, Enterprise | 📅 Planned |

---

## Contributing

OmniBase is built in the open. All contributions welcome:

1. Fork the repository
2. Create a branch: `git checkout -b feature/my-feature`
3. Make your changes with tests
4. Submit a PR — all PRs require CI passing + 1 reviewer

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## License

- **Core Platform**: Apache 2.0
- **Admin Dashboard**: MIT  
- **Client SDKs**: MIT

Built with ❤️ by the OmniBase contributors · [github.com/machinelearningprodigy/OmniBase](https://github.com/machinelearningprodigy/OmniBase)
