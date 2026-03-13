**OMNIBASE**

The Ultimate Open-Source Backend-as-a-Service Platform

**Complete Product Roadmap & Architecture Blueprint**

Version 1.0 · March 2026

| MissionBuild the most complete, open-source, free BaaS ever made — combining the best features of Supabase, Firebase, Appwrite, Convex, and PocketBase into a single unified platform. | Promise100% open source. 100% self-hostable. 100% free forever. No paywalls on core features. No vendor lock-in. Your data stays yours. |
| --- | --- |

# 1\. Executive Summary

OmniBase is a next-generation, open-source Backend-as-a-Service (BaaS) platform designed to be the last backend infrastructure a developer will ever need. It eliminates the need to stitch together Supabase for the database, Firebase for push notifications, Appwrite for file storage and hosting, Convex for reactive queries, and PocketBase for lightweight deployments. OmniBase brings all of these into a single, cohesive, production-grade system.

## 1.1 The Problem OmniBase Solves

Today's developers face a fragmented backend landscape. Every BaaS platform excels at something but is missing critical features that force teams to either accept limitations or bolt on third-party services:

*   Supabase is powerful for Postgres + RLS but lacks web hosting, push notifications, crash reporting, and multi-runtime serverless functions.
*   Firebase has excellent mobile SDKs and crash analytics but is closed-source, vendor-locked into Google, and uses a NoSQL model that scales poorly for complex queries.
*   Appwrite offers the widest feature set but lacks vector search, database branching, and scale-to-zero Postgres.
*   Convex has brilliant reactive queries but no file storage, no GraphQL, and limited self-hosting options.
*   PocketBase is a single binary wonder but not production-grade for large teams.

_OmniBase's answer: Take the absolute best from every platform and engineer them into one. No compromises. No paywalls. No lock-in._

## 1.2 Key Differentiators

| Differentiator | What It Means |
| --- | --- |
| Postgres + Vector + Graph | One DB engine handles relational, AI-semantic, and graph queries |
| Reactive by Default | Clients auto-update when data changes — no polling, no manual cache busting |
| Database Branching | Fork your entire DB per pull request — test migrations safely, merge like code |
| Single Binary Mode | Deploy the entire stack as one ~60MB binary for edge, IoT, or solo dev use |
| Built-in Push + Crash | No Firebase dependency — native cross-platform push notifications and crash analytics |
| 10+ Function Runtimes | Write serverless functions in Node, Python, Go, Rust, Ruby, PHP, Java, Kotlin, Swift, Dart |
| AI-Native | pgvector + hybrid BM25/semantic search + AI query builder built directly into the DB layer |
| Scale to Zero | Compute and DB scale down to zero on no traffic — true zero-cost at idle |

# 2\. System Architecture

OmniBase is built around a layered architecture where each layer is independently deployable, replaceable, and testable. The guiding principle is that no layer should know more than it needs to about adjacent layers.

## 2.1 Architecture Overview

_Core Principle: Every component of OmniBase communicates through well-defined internal APIs. This means any layer can be swapped out (e.g., replace Postgres with CockroachDB, replace the auth engine with a custom one) without touching the rest of the system._

### Layer 1 — Client SDKs

The outermost layer. OmniBase ships first-party SDKs for JavaScript/TypeScript, Dart/Flutter, Swift, Kotlin, Python, and a REST/GraphQL API for all other languages. SDKs handle: real-time subscriptions via WebSocket, local-first caching with optimistic updates, offline queue (retry-on-reconnect), auth token refresh, and typed query builders.

### Layer 2 — API Gateway

The API Gateway is the single entry point for all external traffic. It handles: TLS termination, request routing, rate limiting, global API key validation, WebSocket upgrade, and request logging. Built in Go using the Fiber framework for maximum throughput (500k+ req/sec on commodity hardware).

### Layer 3 — Core Services

The business logic layer. Each service is a standalone Go microservice with its own internal API:

*   Auth Service — user management, JWT issuance, OAuth2/OIDC, MFA, session management
*   Database Service — Postgres connection pooling (PgBouncer), query execution, RLS enforcement, schema migration runner
*   Storage Service — file uploads/downloads, CDN routing, image transformation pipeline
*   Functions Service — multi-runtime serverless function executor (WASM-based isolation)
*   Realtime Service — WebSocket hub, change-data-capture from Postgres WAL, push to subscribed clients
*   Messaging Service — push notification routing (APNs, FCM-compatible, Web Push), in-app messaging, email, SMS via adapters
*   Analytics Service — crash reporting aggregation, performance monitoring, funnel tracking
*   Hosting Service — static site deployment, CDN edge routing, custom domain management

### Layer 4 — Data Layer

PostgreSQL 17+ is the single source of truth. Extensions loaded by default:

*   pgvector — vector embeddings and cosine/L2 similarity search
*   pg\_duckdb — analytical queries 600x faster than standard Postgres
*   pg\_cron — native scheduled jobs inside Postgres
*   pg\_graphql — auto-generates GraphQL schema from Postgres schema
*   pgaudit — full audit trail of all DB operations
*   TimescaleDB — time-series data and continuous aggregates
*   age (Apache Graph Extension) — property graph queries with Cypher language

### Layer 5 — Infrastructure Adapters

The bottom layer. OmniBase ships infrastructure adapters that can target: bare metal (single binary), Docker Compose (development), Kubernetes (production), and cloud-native (AWS/GCP/Azure managed services). Adapters are swap-in replacements — the upper layers never know if they are talking to a local SQLite instance or a 50-node Kubernetes cluster.

## 2.2 Technology Stack

| Layer | Technology | Reason |
| --- | --- | --- |
| API Gateway | Go + Fiber | Fastest HTTP router for Go; handles 500k+ req/sec; great WebSocket support |
| Core Services | Go (microservices) | Static binaries, tiny memory footprint, excellent concurrency with goroutines |
| Primary Database | PostgreSQL 17+ | Battle-tested, extensible, supports RLS, JSONB, full-text, vectors |
| Analytical Queries | DuckDB (via pg_duckdb) | Columnar engine, vectorized execution, 600x faster for analytics |
| Graph Queries | Apache AGE | Cypher query language on top of Postgres; no separate graph DB needed |
| Function Runtimes | WebAssembly (Wasmtime) | Secure sandboxing for user functions across 10+ languages |
| Realtime Engine | Go + WAL replication | Postgres WAL → change events → WebSocket push, sub-50ms latency |
| File Storage | MinIO (S3-compatible) | Self-hosted object storage; drop-in S3 replacement; scales to petabytes |
| Cache / Sessions | Valkey (Redis fork, OSS) | Open-source Redis alternative; session store, pub/sub, distributed lock |
| Message Queue | NATS JetStream | Lightweight, embeddable; handles 10M+ messages/sec; built-in persistence |
| Single Binary Mode | SQLite (via GORM) | Embedded DB for zero-infra deployments; full OmniBase features at ~60MB |
| Admin Dashboard | SvelteKit + TypeScript | Fastest frontend framework; SSR; small bundle; excellent TypeScript DX |
| CLI Toolchain | Go (Cobra) | Cross-platform CLI for project management, migrations, deployments |
| Observability | OpenTelemetry + Grafana | Traces, metrics, logs in one stack; connects to any OTel-compatible sink |

# 3\. Complete Feature Specification

This section defines every feature OmniBase will ship, where it was inspired from, and what makes the OmniBase implementation superior.

## 3.1 Database Engine

### 3.1.1 Core Postgres with Auto-REST and Auto-GraphQL

Inspired by Supabase's PostgREST and Nhost's Hasura integration. OmniBase auto-generates a fully typed REST API and GraphQL API from your database schema the moment you create a table. Zero code required. Both APIs update instantly when the schema changes.

*   REST API — GET/POST/PATCH/DELETE per table, filtering operators (eq, gt, lt, like, in, is), relationship traversal, full-text search, pagination with cursor and offset modes
*   GraphQL API — queries, mutations, subscriptions, nested relationship resolution, dataloader for N+1 prevention, directives for auth, pagination, and ordering
*   Type generation — client SDK types are automatically re-generated on schema change

### 3.1.2 Row Level Security (RLS)

Inspired by Supabase. Data access policies are enforced at the database engine level, not in application code. This means even a misconfigured API endpoint cannot expose unauthorized data because Postgres itself blocks the query.

*   Policy editor in Admin UI with policy testing sandbox
*   AI policy generator — describe access rules in plain English, it writes the SQL policy
*   Role-based policies (anon, authenticated, owner, admin, custom roles)
*   Column-level security — hide specific columns from specific roles

### 3.1.3 Vector Search and AI-Native Queries

Inspired by Supabase's pgvector integration but significantly extended. OmniBase treats vector search as a first-class citizen, not a bolted-on extension.

*   pgvector — store 1536-dim OpenAI embeddings, 3072-dim embeddings, custom dimensions
*   Hybrid search — combine BM25 full-text scoring with cosine vector similarity in one query
*   Automatic embedding generation — define an embedding column and OmniBase auto-calls your embedding model on insert/update
*   Semantic search UI in Admin Dashboard — search your data in plain English
*   Multi-modal support — image, audio, text embeddings stored and queried uniformly

### 3.1.4 Database Branching

Inspired by Supabase and Neon. Every database can be branched like a Git repository. Branches are copy-on-write — they are near-instant and cost almost no storage until they diverge.

*   Branch from any point in time using WAL log replay
*   GitHub/GitLab CI integration — auto-create a branch per pull request
*   Branch merge — apply schema migrations from a branch back to main
*   Branch diff viewer — see exactly what changed between branches
*   Auto-cleanup — branches are garbage-collected after a configurable TTL

### 3.1.5 Scale to Zero Postgres

Inspired by Neon. OmniBase implements a compute/storage separation model where the Postgres compute layer can scale to zero instances with sub-1-second cold start. Storage is always persistent on S3-compatible object storage.

## 3.2 Authentication

### 3.2.1 Identity Providers

*   Email + password (bcrypt, Argon2 selectable)
*   Magic link (passwordless email)
*   OTP via SMS (Twilio, AWS SNS adapters)
*   OAuth2 / OIDC — Google, GitHub, Apple, Discord, Twitter/X, LinkedIn, Slack, Microsoft, Spotify, any OIDC-compatible provider
*   SAML 2.0 — enterprise SSO (Okta, Azure AD, Google Workspace)
*   Web3 wallet auth — MetaMask, WalletConnect signature verification
*   Passkey / WebAuthn — hardware key and biometric auth

### 3.2.2 Session Management

*   JWT with configurable expiry + rotating refresh tokens
*   Session revocation — invalidate a specific device or all sessions
*   Concurrent session limits — maximum N active sessions per user
*   Device fingerprinting — flag suspicious new-device logins
*   MFA — TOTP (Authenticator apps), SMS OTP, backup codes

### 3.2.3 User Management Dashboard

*   Search, filter, ban, impersonate, delete users from Admin UI
*   Custom user metadata fields
*   User activity log — last login, IP, device, geo-location
*   Bulk operations — export CSV, bulk ban, bulk email

## 3.3 Serverless Functions

Inspired by Appwrite's multi-runtime functions and Convex's TypeScript-first functions. OmniBase ships the most flexible serverless function system of any BaaS.

### 3.3.1 Supported Runtimes

| Runtime | Version | Notes |
| --- | --- | --- |
| Node.js | 20 LTS + 22 | Default runtime; npm packages supported |
| Python | 3.11 + 3.12 | pip packages; great for ML/AI functions |
| Go | 1.22+ | Compiled to WASM; fastest cold start (~2ms) |
| Rust | stable | Compiled to WASM; memory-safe ultra-performance |
| PHP | 8.3 | Legacy compatibility; Composer packages |
| Ruby | 3.3 | Gems supported |
| Java | 21 LTS | Maven + Gradle builds |
| Kotlin | 2.0 | JVM-based; Coroutines support |
| Swift | 5.10 | Native Apple ecosystem integration |
| Dart | 3.x | Flutter backend functions |

### 3.3.2 Function Features

*   Trigger types: HTTP, cron schedule, database event (insert/update/delete), storage event, queue message, auth event
*   Environment variables — per-function, per-project, and global scopes
*   Secret management — encrypted secrets, runtime injection, no secret in logs
*   Function versions — deploy new version, instant rollback to any previous version
*   Edge deployment — functions deployed to 30+ global PoPs for sub-5ms latency
*   ACID transactions — functions can wrap DB operations in transactions with full rollback on error
*   Built-in DB client — every function gets a pre-authenticated OmniBase client (no credentials needed in function code)

## 3.4 Realtime Engine

Inspired by Convex's reactive-by-default model and Supabase's realtime service. OmniBase's realtime engine is built on Postgres WAL (Write-Ahead Log) streaming, not polling.

*   Table subscriptions — subscribe to inserts, updates, deletes on any table or row
*   Filtered subscriptions — subscribe to changes matching a WHERE clause
*   Presence — track which users are online in a channel (like Figma's collaborative cursors)
*   Broadcast — send ephemeral messages to all clients in a channel (game state, collaborative editing ops)
*   Reactive queries (Convex-style) — define a query function; all clients automatically re-run it when underlying data changes
*   Sub-50ms latency — from DB write to client notification
*   99.99% delivery guarantee — missed events during disconnection are replayed on reconnect

## 3.5 File Storage

### 3.5.1 Core Storage Features

*   S3-compatible API — any S3 SDK works out of the box
*   Bucket-level policies — public, private, authenticated-only, per-user
*   File-level RLS — same Row Level Security model applied to files
*   Resumable uploads — large file uploads survive network interruptions
*   Chunked uploads — parallel chunk uploading for maximum speed

### 3.5.2 Media Transformation Pipeline

Inspired by Cloudinary/Imgix but built-in and free. Transform images and videos on-the-fly via URL parameters.

*   Image: resize, crop, format conversion (WebP, AVIF, HEIC), quality, rotate, blur, watermark
*   Video: thumbnail extraction, transcoding, HLS adaptive streaming
*   CDN integration — transformed assets cached at edge, served from nearest PoP

## 3.6 Push Notifications and Messaging

This is the biggest gap in all existing BaaS platforms. OmniBase ships a complete cross-platform notification system inspired by Firebase Cloud Messaging but fully self-hosted and open source.

### 3.6.1 Push Notifications

*   APNs (Apple Push Notification service) — iOS, macOS, watchOS, tvOS
*   FCM-compatible — Android native; also supports Firebase project migration
*   Web Push (VAPID) — browser push notifications without an app
*   Notification topics — subscribe devices to topics, send to topic = broadcast to all subscribers
*   Segmented sends — target by user attributes, location, last-seen date, custom tags
*   Rich notifications — images, action buttons, deep links, custom sounds
*   Delivery tracking — sent, delivered, opened, dismissed, conversion events

### 3.6.2 In-App Messaging

*   Real-time in-app notification inbox (like Slack's notification bell)
*   Channels — subscribe users to named channels, broadcast to channels
*   Read receipts and delivery status
*   Message templates — define reusable templates with variable substitution

### 3.6.3 Email and SMS

*   Transactional email via SMTP (self-hosted), SendGrid, Mailgun, Resend, AWS SES adapters
*   Email templates with React Email / MJML support
*   SMS via Twilio, AWS SNS, Vonage adapters
*   Unified messaging API — same SDK call, platform is configured per adapter

## 3.7 Crash Reporting and Analytics

Inspired by Firebase Crashlytics and Firebase Analytics. OmniBase ships first-party observability tools so developers never need to add a third-party analytics SDK.

### 3.7.1 Crash Reporting

*   Automatic crash capture — integrate SDK, zero config, crashes are captured and uploaded
*   Symbolication — deobfuscate and desymbolicate stack traces using uploaded dSYM / ProGuard maps
*   Issue grouping — stack-trace fingerprinting clusters similar crashes into single issues
*   Breadcrumbs — log the sequence of events leading up to the crash
*   User impact — see how many users hit each crash, prioritize by reach
*   Custom error logging — log.exception() from any code path

### 3.7.2 Performance Monitoring

*   Network request tracing — latency, payload size, HTTP status per endpoint
*   Screen rendering performance — FPS, slow frame detection for mobile
*   Custom traces — wrap any code block to measure its performance
*   Cold start monitoring — app launch time breakdown

### 3.7.3 Product Analytics

*   Event tracking — track any user action with custom properties
*   Funnel analysis — define multi-step funnels, see where users drop off
*   Retention cohorts — D1/D7/D30 retention, cohort comparison
*   A/B testing — define experiments, assign users to variants, measure impact
*   Session replay — optional, privacy-preserving user session recording

## 3.8 Remote Configuration

Inspired by Firebase Remote Config. Change app behavior without shipping a new app version.

*   Key-value config parameters with typed values (string, number, boolean, JSON)
*   Audience targeting — different values for different user segments
*   Rollout percentages — gradually roll out a feature to 1% → 10% → 100% of users
*   Fetch with TTL — configurable cache duration, instant force-fetch available
*   Emergency kill switch — disable a broken feature instantly without a release

## 3.9 Web Hosting

Inspired by Appwrite's hosting and Netlify/Vercel. OmniBase lets you deploy your frontend from the same platform as your backend. No separate hosting service needed.

*   Static site hosting — upload a build folder, get a URL in seconds
*   Git-push deployment — connect a GitHub/GitLab repo, every push triggers a build and deploy
*   Build pipeline — built-in support for Vite, Next.js, SvelteKit, Astro, Nuxt, Remix, CRA
*   Preview deployments — every branch/PR gets its own preview URL
*   Custom domains — add any domain with automatic TLS certificate via Let's Encrypt
*   Edge CDN — static assets served from 30+ global PoPs
*   Redirect and rewrite rules
*   Environment variables per deployment target (preview vs production)

## 3.10 Single Binary Mode

Inspired by PocketBase. OmniBase compiles to a single self-contained binary (~60MB) that runs the entire stack with zero dependencies. SQLite replaces Postgres in this mode. Perfect for: local development without Docker, IoT/edge devices, Raspberry Pi deployments, desktop apps with embedded backend, solo developer side projects.

*   omnibase serve — starts the entire platform on one port
*   All features available (storage uses local filesystem, messaging uses local queue)
*   Admin dashboard embedded in binary
*   One-command migration to full Postgres mode when you outgrow SQLite
*   ~60MB binary — smaller than most Electron apps

# 4\. Development Roadmap

The roadmap is divided into 6 phases. Each phase builds on the previous and results in a shippable, usable product. Phase 1 alone is already more usable than many existing BaaS platforms.

## 4.1 Phase Overview

| Phase | Timeline | Status | Priority | Goal |
| --- | --- | --- | --- | --- |
| Phase 1 Foundation | Months 1–3 | Start Here | Critical | Working Postgres BaaS with auth, auto-API, realtime, storage, and admin UI |
| Phase 2 Functions | Months 4–6 | High Priority | Critical | Multi-runtime serverless functions, cron jobs, DB event triggers |
| Phase 3 Mobile | Months 7–9 | High Priority | High | Push notifications, crash reporting, remote config, mobile SDKs |
| Phase 4 AI Native | Months 10–12 | Planned | High | Vector search, hybrid search, AI query builder, embedding auto-generation |
| Phase 5 Hosting | Months 13–15 | Planned | Medium | Web hosting, git-deploy, preview URLs, CDN edge, analytics |
| Phase 6 Enterprise | Months 16–18 | Future | Medium | DB branching, scale-to-zero, SAML SSO, audit logs, compliance (HIPAA/GDPR) |

## 4.2 Phase 1 — Foundation (Months 1–3)

_Goal: A developer can spin up OmniBase, create a Postgres database, get auto-generated REST + GraphQL APIs, configure RLS, add auth, upload files, and watch realtime data changes. All from a single Docker Compose command._

### Month 1 — Core Infrastructure

**Week 1–2: Project Setup**

1.  Initialize Go monorepo with workspace mode (go.work)
2.  Set up GitHub repository with branch protection, CI/CD via GitHub Actions
3.  Define internal service communication contracts (Protocol Buffers)
4.  Set up local development environment: Docker Compose with Postgres 17, Valkey, NATS
5.  Initialize SvelteKit admin dashboard project
6.  Set up E2E test framework (Playwright) and unit test conventions

**Week 3–4: API Gateway**

1.  Build Go + Fiber API gateway with: routing, middleware chain, TLS, rate limiting
2.  Implement API key authentication at gateway level
3.  WebSocket upgrade handler for realtime connections
4.  Request logging to structured JSON (OpenTelemetry compatible)
5.  Health check endpoints (/health, /ready) for Kubernetes probes

### Month 2 — Auth + Database Services

**Auth Service**

1.  User registration with email + password (Argon2 hashing)
2.  JWT issuance (access token 15min, refresh token 30d with rotation)
3.  Email verification flow with configurable SMTP adapter
4.  OAuth2 social login: Google, GitHub (extensible provider interface)
5.  Session management with Valkey session store
6.  MFA with TOTP (compatible with Google Authenticator)

**Database Service**

1.  PgBouncer connection pooler (transaction mode for REST, session mode for functions)
2.  PostgREST integration for auto-REST API generation
3.  pg\_graphql integration for auto-GraphQL API generation
4.  RLS enforcement — all queries run as the authenticated user role
5.  Schema migration runner with version tracking (compatible with Flyway/Liquibase format)
6.  Database introspection API — returns table/column/relationship metadata

### Month 3 — Storage + Realtime + Admin UI

**Storage Service**

1.  MinIO integration with OmniBase bucket management API
2.  Signed URL generation for private file access
3.  File-level RLS policies
4.  Image transformation pipeline using libvips (resize, crop, format, quality)
5.  Resumable upload protocol (TUS)

**Realtime Service**

1.  Postgres WAL logical replication consumer
2.  WebSocket connection manager with authentication
3.  Table-level change subscriptions (INSERT, UPDATE, DELETE events)
4.  Row-level filtered subscriptions with RLS enforcement
5.  Presence tracking — connected users per channel
6.  Broadcast messaging — ephemeral pub/sub between clients

**Admin Dashboard (Phase 1 scope)**

1.  Project creation and API key management
2.  Database table browser — view, filter, edit rows
3.  SQL editor with syntax highlighting and query history
4.  Schema editor — create tables, add columns, define relationships via UI
5.  RLS policy editor with testing sandbox
6.  Auth user management — list, search, ban, delete users
7.  Storage bucket browser — upload, preview, delete files
8.  API logs viewer — real-time request log with filtering

## 4.3 Phase 2 — Serverless Functions (Months 4–6)

_Goal: Developers can write, deploy, and trigger serverless functions in their language of choice. Functions have built-in DB access, secrets, versioning, and run in secure WASM sandboxes._

### Month 4 — Function Runtime Infrastructure

1.  Wasmtime WASM executor integration in Go — compile user code to WASM, execute in isolated sandbox
2.  Node.js runtime: Deno-based WASM compilation, npm package bundling (esbuild)
3.  Python runtime: Pyodide-based WASM, pip dependency bundling
4.  Function deployment API: upload code → build → deploy pipeline
5.  Per-function environment variables and encrypted secrets vault
6.  Built-in OmniBase client injection — every function gets a pre-authenticated client

### Month 5 — Trigger System and More Runtimes

1.  HTTP trigger — function exposed at /functions/v1/{name}
2.  Cron trigger — pg\_cron-backed scheduling with cron expression syntax
3.  Database event trigger — hooks on INSERT/UPDATE/DELETE per table
4.  Storage event trigger — hooks on file upload/delete/rename
5.  Auth event trigger — hooks on user signup, login, password reset
6.  Go runtime (compiled to WASM via TinyGo)
7.  Rust runtime (compiled to WASM via wasm-pack)
8.  PHP runtime (via php-wasm)

### Month 6 — Functions Dashboard + Job Queues

1.  Function logs viewer — real-time streaming logs with structured filtering
2.  Function metrics — invocations, duration, error rate, p95 latency
3.  Version management — deploy, view history, rollback to any version
4.  NATS JetStream job queue integration — functions as queue consumers
5.  Dead letter queue — failed jobs after N retries go to DLQ for inspection
6.  Edge deployment — package functions for deployment to edge PoPs
7.  ACID transaction support in functions — DB operations wrapped in transactions

## 4.4 Phase 3 — Mobile and Messaging (Months 7–9)

_Goal: Mobile developers can replace Firebase entirely — push notifications, crash reporting, performance monitoring, and remote config all built into OmniBase._

### Month 7 — Push Notifications

1.  APNs integration — iOS/macOS push via HTTP/2 APNs protocol
2.  Android push — FCM HTTP v1 API integration
3.  Web Push — VAPID key management, ServiceWorker compatible
4.  Device registration API — SDKs register device tokens automatically
5.  Topic subscriptions — subscribe devices to named topics
6.  Notification scheduling — send at a specific time or after a delay
7.  Delivery tracking — webhook callbacks for delivered/opened/failed

### Month 8 — Crash Reporting

1.  JavaScript SDK — automatic unhandledrejection and window.onerror capture
2.  React Native SDK — native crash handler for iOS + Android
3.  Flutter SDK — platform-channel crash integration
4.  dSYM / ProGuard upload API for symbolication
5.  Issue grouping algorithm — stack trace fingerprinting
6.  Breadcrumb capture — network requests, user actions, log statements
7.  Crash dashboard — issue list, trend chart, affected users count, stacktrace viewer

### Month 9 — Remote Config + In-App Messaging

1.  Remote config parameter store with typed values
2.  Audience segmentation engine — target by user property, platform, app version
3.  Rollout percentage control with consistent user assignment (hash-based)
4.  SDK fetch with background refresh and configurable TTL
5.  In-app notification inbox API — store, retrieve, mark-read
6.  Transactional email service with SMTP adapter
7.  SMS adapter framework (Twilio first)
8.  React Native, Flutter, iOS, Android mobile SDKs reaching feature parity

## 4.5 Phase 4 — AI-Native Features (Months 10–12)

_Goal: OmniBase becomes the best backend for AI-powered applications — vector search, hybrid search, automatic embedding generation, AI-assisted query building, and agent workflow support._

### Month 10 — Vector Search Foundation

1.  pgvector extension configuration — HNSW and IVFFlat indexes, configurable dimensions
2.  Embedding column type in schema editor — declare a column as embedding, pick dimensions
3.  Embedding adapter framework — OpenAI, Cohere, Mistral, local Ollama, custom HTTP adapter
4.  Automatic embedding triggers — on INSERT/UPDATE, auto-call embedding model and store vector
5.  Vector similarity search API — cosine, L2, inner product similarity with top-K and threshold

### Month 11 — Hybrid Search and AI Query Builder

1.  BM25 full-text search index on text columns (via pg\_bm25)
2.  Hybrid search endpoint — Reciprocal Rank Fusion (RRF) combining BM25 score + vector similarity
3.  Re-ranking pipeline — optional cross-encoder re-ranker for higher-quality results
4.  AI query builder in Admin Dashboard — type a question in English, get a SQL query
5.  AI table filter — search data with natural language in the table browser
6.  AI policy generator — describe RLS rules in English, get Postgres policies

### Month 12 — AI Agent Workflow Support

1.  Agent state table — persistent key-value store for agent working memory
2.  Agent workflow API — multi-step function chains with state passing between steps
3.  Tool call result caching — memoize expensive tool calls with configurable TTL
4.  Long-running function support — async functions that run for up to 15 minutes
5.  Streaming responses — server-sent events from functions for LLM streaming
6.  Semantic caching — cache LLM responses by semantic similarity of the input query

## 4.6 Phase 5 — Web Hosting and Analytics (Months 13–15)

_Goal: Developers can deploy their full-stack application — frontend hosting, CDN, git-push deploys, preview URLs, and product analytics — all from one OmniBase project._

### Month 13 — Static Hosting and CDN

1.  Static file hosting service — accept build output, serve via CDN
2.  Custom domain management — DNS verification, automatic TLS via Let's Encrypt
3.  CDN edge node integration — push assets to Cloudflare R2 / Fastly / self-hosted Varnish
4.  Redirect and rewrite rule engine (Netlify \_redirects format compatible)
5.  Hosting dashboard — deployments list, domain management, bandwidth stats

### Month 14 — Git-Push Deployment Pipeline

1.  GitHub App and GitLab integration — webhook on push/PR
2.  Build pipeline runner — detect framework (Vite, Next.js, SvelteKit, Astro), run build command
3.  Preview deployment per branch/PR — unique URL for every open PR
4.  Production deployment on merge to main
5.  Build logs streaming — watch the build log live in dashboard
6.  Rollback — one-click revert to any previous deployment

### Month 15 — Product Analytics

1.  Client-side analytics SDK — auto-track page views, sessions, clicks
2.  Custom event tracking API — track('button\_click', { label: 'signup' })
3.  Funnel builder — define multi-step funnels, visualize conversion rates
4.  Retention cohort analysis — D1/D7/D30 user retention charts
5.  A/B testing engine — create experiments, assign users, measure conversion
6.  Analytics dashboard — DAU/MAU charts, top events, funnel overview

## 4.7 Phase 6 — Enterprise and Scale (Months 16–18)

_Goal: OmniBase is ready for large enterprises — database branching per PR, scale-to-zero, SAML SSO, compliance certifications, audit logs, and multi-region deployments._

### Month 16 — Database Branching

1.  WAL-based copy-on-write branching — instant branch creation, minimal storage overhead
2.  Branch management API and dashboard UI
3.  GitHub Actions integration — auto create/delete DB branch per PR lifecycle
4.  Branch schema diff — visualize schema changes between branch and main
5.  Branch merge — apply migrations from branch to parent
6.  TTL-based auto-cleanup of stale branches

### Month 17 — Scale to Zero and Performance

1.  Compute/storage separation — Postgres WAL streamed to S3-compatible storage
2.  Cold start optimization — target sub-500ms from zero to accepting queries
3.  Auto-scaling compute — scale horizontally based on connection count and CPU
4.  pg\_duckdb integration — columnar analytical queries 600x faster
5.  TimescaleDB integration — time-series data, continuous aggregates, data retention policies
6.  Apache AGE integration — graph queries with Cypher language on Postgres data

### Month 18 — Enterprise Security and Compliance

1.  SAML 2.0 SSO — Okta, Azure AD, Google Workspace, any SAML IdP
2.  SOC 2 Type II compliance documentation and controls
3.  HIPAA-ready mode — encryption at rest, audit logs, BAA template
4.  GDPR tools — right to erasure API, data export, consent management
5.  pgaudit full audit trail — every DB operation logged with user, timestamp, query
6.  IP allowlisting — restrict API access by CIDR range
7.  Multi-region deployment — active-active or active-passive across geographic regions
8.  Disaster recovery runbooks — automated backup validation, RTO/RPO SLA documentation

# 5\. SDK and Developer Experience Strategy

The SDK is the most visible part of OmniBase to developers. A confusing or bug-prone SDK will kill adoption regardless of how powerful the backend is. OmniBase invests heavily in DX from day one.

## 5.1 TypeScript/JavaScript SDK (Primary)

The JS/TS SDK is the reference implementation. All other SDKs must reach feature parity with it.

*   Fully typed — TypeScript generics infer table types from your schema, giving autocomplete on column names, filter operators, and relationships
*   Tree-shakeable — import only what you use; zero unused code in your bundle
*   Works everywhere — browser, Node.js, Deno, Bun, Edge Runtime, React Native
*   Reactive hooks — React, Vue, Svelte hooks that automatically re-render when subscribed data changes
*   Optimistic updates — UI updates instantly on mutation; reverts automatically if server rejects
*   Offline queue — mutations queued locally when offline; replayed in order when reconnected

## 5.2 Type Generation CLI

The omnibase gen command connects to a running project and generates TypeScript types for every table, view, function, and storage bucket. This means:

*   type User = Database\['public'\]\['Tables'\]\['users'\]\['Row'\] — automatically generated, always in sync
*   Query builder provides full autocomplete: from('users').select('email, name').eq('id', userId)
*   RLS violations surface as TypeScript compile errors when possible

## 5.3 CLI Tool (omnibase)

| Command | Description |
| --- | --- |
| omnibase init | Scaffold a new project with chosen framework template |
| omnibase dev | Start local OmniBase instance (single binary mode) |
| omnibase gen types | Generate TypeScript types from live schema |
| omnibase db migrate | Run pending migrations against target environment |
| omnibase db branch | Create a DB branch from current state |
| omnibase functions deploy | Deploy function(s) to cloud or self-hosted |
| omnibase functions logs | Stream function logs to terminal |
| omnibase hosting deploy | Deploy frontend build to hosting |
| omnibase secrets set | Set encrypted environment variable for a function |
| omnibase inspect | Inspect running project health and metrics |
| omnibase upgrade | Upgrade OmniBase version with zero-downtime migration |

## 5.4 Framework Integration Guides

OmniBase ships official starter templates and integration guides for every major framework on day one of Phase 1:

*   Next.js — App Router, Server Actions, Middleware auth, SSR data fetching patterns
*   SvelteKit — load functions, server-side auth, form actions
*   Nuxt 3 — composables, server routes, pinia store integration
*   Remix — loader/action patterns, session management
*   React Native + Expo — push notification setup, offline sync, auth flow
*   Flutter — Dart SDK, push notifications, offline-first patterns
*   Swift / SwiftUI — iOS SDK, push notifications, iCloud keychain integration

# 6\. Deployment Strategy

## 6.1 Self-Hosting Options

### Option A — Single Binary (Simplest)

_omnibase serve --port 8080 — That is the entire deployment command. One binary. Zero dependencies. Full OmniBase._

*   SQLite as database — persistent, reliable for up to ~10k users
*   Local filesystem storage
*   Admin UI accessible at http://localhost:8080
*   Ideal for: solo projects, prototypes, internal tools, IoT, desktop apps with embedded backend

### Option B — Docker Compose (Recommended for Teams)

*   docker compose up — spins up: OmniBase API, Postgres 17, MinIO, Valkey, NATS, Grafana
*   Persistent volumes for all data
*   Single .env file for all configuration
*   Ideal for: small-to-medium teams, VPS deployment, staging environments

### Option C — Kubernetes (Production)

*   Official Helm chart with sensible defaults
*   Each service as a separate Deployment with HPA for auto-scaling
*   PodDisruptionBudgets for zero-downtime upgrades
*   Prometheus + Grafana dashboards included in Helm chart
*   Compatible with EKS, GKE, AKS, k3s, k0s

### Option D — OmniBase Cloud (Hosted)

*   Managed cloud offering for teams who don't want to manage infrastructure
*   100% feature-parity with self-hosted
*   Free tier: 2 projects, 500MB DB, 1GB storage, 100k function invocations/month
*   Pricing: usage-based, significantly cheaper than Supabase/Firebase for equivalent usage

## 6.2 Migration Paths

OmniBase provides migration tooling for teams moving from existing platforms:

*   From Supabase — direct Postgres dump import; auth user migration with password preservation; storage bucket migration; function code compatible
*   From Firebase Firestore — export JSON transformer to Postgres tables with equivalent schema
*   From PocketBase — SQLite to Postgres migration script with schema translation
*   From Appwrite — API-compatible shim layer for Appwrite SDK calls during migration period

# 7\. Open Source and Community Strategy

## 7.1 License Strategy

OmniBase uses a tiered licensing model designed to be maximally permissive for self-hosters while protecting the project's sustainability:

*   Core platform — Apache 2.0 license. Freely use, modify, distribute, including commercially. No restrictions.
*   Admin Dashboard — MIT license. Fully open, modify the UI as you wish.
*   Client SDKs — MIT license. No restrictions on commercial use.
*   OmniBase Cloud — proprietary service built on the open-source core. Revenue funds development.

## 7.2 Repository Structure

| Repository | Contents |
| --- | --- |
| omnibase/omnibase | Main monorepo — all Go services, single binary, Helm chart |
| omnibase/dashboard | SvelteKit admin dashboard |
| omnibase/sdk-js | JavaScript/TypeScript SDK |
| omnibase/sdk-dart | Dart/Flutter SDK |
| omnibase/sdk-swift | Swift SDK for iOS/macOS |
| omnibase/sdk-kotlin | Kotlin SDK for Android |
| omnibase/sdk-python | Python SDK |
| omnibase/cli | omnibase CLI tool |
| omnibase/examples | Example apps for every major framework |
| omnibase/docs | Documentation site (Docusaurus) |

## 7.3 Contribution Guidelines

*   All contributions go through pull requests — no direct pushes to main
*   Every PR requires: passing CI (unit + integration + E2E tests), at least one reviewer approval, changelog entry
*   Issue labels: good first issue (no complex context needed), help wanted (complex but external-friendly), core team only (architecture-level decisions)
*   Monthly community calls — roadmap updates, contributor showcases, Q&A
*   Discord server — #help, #contributors, #showcase, #random channels
*   Hacktoberfest participation every October — curated good first issues

## 7.4 Documentation Philosophy

*   Every feature ships with docs before it ships to production — no undocumented features
*   Docs include: concept explanation, quickstart, API reference, common patterns, gotchas
*   Interactive examples — runnable code snippets directly in the docs
*   Comparison pages — 'Coming from Supabase', 'Coming from Firebase', 'Coming from Appwrite'
*   Video walkthroughs for complex features (DB branching, RLS setup, function deployment)

# 8\. Team Structure and Resource Requirements

## 8.1 Minimum Viable Team (Phase 1–2)

| Role | Count | Responsibilities |
| --- | --- | --- |
| Lead Backend Engineer | 1 | Go services architecture, API gateway, database service, auth service |
| Backend Engineer | 1 | Storage service, realtime engine, functions runtime |
| Frontend Engineer | 1 | SvelteKit admin dashboard, SDK React hooks |
| DevOps / Infra Engineer | 1 (part-time) | Docker Compose, Helm chart, CI/CD, observability stack |
| SDK Engineer | 1 (part-time) | TypeScript SDK, type generation, CLI tool |
| Technical Writer | 1 (part-time) | Docs site, API reference, tutorials, changelog |

## 8.2 Infrastructure Costs (Self-Hosted Development)

For development and initial cloud offering, the estimated monthly infrastructure cost is minimal thanks to the open-source tools selected:

*   Development servers (3x VPS): ~$60/month
*   CI/CD (GitHub Actions): free tier sufficient for Phase 1–2
*   Domain + TLS: ~$15/year
*   Object storage for build artifacts (Backblaze B2): ~$5/month
*   Total Phase 1 monthly infrastructure cost: under $100

## 8.3 Growth Phase Team (Phase 3–6)

*   Add 2 backend engineers (mobile features, AI integrations)
*   Add 1 Android SDK engineer and 1 iOS SDK engineer
*   Add 1 security engineer (for Phase 6 compliance work)
*   Add 1 DevRel engineer (community management, conference talks, tutorial content)
*   Full team size at Phase 6: 8–10 engineers + 2 DevRel/community

# 9\. Competitive Positioning

## 9.1 Feature Comparison at Full Completion (Phase 6)

| Feature | OmniBase | Supabase | Firebase | Appwrite | Convex | PocketBase |
| --- | --- | --- | --- | --- | --- | --- |
| Auto REST + GraphQL API | ✓ Both | ✓ REST | ✗ | ✓ REST | ✗ | ✓ REST |
| Row Level Security | ✓ | ✓ | ✗ | ✓ | ✗ | ✓ |
| Vector + Hybrid Search | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| Database Branching | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| Reactive Queries | ✓ | Partial | ✓ | ✗ | ✓ | ✗ |
| Multi-Runtime Functions | ✓ 10+ | ✓ TS only | ✓ JS/Python | ✓ 10+ | ✓ TS | ✗ |
| Built-in Push Notifications | ✓ | ✗ | ✓ | ✓ | ✗ | ✗ |
| Crash Reporting | ✓ | ✗ | ✓ | ✗ | ✗ | ✗ |
| Remote Config | ✓ | ✗ | ✓ | ✗ | ✗ | ✗ |
| Web Hosting + CDN | ✓ | ✗ | ✓ | ✓ | ✗ | ✗ |
| Product Analytics | ✓ | ✗ | ✓ | ✗ | ✗ | ✗ |
| Scale to Zero | ✓ | ✓ | ✓ | ✗ | ✓ | ✗ |
| Single Binary Mode | ✓ | ✗ | ✗ | ✗ | ✗ | ✓ |
| 100% Open Source | ✓ | Partial | ✗ | ✓ | ✗ | ✓ |
| SAML SSO | ✓ | ✓ (paid) | ✗ | ✗ | ✗ | ✗ |
| Graph Queries (Cypher) | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ |

_At Phase 6 completion, OmniBase will be the only platform in existence to offer all of these features in a single, open-source, self-hostable product. This is the moat._

# 10\. Success Metrics

## 10.1 Phase-by-Phase Success Criteria

| Phase | Timeline | Success Criteria |
| --- | --- | --- |
| Phase 1 | Month 3 | 100 developers using OmniBase in production; 500 GitHub stars; all core features stable |
| Phase 2 | Month 6 | 500 active projects; 1,000 GitHub stars; functions runtime handling 1M+ invocations/month |
| Phase 3 | Month 9 | First mobile app live with OmniBase push + crash; 2,500 GitHub stars; positive mobile DX feedback |
| Phase 4 | Month 12 | 50+ AI-powered apps using OmniBase vector search; recognized in AI tooling discussions |
| Phase 5 | Month 15 | Full-stack apps deployed start-to-finish on OmniBase; developer tweets showing it as Vercel alternative |
| Phase 6 | Month 18 | 10,000+ GitHub stars; 3+ enterprise customers on self-hosted; recognized as top 3 BaaS platform |

## 10.2 Technical Performance Targets

*   API gateway throughput: 500,000+ requests/second on a single 8-core node
*   Realtime latency: under 50ms from DB write to WebSocket client notification (p99)
*   Function cold start: under 100ms for Node/Python, under 10ms for Go/Rust (WASM)
*   Single binary startup time: under 2 seconds from process start to accepting requests
*   DB branch creation time: under 5 seconds regardless of database size
*   Storage upload speed: saturate available network bandwidth (no artificial throttling)
*   Uptime SLA (cloud): 99.9% monthly uptime commitment from Phase 1 cloud launch

## 10.3 Developer Experience Targets

*   Time from zero to first API call: under 5 minutes with single binary mode
*   Time from zero to production-ready app: under 30 minutes with Docker Compose
*   TypeScript types generation: under 3 seconds (omnibase gen types)
*   Documentation coverage: 100% of public API surface documented before each release
*   Issue response time: first response within 24 hours on GitHub Issues

# 11\. Next Steps — Getting Started Today

_The best time to start was yesterday. The second best time is today. Here is exactly what to do in the first week._

1.  Create the GitHub organization and repositories (omnibase org, main monorepo, dashboard repo)
2.  Set up the Go workspace monorepo with the service directory structure outlined in Section 2.5
3.  Write the API gateway in Go + Fiber — this is the skeleton everything else plugs into
4.  Stand up a Docker Compose environment with Postgres 17 and all dependencies
5.  Implement the database service with PgBouncer + PostgREST integration
6.  Ship the first working demo: create a table in admin UI, insert a row, see it via the auto-generated REST API
7.  Write the Getting Started documentation page alongside the code — not after
8.  Post on Hacker News 'Show HN: OmniBase — Open-source BaaS combining the best of Supabase, Firebase, and Appwrite' when Phase 1 is complete

_OmniBase has the potential to be the Rails of backend infrastructure — opinionated, batteries-included, open-source, and beloved by developers for a decade. Build it with that ambition._

OmniBase Roadmap v1.0 — Open Source · Free Forever · Built for Developers

**github.com/machinelearningprodigy/OmniBase**