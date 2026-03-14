<script lang="ts">
  import { onMount } from 'svelte'
  import { getHeaders, getOmniBaseUrl } from '$lib/api'

  interface Provider {
    id: string
    name: string
    enabled: boolean
    client_id_configured?: boolean
  }

  let providers = $state<Provider[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)

  async function loadProviders() {
    try {
      loading = true
      error = null
      const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/providers`, { headers: getHeaders(false) })
      if (resp.ok) {
        const data = await resp.json()
        providers = Array.isArray(data) ? data : (data.providers || [])
      } else {
        const d = await resp.json().catch(() => ({}))
        error = (d as any).message || 'Failed to load providers'
        providers = []
      }
    } catch (e) {
      error = 'Connection error — is the API running?'
      providers = []
    } finally {
      loading = false
    }
  }

  onMount(loadProviders)
</script>

<svelte:head>
  <title>Auth Providers — OmniBase</title>
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">Auth Providers</h1>
    <p class="page-subtitle">OAuth and identity providers available for sign-in</p>
  </div>
  <button class="btn btn-secondary btn-sm" onclick={loadProviders} disabled={loading}>
    {loading ? 'Loading...' : 'Refresh'}
  </button>
</div>

<div class="page-content">
  {#if error}
    <div class="card" style="border-color: var(--status-error); margin-bottom: 24px;">
      <p style="color: var(--status-error);">{error}</p>
    </div>
  {/if}

  <div class="card animate-fade-in">
    <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 20px;">
      Configure OAuth providers via environment variables. Set <code>GOOGLE_CLIENT_ID</code>, <code>GOOGLE_CLIENT_SECRET</code>,
      <code>GITHUB_CLIENT_ID</code>, <code>GITHUB_CLIENT_SECRET</code> in your <code>.env</code> or Docker environment, then restart the auth service.
    </p>
    {#if loading}
      <div class="skeleton" style="height: 200px; border-radius: 12px;"></div>
    {:else}
      <div class="table-wrapper">
        <table>
          <thead>
            <tr>
              <th>Provider</th>
              <th>Status</th>
              <th>Notes</th>
            </tr>
          </thead>
          <tbody>
            {#each providers.length ? providers : [
              { id: 'google', name: 'Google', enabled: false },
              { id: 'github', name: 'GitHub', enabled: false },
            ] as p}
              <tr>
                <td style="font-weight: 500;">{p.name}</td>
                <td>
                  <span class="badge {p.enabled || p.client_id_configured ? 'badge-success' : 'badge-neutral'}">
                    {p.enabled || p.client_id_configured ? 'Configured' : 'Not configured'}
                  </span>
                </td>
                <td style="font-size: 12px; color: var(--text-muted);">
                  {#if p.id === 'google'}
                    Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET
                  {:else if p.id === 'github'}
                    Set GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET
                  {:else}
                    —
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>

  <div class="card animate-fade-in" style="margin-top: 24px;">
    <h2 style="font-size: 15px; margin-bottom: 12px;">Sign-in URL</h2>
    <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 12px;">
      Users can sign in with a provider by visiting:
    </p>
    <code style="display: block; padding: 12px; background: var(--bg-elevated); border-radius: var(--radius-md); font-size: 12px;">
      {getOmniBaseUrl()}/auth/v1/authorize?provider=google&redirect_to={encodeURIComponent('http://localhost:3001')}
    </code>
    <p style="font-size: 12px; color: var(--text-muted); margin-top: 12px;">
      Replace <code>provider</code> with <code>google</code> or <code>github</code>, and <code>redirect_to</code> with your app URL.
    </p>
  </div>
</div>
