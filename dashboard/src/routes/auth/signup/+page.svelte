<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/stores'
  import { getOmniBaseUrl } from '$lib/api'

  let email = $state('')
  let password = $state('')
  let confirmPassword = $state('')
  let error = $state<string | null>(null)
  let loading = $state(false)

  async function handleSubmit(e: Event) {
    e.preventDefault()
    error = null
    if (!email.trim() || !password) {
      error = 'Email and password are required'
      return
    }
    if (password.length < 8) {
      error = 'Password must be at least 8 characters'
      return
    }
    if (password !== confirmPassword) {
      error = 'Passwords do not match'
      return
    }
    loading = true
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/signup`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: email.trim(), password }),
      })
      const data = await resp.json()
      if (resp.ok) {
        // Store session if tokens returned (e.g. in development mode)
        if (data.access_token) {
          localStorage.setItem('omnibase.session', JSON.stringify({
            access_token: data.access_token,
            refresh_token: data.refresh_token || '',
          }))
        }
        // After signup, always go to project creation first
        const redirect = $page.url.searchParams.get('redirect')
        await goto(redirect ? `/projects/new?redirect=${encodeURIComponent(redirect)}` : '/projects/new')
      } else {
        error = (data as any).message || 'Sign up failed'
      }
    } catch (err) {
      error = 'Connection error — is the API running?'
    } finally {
      loading = false
    }
  }
</script>

<svelte:head>
  <title>Create account — OmniBase</title>
</svelte:head>

<div class="login-page">
  <div class="login-card">
    <a href="/" class="logo-link">
      <span class="logo-icon">Ω</span>
      <span>OmniBase</span>
    </a>
    <h1>Create account</h1>
    <p class="subtitle">Sign up to access the OmniBase dashboard</p>

    <form onsubmit={handleSubmit} class="login-form">
      {#if error}
        <div class="error-banner" role="alert">{error}</div>
      {/if}
      <div class="form-group">
        <label for="email">Email</label>
        <input
          id="email"
          type="email"
          bind:value={email}
          placeholder="you@example.com"
          autocomplete="email"
          required
          disabled={loading}
        />
      </div>
      <div class="form-group">
        <label for="password">Password</label>
        <input
          id="password"
          type="password"
          bind:value={password}
          placeholder="At least 8 characters"
          autocomplete="new-password"
          required
          minlength="8"
          disabled={loading}
        />
      </div>
      <div class="form-group">
        <label for="confirmPassword">Confirm password</label>
        <input
          id="confirmPassword"
          type="password"
          bind:value={confirmPassword}
          placeholder="Repeat password"
          autocomplete="new-password"
          required
          disabled={loading}
        />
      </div>
      <button type="submit" class="btn btn-primary btn-block" disabled={loading}>
        {loading ? 'Creating account...' : 'Create account'}
      </button>
    </form>

    <p class="signin-link">
      Already have an account? <a href={`/auth/login${$page.url.search}`}>Sign in</a>
    </p>
    <p class="api-hint">API: {getOmniBaseUrl()}</p>
  </div>
</div>

<style>
  .login-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-base);
    padding: 24px;
  }
  .login-card {
    width: 100%;
    max-width: 400px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    padding: 32px;
    box-shadow: var(--shadow-xl);
  }
  .logo-link {
    display: flex;
    align-items: center;
    gap: 10px;
    text-decoration: none;
    color: var(--text-primary);
    font-weight: 700;
    font-size: 20px;
    margin-bottom: 24px;
  }
  .logo-icon {
    width: 40px;
    height: 40px;
    background: var(--gradient-brand);
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 22px;
    color: #fff;
  }
  h1 { font-size: 22px; margin: 0 0 8px; }
  .subtitle { font-size: 14px; color: var(--text-muted); margin: 0 0 24px; }
  .login-form { display: flex; flex-direction: column; gap: 16px; }
  .error-banner {
    padding: 12px 14px;
    background: rgba(255, 82, 82, 0.1);
    border: 1px solid rgba(255, 82, 82, 0.3);
    border-radius: var(--radius-md);
    color: var(--status-error);
    font-size: 13px;
  }
  .form-group { display: flex; flex-direction: column; gap: 6px; }
  .form-group label { font-size: 13px; font-weight: 500; color: var(--text-secondary); }
  .form-group input {
    padding: 10px 14px;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    background: var(--bg-elevated);
    color: var(--text-primary);
    font-size: 14px;
  }
  .form-group input:focus {
    outline: none;
    border-color: var(--brand-primary);
  }
  .btn-block { width: 100%; margin-top: 8px; padding: 12px; }
  .signin-link { margin-top: 16px; font-size: 14px; color: var(--text-muted); text-align: center; }
  .signin-link a { color: var(--brand-primary); text-decoration: none; }
  .api-hint { margin-top: 20px; font-size: 11px; color: var(--text-muted); font-family: var(--font-mono); }
</style>
