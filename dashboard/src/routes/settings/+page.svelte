<script lang="ts">
  import { onMount } from 'svelte'
  import { getHeaders, getOmniBaseUrl } from '$lib/api'
  import { authStore } from '$lib/stores/auth'

  let url = $state(getOmniBaseUrl())
  let anonKey = $state('')
  let serviceKey = $state('')
  let loading = $state(true)
  let error = $state<string | null>(null)
  let apiUrlInput = $state('')

  function copy(text: string) {
    navigator.clipboard.writeText(text)
  }

  function saveApiUrl() {
    const u = apiUrlInput.trim() || url
    if (u) {
      if (typeof localStorage !== 'undefined') localStorage.setItem('omnibase.api_url', u)
      url = u
      window.location.reload()
    }
  }

  async function loadConfig() {
    try {
      loading = true
      apiUrlInput = getOmniBaseUrl()
      if ($authStore.user) {
        const resp = await fetch(`${getOmniBaseUrl()}/admin/v1/config`, { headers: getHeaders() })
        const data = await resp.json()
        if (!resp.ok) throw new Error(data.message || 'Failed to load project configuration')
        url = data.url ?? url
        anonKey = data.anonKey ?? ''
        serviceKey = data.serviceKey ?? ''
      }
      error = null
    } catch (err: any) {
      error = err.message || 'Failed to load project configuration'
    } finally {
      loading = false
    }
  }

  onMount(loadConfig)
</script>

<svelte:head>
  <title>API Settings — OmniBase</title>
</svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">API Settings</h1>
    <p class="page-subtitle">Project URL and API keys for the OmniBase client SDKs</p>
  </div>
</div>

<div class="page-content" style="max-width: 800px;">
  {#if error}
    <div class="card" style="margin-bottom: 24px; border-color: var(--status-error); color: var(--status-error);">
      {error}
    </div>
  {/if}

  <div class="card" style="margin-bottom: 24px;">
    <h2 style="font-size: 15px; font-weight: 600; margin-bottom: 8px;">Project URL</h2>
    <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 16px;">OmniBase API gateway URL. Change this if you are connecting to a different instance.</p>
    <div style="display: flex; gap: 8px; margin-bottom: 12px;">
      <input type="text" class="input" bind:value={apiUrlInput} placeholder="http://localhost:8000" style="flex: 1; font-family: var(--font-mono);" />
      <button class="btn btn-primary" onclick={saveApiUrl} type="button">Save</button>
      <button class="btn btn-secondary" onclick={() => copy(url)} disabled={!url}>Copy</button>
    </div>
    <p style="font-size: 12px; color: var(--text-muted);">Current: <code style="font-family: var(--font-mono);">{url}</code></p>
  </div>

  <div class="card">
    <h2 style="font-size: 15px; font-weight: 600; margin-bottom: 8px;">Project API Keys</h2>
    {#if !$authStore.user}
      <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 16px;">
        <a href="/auth/login">Sign in</a> to view and copy your anon and service role keys.
      </p>
    {:else}
    <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 24px;">
      Your API keys allow you to authenticate with the OmniBase backend from your client apps or server.
    </p>

    <!-- Anon Key -->
    <div style="margin-bottom: 24px;">
      <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
        <span class="badge badge-success">anon</span>
        <span class="badge badge-neutral">public</span>
      </div>
      <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 8px;">
        This key is safe to use in a browser if you have enabled Row Level Security (RLS) for your tables and configured policies.
      </p>
      <div style="display: flex; gap: 8px;">
        <input type="password" class="input" value={loading ? 'Loading...' : anonKey} readonly style="flex: 1; font-family: var(--font-mono);" />
        <button class="btn btn-secondary" onclick={() => copy(anonKey)} disabled={loading || !anonKey}>Copy</button>
      </div>
    </div>

    <hr style="border: 0; border-top: 1px solid var(--border-subtle); margin: 24px 0;" />

    <!-- Service Role Key -->
    <div>
      <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
        <span class="badge badge-error" style="background: rgba(255,82,82,0.1); color: #ff5252;">service_role</span>
        <span class="badge badge-neutral">secret</span>
      </div>
      <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 8px;">
        This key has the ability to bypass Row Level Security. <strong>Never</strong> share it publicly or use it in the browser.
      </p>
      <div style="display: flex; gap: 8px;">
        <input type="password" class="input" value={loading ? 'Loading...' : serviceKey} readonly style="flex: 1; font-family: var(--font-mono);" />
        <button class="btn btn-secondary" onclick={() => copy(serviceKey)} disabled={loading || !serviceKey}>Copy</button>
      </div>
    </div>
    {/if}
  </div>
</div>
