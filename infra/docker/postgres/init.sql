-- ─────────────────────────────────────────────────────────────────────────────
-- OmniBase Initial Postgres Setup
-- This runs once when the container starts for the first time.
-- ─────────────────────────────────────────────────────────────────────────────

-- Create the extensions we need in Phase 1
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Create OmniBase schemas
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS storage;
CREATE SCHEMA IF NOT EXISTS realtime;
CREATE SCHEMA IF NOT EXISTS extensions;

-- Grant schema usage to the main user
GRANT USAGE ON SCHEMA auth TO omnibase;
GRANT USAGE ON SCHEMA storage TO omnibase;
GRANT USAGE ON SCHEMA realtime TO omnibase;

-- ─────────────────────────────────────────────────────────────────────────────
-- Roles (mirrors Supabase role structure for compatibility)
-- ─────────────────────────────────────────────────────────────────────────────
DO $$
BEGIN
  -- anon: unauthenticated user (very limited access via RLS)
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'anon') THEN
    CREATE ROLE anon NOLOGIN NOINHERIT;
  END IF;

  -- authenticated: logged-in user
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'authenticated') THEN
    CREATE ROLE authenticated NOLOGIN NOINHERIT;
  END IF;

  -- service_role: admin access — bypasses RLS
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'service_role') THEN
    CREATE ROLE service_role NOLOGIN NOINHERIT BYPASSRLS;
  END IF;

  -- authenticator: the role PostgREST uses to connect
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'authenticator') THEN
    CREATE ROLE authenticator NOINHERIT LOGIN PASSWORD 'authenticator_password';
  END IF;

  GRANT anon TO authenticator;
  GRANT authenticated TO authenticator;
  GRANT service_role TO authenticator;
  GRANT omnibase TO authenticator;
END
$$;

-- Give public schema permissions
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO public;

-- ─────────────────────────────────────────────────────────────────────────────
-- Enable WAL Logical Replication for the Realtime Service
-- ─────────────────────────────────────────────────────────────────────────────
-- Create a publication for all tables (Realtime service subscribes to this)
CREATE PUBLICATION omnibase_publication FOR ALL TABLES;

-- ─────────────────────────────────────────────────────────────────────────────
-- Auth Schema
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS auth.users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT UNIQUE,
    phone           TEXT UNIQUE,
    password_hash   TEXT,
    role            TEXT NOT NULL DEFAULT 'authenticated',
    raw_user_meta_data  JSONB DEFAULT '{}'::jsonb,
    raw_app_meta_data   JSONB DEFAULT '{}'::jsonb,
    is_super_admin  BOOLEAN DEFAULT FALSE,
    is_banned       BOOLEAN DEFAULT FALSE,
    email_confirmed_at  TIMESTAMPTZ,
    phone_confirmed_at  TIMESTAMPTZ,
    last_sign_in_at     TIMESTAMPTZ,
    banned_until        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS auth_users_email_idx ON auth.users(email);
CREATE INDEX IF NOT EXISTS auth_users_phone_idx ON auth.users(phone);
CREATE INDEX IF NOT EXISTS auth_users_created_at_idx ON auth.users(created_at DESC);

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION auth.update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER auth_users_updated_at
    BEFORE UPDATE ON auth.users
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at();

-- Sessions table  
CREATE TABLE IF NOT EXISTS auth.sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    refresh_token   TEXT NOT NULL UNIQUE,
    user_agent      TEXT,
    ip              INET,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS auth_sessions_user_id_idx ON auth.sessions(user_id);
CREATE INDEX IF NOT EXISTS auth_sessions_refresh_token_idx ON auth.sessions(refresh_token);

-- OAuth accounts
CREATE TABLE IF NOT EXISTS auth.oauth_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    provider        TEXT NOT NULL,
    provider_id     TEXT NOT NULL,
    access_token    TEXT,
    refresh_token   TEXT,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_id)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Storage Schema
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS storage.buckets (
    id              TEXT PRIMARY KEY,  -- user-defined name
    name            TEXT NOT NULL UNIQUE,
    owner           UUID REFERENCES auth.users(id),
    public          BOOLEAN DEFAULT FALSE,
    allowed_mime_types  TEXT[],
    file_size_limit BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS storage.objects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bucket_id       TEXT NOT NULL REFERENCES storage.buckets(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    owner           UUID REFERENCES auth.users(id),
    content_type    TEXT,
    size            BIGINT DEFAULT 0,
    metadata        JSONB DEFAULT '{}'::jsonb,
    path_tokens     TEXT[] GENERATED ALWAYS AS (string_to_array(name, '/')) STORED,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(bucket_id, name)
);

CREATE INDEX IF NOT EXISTS storage_objects_bucket_id_idx ON storage.objects(bucket_id);
CREATE INDEX IF NOT EXISTS storage_objects_name_idx ON storage.objects(name);
CREATE INDEX IF NOT EXISTS storage_objects_owner_idx ON storage.objects(owner);

-- ─────────────────────────────────────────────────────────────────────────────
-- Row Level Security Setup
-- ─────────────────────────────────────────────────────────────────────────────
-- Enable RLS on public tables (user tables added via schema editor will have RLS enabled)
-- Storage RLS
ALTER TABLE storage.objects ENABLE ROW LEVEL SECURITY;
ALTER TABLE storage.buckets ENABLE ROW LEVEL SECURITY;

-- Default policies (allow public buckets to be read by everyone)
CREATE POLICY "Public buckets are readable" ON storage.objects
    FOR SELECT USING (
        (SELECT public FROM storage.buckets WHERE id = bucket_id) = TRUE
    );

CREATE POLICY "Owners can modify their objects" ON storage.objects
    FOR ALL USING (owner = auth.uid());

-- Function to get current user ID from JWT (used in RLS policies)
CREATE OR REPLACE FUNCTION auth.uid() RETURNS UUID AS $$
    SELECT NULLIF(current_setting('request.jwt.claims', TRUE)::jsonb->>'sub', '')::UUID
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION auth.role() RETURNS TEXT AS $$
    SELECT NULLIF(current_setting('request.jwt.claims', TRUE)::jsonb->>'role', '')
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION auth.email() RETURNS TEXT AS $$
    SELECT NULLIF(current_setting('request.jwt.claims', TRUE)::jsonb->>'email', '')
$$ LANGUAGE sql STABLE;

-- ─────────────────────────────────────────────────────────────────────────────
-- Grant permissions to roles
-- ─────────────────────────────────────────────────────────────────────────────
GRANT SELECT ON auth.users TO authenticated;
GRANT ALL ON storage.buckets TO authenticated;
GRANT ALL ON storage.objects TO authenticated;
GRANT SELECT ON storage.buckets TO anon;
GRANT SELECT ON storage.objects TO anon;

-- Grant usage of all sequences
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO anon, authenticated, service_role;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA auth TO authenticated, service_role;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA storage TO authenticated, service_role;

-- Service role bypasses everything
GRANT ALL ON ALL TABLES IN SCHEMA auth TO service_role;
GRANT ALL ON ALL TABLES IN SCHEMA storage TO service_role;
GRANT ALL ON ALL TABLES IN SCHEMA public TO service_role;

-- ─────────────────────────────────────────────────────────────────────────────
-- PostgREST configuration
-- ─────────────────────────────────────────────────────────────────────────────
-- Comment schemas for PostgREST introspection
COMMENT ON SCHEMA public IS 'OmniBase public data schema — auto-REST and auto-GraphQL APIs are generated from this schema';
COMMENT ON SCHEMA auth IS 'OmniBase authentication schema';
COMMENT ON SCHEMA storage IS 'OmniBase storage schema';
