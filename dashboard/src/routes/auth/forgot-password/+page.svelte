<script lang="ts">
  import { getOmniBaseUrl } from '$lib/api'

  let email = $state('')
  let error = $state<string | null>(null)
  let loading = $state(false)
  let sent = $state(false)

  async function handleSubmit(e: Event) {
    e.preventDefault()
    error = null
    if (!email.trim()) {
      error = 'Email is required'
      return
    }
    loading = true
    try {
      const redirectTo = typeof window !== 'undefined' ? window.location.origin + '/auth/reset-password' : '/auth/reset-password'
      const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/recover`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: email.trim(), redirect_to: redirectTo }),
      })
      const data = await resp.json()
      if (resp.ok) {
        sent = true
      } else {
        error = (data as any).message || 'Failed to send reset link'
      }
    } catch (err) {
      error = 'Connection error — is the API running?'
    } finally {
      loading = false
    }
  }
</script>

<svelte:head>
  <title>Forgot password — OmniBase</title>
</svelte:head>

<div class="page">
  <div class="card">
    <a href="/auth/login" class="logo-link">← Back to sign in</a>
    <h1>Forgot password</h1>
    <p class="subtitle">Enter your email and we'll send a link to reset your password.</p>
    {#if sent}
      <div class="success-banner">Check your email for a reset link (valid 1 hour).</div>
      <a href="/auth/login" class="btn btn-primary btn-block">Back to sign in</a>
    {:else}
      <form onsubmit={handleSubmit} class="form">
        {#if error}
          <div class="error-banner" role="alert">{error}</div>
        {/if}
        <div class="form-group">
          <label for="email">Email</label>
          <input id="email" type="email" bind:value={email} placeholder="you@example.com" required disabled={loading} />
        </div>
        <button type="submit" class="btn btn-primary btn-block" disabled={loading}>
          {loading ? 'Sending...' : 'Send reset link'}
        </button>
      </form>
    {/if}
  </div>
</div>

<style>
  .page { min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 24px; background: var(--bg-base); }
  .card { width: 100%; max-width: 400px; background: var(--bg-surface); border: 1px solid var(--border-subtle); border-radius: var(--radius-lg); padding: 32px; }
  .logo-link { display: inline-block; margin-bottom: 20px; font-size: 14px; color: var(--text-muted); text-decoration: none; }
  h1 { font-size: 22px; margin: 0 0 8px; }
  .subtitle { font-size: 14px; color: var(--text-muted); margin: 0 0 24px; }
  .form { display: flex; flex-direction: column; gap: 16px; }
  .error-banner { padding: 12px; background: rgba(255,82,82,0.1); border-radius: var(--radius-md); color: var(--status-error); font-size: 13px; }
  .success-banner { padding: 14px; background: rgba(74,222,128,0.1); border-radius: var(--radius-md); color: var(--status-success); margin-bottom: 16px; }
  .form-group { display: flex; flex-direction: column; gap: 6px; }
  .form-group label { font-size: 13px; font-weight: 500; color: var(--text-secondary); }
  .form-group input { padding: 10px 14px; border: 1px solid var(--border-subtle); border-radius: var(--radius-md); background: var(--bg-elevated); color: var(--text-primary); }
  .btn-block { width: 100%; padding: 12px; text-align: center; text-decoration: none; display: block; }
</style>
