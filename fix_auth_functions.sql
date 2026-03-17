-- Step 1: Create auth helper functions (must exist before any RLS policy uses them)
CREATE OR REPLACE FUNCTION auth.uid() RETURNS UUID AS $$
    SELECT NULLIF(current_setting('request.jwt.claims', TRUE)::jsonb->>'sub', '')::UUID
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION auth.role() RETURNS TEXT AS $$
    SELECT NULLIF(current_setting('request.jwt.claims', TRUE)::jsonb->>'role', '')
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION auth.email() RETURNS TEXT AS $$
    SELECT NULLIF(current_setting('request.jwt.claims', TRUE)::jsonb->>'email', '')
$$ LANGUAGE sql STABLE;

-- Step 2: Grant execute on these functions to all relevant roles
GRANT EXECUTE ON FUNCTION auth.uid() TO authenticated, anon, service_role;
GRANT EXECUTE ON FUNCTION auth.role() TO authenticated, anon, service_role;
GRANT EXECUTE ON FUNCTION auth.email() TO authenticated, anon, service_role;
