<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/stores'
  import { getOmniBaseUrl } from '$lib/api'

  let token = $state('')
  let password = $state('')
  let confirmPassword = $state('')
  let error = $state<string | null>(null)
  let loading = $state(false)
  let success = $state(false)

  $effect(() => {
    const t = $page.url.searchParams.get('token')
    if (t) token = t
  })

  async function handleSubmit(e: Event) {
    e.preventDefault()
    error = null
    if (!token) {
      error = 'Reset link is invalid or expired. Request a new one from the forgot password page.'
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
      const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/reset-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, password }),
      })
      const data = await resp.json()
      if (resp.ok) {
        success = true
        setTimeout(() => goto('/auth/login'), 2000)
      } else {
        error = (data as any).message || 'Failed to reset password'
      }
    } catch (err) {
      error = 'Connection error'
    } finally {
      loading = false
    }
  }
</script>

<svelte:head>
  <title>Reset password — OmniBase</title>
</svelte:head>

<div class="page">
  <div class="card">
    <a href="/auth/login" class="logo-link">← Back to sign in</a>
    <h1>Reset password</h1>
    {#if success}
      <div class="success-banner">Password updated. Redirecting to sign in...</div>
    {:else}
      <form onsubmit={handleSubmit} class="form">
        {#if error}
          <div class="error-banner" role="alert">{error}</div>
        {/if}
        {#if !token}
          <p class="subtitle">Use the link from your email, or <a href="/auth/forgot-password">request a new reset link</a>.</p>
        {:else}
          <div class="form-group">
            <label for="password">New password</label>
            <input id="password" type="password" bind:value={password} placeholder="At least 8 characters" minlength="8" required disabled={loading} />
          </div>
          <div class="form-group">
            <label for="confirm">Confirm password</label>
            <input id="confirm" type="password" bind:value={confirmPassword} placeholder="Repeat password" required disabled={loading} />
          </div>
          <button type="submit" class="btn btn-primary btn-block" disabled={loading}>
            {loading ? 'Updating...' : 'Update password'}
          </button>
        {/if}
      </form>
    {/if}
  </div>
</div>

<style>
  .page { min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 24px; background: var(--bg-base); }
  .card { width: 100%; max-width: 400px; background: var(--bg-surface); border: 1px solid var(--border-subtle); border-radius: var(--radius-lg); padding: 32px; }
  .logo-link { display: inline-block; margin-bottom: 20px; font-size: 14px; color: var(--text-muted); text-decoration: none; }
  h1 { font-size: 22px; margin: 0 0 16px; }
  .subtitle { font-size: 14px; color: var(--text-muted); margin-bottom: 16px; }
  .subtitle a { color: var(--brand-primary); }
  .form { display: flex; flex-direction: column; gap: 16px; }
  .error-banner { padding: 12px; background: rgba(255,82,82,0.1); border-radius: var(--radius-md); color: var(--status-error); font-size: 13px; }
  .success-banner { padding: 14px; background: rgba(74,222,128,0.1); border-radius: var(--radius-md); color: var(--status-success); }
  .form-group { display: flex; flex-direction: column; gap: 6px; }
  .form-group label { font-size: 13px; font-weight: 500; color: var(--text-secondary); }
  .form-group input { padding: 10px 14px; border: 1px solid var(--border-subtle); border-radius: var(--radius-md); background: var(--bg-elevated); color: var(--text-primary); }
  .btn-block { width: 100%; padding: 12px; }
</style>
