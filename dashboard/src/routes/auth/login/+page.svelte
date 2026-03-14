<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/stores'
  import { authStore } from '$lib/stores/auth'
  import { getOmniBaseUrl } from '$lib/api'

  let email = $state('')
  let password = $state('')
  let error = $state<string | null>(null)
  let loading = $state(false)
  let magicLinkMode = $state(false)
  let magicLinkSent = $state(false)

  async function handleSubmit(e: Event) {
    e.preventDefault()
    error = null
    if (!email.trim() || !password) {
      error = 'Email and password are required'
      return
    }
    loading = true
    try {
      const result = await authStore.signIn(email, password)
      if (result.error) {
        error = (result.error as any).message || 'Invalid email or password'
        return
      }
      const redirect = $page.url.searchParams.get('redirect')
      await goto(redirect || '/')
    } catch (err) {
      error = 'Connection error — is the API running?'
    } finally {
      loading = false
    }
  }

  async function sendMagicLink(e: Event) {
    e.preventDefault()
    error = null
    if (!email.trim()) {
      error = 'Email is required'
      return
    }
    loading = true
    try {
      const redirectPath = $page.url.searchParams.get('redirect') || '/'
      const redirectTo = typeof window !== 'undefined' ? window.location.origin + redirectPath : redirectPath
      const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/magiclink`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: email.trim(), redirect_to: redirectTo }),
      })
      const data = await resp.json()
      if (resp.ok) {
        magicLinkSent = true
      } else {
        error = (data as any).message || 'Failed to send link'
      }
    } catch (err) {
      error = 'Connection error — is the API running?'
    } finally {
      loading = false
    }
  }
</script>

<svelte:head>
  <title>Sign in — OmniBase</title>
</svelte:head>

<div class="login-page">
  <div class="login-card">
    <a href="/" class="logo-link">
      <span class="logo-icon">Ω</span>
      <span>OmniBase</span>
    </a>
    <h1>Sign in</h1>
    <p class="subtitle">Use your OmniBase account to access the dashboard</p>

    {#if magicLinkSent}
      <div class="success-banner">
        Check your email for a sign-in link. Click it to sign in (link expires in 1 hour).
      </div>
      <p class="signin-link">
        <button type="button" class="link-btn" onclick={() => { magicLinkSent = false }}>Send another link</button>
      </p>
    {:else if magicLinkMode}
      <form onsubmit={sendMagicLink} class="login-form">
        {#if error}
          <div class="error-banner" role="alert">{error}</div>
        {/if}
        <div class="form-group">
          <label for="magic-email">Email</label>
          <input
            id="magic-email"
            type="email"
            bind:value={email}
            placeholder="you@example.com"
            autocomplete="email"
            required
            disabled={loading}
          />
        </div>
        <button type="submit" class="btn btn-primary btn-block" disabled={loading}>
          {loading ? 'Sending...' : 'Send magic link'}
        </button>
      </form>
      <p class="signin-link">
        <button type="button" class="link-btn" onclick={() => { magicLinkMode = false; error = null }}>← Back to password</button>
      </p>
    {:else}
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
            placeholder="••••••••"
            autocomplete="current-password"
            required
            disabled={loading}
          />
        </div>
        <p class="forgot-row">
          <a href="/auth/forgot-password" class="forgot-link">Forgot password?</a>
        </p>
        <button type="submit" class="btn btn-primary btn-block" disabled={loading}>
          {loading ? 'Signing in...' : 'Sign in'}
        </button>
      </form>
      <p class="signin-link">
        <button type="button" class="link-btn" onclick={() => { magicLinkMode = true; error = null }}>Sign in with magic link (no password)</button>
      </p>
    {/if}

    <p class="signin-link">
      Don't have an account? <a href={`/auth/signup${$page.url.search}`}>Create one</a>
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
  h1 {
    font-size: 22px;
    margin: 0 0 8px;
  }
  .subtitle {
    font-size: 14px;
    color: var(--text-muted);
    margin: 0 0 24px;
  }
  .login-form {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .error-banner {
    padding: 12px 14px;
    background: rgba(255, 82, 82, 0.1);
    border: 1px solid rgba(255, 82, 82, 0.3);
    border-radius: var(--radius-md);
    color: var(--status-error);
    font-size: 13px;
  }
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .form-group label {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-secondary);
  }
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
  .btn-block {
    width: 100%;
    margin-top: 8px;
    padding: 12px;
  }
  .signin-link {
    margin-top: 16px;
    font-size: 14px;
    color: var(--text-muted);
    text-align: center;
  }
  .signin-link a {
    color: var(--brand-primary);
    text-decoration: none;
  }
  .link-btn {
    background: none;
    border: none;
    color: var(--brand-primary);
    cursor: pointer;
    font-size: 14px;
    padding: 0;
  }
  .forgot-row { margin: -8px 0 0; text-align: right; }
  .forgot-link { font-size: 13px; color: var(--text-muted); text-decoration: none; }
  .success-banner {
    padding: 14px;
    background: rgba(74, 222, 128, 0.1);
    border: 1px solid rgba(74, 222, 128, 0.3);
    border-radius: var(--radius-md);
    color: var(--status-success);
    font-size: 13px;
    margin-bottom: 16px;
  }
  .api-hint {
    margin-top: 20px;
    font-size: 11px;
    color: var(--text-muted);
    font-family: var(--font-mono);
  }
</style>
