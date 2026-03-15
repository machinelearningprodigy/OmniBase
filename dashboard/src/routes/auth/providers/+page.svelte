<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getOmniBaseUrl, getErrorMessage } from '$lib/api'
  import { authStore } from '$lib/stores/auth'
  import { fade, slide, scale } from 'svelte/transition'

  interface Provider {
    id?: string
    name: string
    type: string
    client_id: string
    client_secret: string
    is_active: boolean
    metadata_url?: string
  }

  let providers = $state<Provider[]>([])
  let loading = $state(true)
  let saving = $state(false)
  let error = $state<string | null>(null)
  let success = $state<string | null>(null)

  let selectedProvider = $state<Provider | null>(null)
  let showModal = $state(false)

  const icons: Record<string, string> = {
    google: `<svg viewBox="0 0 24 24" width="20" height="20"><path fill="#EA4335" d="M5.266 9.765A7.077 7.077 0 0 1 12 4.909c1.69 0 3.218.6 4.418 1.582L19.91 3C17.782 1.145 15.055 0 12 0 7.27 0 3.198 2.698 1.24 6.65l4.026 3.115z"/><path fill="#34A853" d="M16.04 18.013c-1.09.693-2.459 1.096-4.04 1.096-3.132 0-5.833-2.123-6.777-5.006l-4.04 3.136C3.12 21.19 7.21 24 12 24c3.15 0 5.86-1.05 7.9-2.85l-3.86-3.137z"/><path fill="#4285F4" d="M19.9 21.15c2.53-2.23 3.97-5.52 3.97-9.15 0-.82-.07-1.61-.21-2.38H12v4.51h6.68c-.29 1.56-1.17 2.87-2.5 3.75l3.72 3.27z"/><path fill="#FBBC05" d="M5.263 14.103a7.03 7.03 0 0 1 0-4.343l-4.024-3.116A11.977 11.977 0 0 0 0 12c0 1.89.44 3.666 1.22 5.253l4.043-3.15z"/></svg>`,
    github: `<svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>`,
    discord: `<svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037 19.736 19.736 0 0 0-4.885 1.515.069.069 0 0 0-.032.027C.533 9.048-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028c.462-.63.862-1.295 1.198-1.996a.076.076 0 0 0-.041-.106 13.096 13.096 0 0 1-1.873-.894.077.077 0 0 1-.008-.128c.125-.094.252-.192.372-.29a.074.074 0 0 1 .077-.01 12.41 12.41 0 0 0 10.962 0 .074.074 0 0 1 .077.01c.12.098.246.196.372.29a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.874.894.077.077 0 0 0-.041.107c.337.7.737 1.365 1.198 1.996a.078.078 0 0 0 .084.028 20.016 20.016 0 0 0 5.994-3.03.076.076 0 0 0 .031-.056c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.027z"/></svg>`,
    facebook: `<svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/></svg>`,
    microsoft: `<svg viewBox="0 0 23 23" width="20" height="20"><path fill="#f35325" d="M1 1h10v10H1z"/><path fill="#81bc06" d="M12 1h10v10H12z"/><path fill="#05a6f0" d="M1 12h10v10H1z"/><path fill="#ffba08" d="M12 12h10v10H12z"/></svg>`,
    apple: `<svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor"><path d="M17.066 11.355c0-2.29 1.86-3.385 1.942-3.442-1.072-1.558-2.723-1.77-3.303-1.794-1.385-.141-2.712.825-3.415.825-.702 0-1.8-.797-2.964-.774-1.528.024-2.937.893-3.723 2.261-1.583 2.756-.407 6.837 1.13 9.06 0.753 1.087 1.637 2.305 2.809 2.262 1.127-.044 1.556-.728 2.915-.728 1.356 0 1.748.728 2.94.704 1.21-.02 1.97-1.107 2.716-2.197 0.865-1.258 1.22-2.474 1.242-2.536-.026-.011-2.388-.916-2.388-3.642zM14.53 4.296c.62-.751 1.037-1.794.922-2.835-.89.037-1.968.599-2.607 1.35-.572.664-1.073 1.728-.937 2.747.99.076 2.0-.51 2.622-1.262z"/></svg>`,
  }

  const availableOAuth = [
    { id: 'google', name: 'Google', color: '#4285F4' },
    { id: 'github', name: 'GitHub', color: '#ffffff' },
    { id: 'discord', name: 'Discord', color: '#5865F2' },
    { id: 'facebook', name: 'Facebook', color: '#1877F2' },
    { id: 'microsoft', name: 'Microsoft', color: '#00A4EF' },
    { id: 'apple', name: 'Apple', color: '#ffffff' },
  ]

  async function loadProviders() {
    try {
      loading = true
      error = null
      // Use apiFetch instead of raw fetch
      const resp = await apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/oauth/apps`)
      if (resp.ok) {
        providers = await resp.json()
      } else {
        error = await getErrorMessage(resp, 'Failed to load providers')
      }
    } catch (e) {
      error = `Connection failed: ${e instanceof Error ? e.message : 'Unknown error'}. Ensure OmniBase Gateway is running at ${getOmniBaseUrl()}`
    } finally {
      loading = false
    }
  }

  async function saveProvider() {
    if (!selectedProvider) return
    try {
      saving = true
      error = null
      success = null
      const resp = await apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/oauth/apps`, {
        method: 'POST',
        body: JSON.stringify(selectedProvider)
      })
      if (resp.ok) {
        success = `${selectedProvider.name.toUpperCase()} identity updated`
        showModal = false
        await loadProviders()
      } else {
        error = await getErrorMessage(resp, 'Failed to save configuration')
      }
    } catch (e) {
      error = 'Request failed'
    } finally {
      saving = false
    }
  }

  function openConfig(id: string) {
    const existing = providers.find(p => p.name.toLowerCase() === id)
    selectedProvider = {
      name: id,
      type: 'oauth',
      client_id: existing?.client_id || '',
      client_secret: existing?.client_secret || '',
      is_active: existing?.is_active ?? true
    }
    showModal = true
  }

  onMount(loadProviders)

  function isConfigured(id: string) {
    return providers.some(p => p.name.toLowerCase() === id && p.client_id)
  }

  function isActive(id: string) {
    return providers.find(p => p.name.toLowerCase() === id)?.is_active ?? false
  }

  let copySuccess = $state(false)
  async function copyToClipboard(text: string) {
    await navigator.clipboard.writeText(text)
    copySuccess = true
    setTimeout(() => copySuccess = false, 2000)
  }
</script>

<svelte:head>
  <title>Auth Settings — OmniBase</title>
</svelte:head>

<main class="page-container" transition:fade>
  <header class="header">
    <div class="breadcrumb">Settings / <span class="highlight">Authentication</span></div>
    <div class="header-row">
      <div>
        <h1 class="title">Identity Providers</h1>
        <p class="subtitle">Enable social login and manage OAuth configuration.</p>
      </div>
      <button class="btn-refresh {loading ? 'spinning' : ''}" onclick={loadProviders}>
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>
      </button>
    </div>
  </header>

  {#if error}
    <div class="toast error" transition:slide>
      <div class="toast-content">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="3"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        <span>{error}</span>
      </div>
      <button class="toast-close" onclick={() => error = null}>&times;</button>
    </div>
  {/if}

  {#if success}
    <div class="toast success" transition:slide>
      <div class="toast-content">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
        <span>{success}</span>
      </div>
      <button class="toast-close" onclick={() => success = null}>&times;</button>
    </div>
  {/if}

  <div class="section-label">External Providers</div>
  <div class="provider-grid">
    {#each availableOAuth as item}
      {@const configured = isConfigured(item.id)}
      {@const active = isActive(item.id)}
      <button class="card {configured ? 'configured' : ''}" onclick={() => openConfig(item.id)}>
        <div class="card-main">
          <div class="card-icon" style="--icon-color: {item.color}">
            {@html icons[item.id]}
            {#if configured && active}
              <div class="status-indicator"></div>
            {/if}
          </div>
          <div class="card-text">
            <div class="provider-name">{item.name}</div>
            <div class="provider-status {configured ? (active ? 'on' : 'paused') : 'off'}">
              {configured ? (active ? 'Enabled' : 'Paused') : 'Setup Needed'}
            </div>
          </div>
        </div>
        <div class="card-action">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3"><polyline points="9 18 15 12 9 6"/></svg>
        </div>
      </button>
    {/each}
  </div>

  <div class="guide-box" transition:fade={{ delay: 100 }}>
    <div class="guide-header">
       <div class="guide-icon"><svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg></div>
       <div class="guide-titles">
         <h3>Authorization Endpoint</h3>
         <p>Standard URL to initiate login from your client apps.</p>
       </div>
    </div>
    <div class="code-bar">
      <code class="code">
        {getOmniBaseUrl()}/auth/v1/authorize?provider=google&project_id={$authStore.activeProject?.id || '...'}&redirect_to=YOUR_APP_URL
      </code>
      <button class="copy-btn {copySuccess ? 'copied' : ''}" onclick={() => copyToClipboard(`${getOmniBaseUrl()}/auth/v1/authorize?provider=google&project_id=${$authStore.activeProject?.id || 'PROJECT_ID'}&redirect_to=https://yourapp.com`)}>
        {#if copySuccess}<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>{:else}<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>{/if}
      </button>
    </div>
    <p class="guide-note">Whitelist your <code>redirect_to</code> domain in the provider's developer console.</p>
  </div>
</main>

{#if showModal && selectedProvider}
  <div class="modal-overlay" onclick={() => showModal = false} transition:fade={{ duration: 150 }}>
    <div class="modal-content" onclick={e => e.stopPropagation()} in:scale={{ start: 0.98, duration: 200 }}>
      <header class="modal-header">
        <div class="header-info">
          <div class="id-icon" style="background: {availableOAuth.find(a => a.id === selectedProvider?.name)?.color}10; color: {availableOAuth.find(a => a.id === selectedProvider?.name)?.color}">
            {@html icons[selectedProvider.name]}
          </div>
          <div>
            <h2>{selectedProvider.name.toUpperCase()} Config</h2>
            <p>Enter your application credentials below.</p>
          </div>
        </div>
        <button class="modal-close" onclick={() => showModal = false}>&times;</button>
      </header>

      <div class="modal-body">
        <div class="form-field">
          <label for="c-id">Client ID</label>
          <input type="text" id="c-id" bind:value={selectedProvider.client_id} placeholder="Public identifier from provider..." />
        </div>
        <div class="form-field">
          <label for="c-sec">Client Secret</label>
          <div class="password-wrap">
            <input type="password" id="c-sec" bind:value={selectedProvider.client_secret} placeholder="••••••••••••••••" />
            <svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor"><path d="M12 2a5 5 0 0 0-5 5v3H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8a2 2 0 0 0-2-2h-1V7a5 5 0 0 0-5-5z"/></svg>
          </div>
        </div>

        <div class="divider"></div>

        <div class="toggle-row">
          <div class="toggle-info">
            <div class="toggle-title">Status</div>
            <div class="toggle-desc">Allow users to use this provider</div>
          </div>
          <label class="switch">
            <input type="checkbox" bind:checked={selectedProvider.is_active} />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <footer class="modal-footer">
        <button class="btn-ghost" onclick={() => showModal = false}>Cancel</button>
        <button class="btn-primary" onclick={saveProvider} disabled={saving}>
          {saving ? 'Saving...' : 'Save Configuration'}
        </button>
      </footer>
    </div>
  </div>
{/if}

<style>
  :global(:root) {
    --primary: #6c47ff;
    --primary-dim: rgba(108, 71, 255, 0.1);
    --bg: #050507;
    --card: #0c0c11;
    --card-hover: #14141d;
    --border: rgba(255, 255, 255, 0.08);
    --text-sec: #8d8d9f;
    --text-prim: #ffffff;
  }

  .page-container {
    max-width: 900px;
    margin: 0 auto;
    padding: 40px 24px;
    font-family: 'Inter', system-ui, sans-serif;
    color: var(--text-prim);
  }

  .breadcrumb { font-size: 11px; font-weight: 700; color: var(--text-sec); text-transform: uppercase; letter-spacing: 0.1em; margin-bottom: 8px; }
  .breadcrumb .highlight { color: var(--primary); }

  .header-row { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 40px; }
  .title { font-size: 32px; font-weight: 850; letter-spacing: -0.03em; margin: 0 0 4px; }
  .subtitle { font-size: 15px; color: var(--text-sec); margin: 0; }

  .btn-refresh {
    width: 38px; height: 38px; border-radius: 12px;
    background: var(--card); border: 1px solid var(--border);
    color: var(--text-sec); cursor: pointer; display: grid; place-items: center;
    transition: 0.2s;
  }
  .btn-refresh:hover { color: #fff; background: var(--card-hover); border-color: rgba(255,255,255,0.2); }
  .btn-refresh.spinning svg { animation: spin 0.8s linear infinite; }
  @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

  /* Toasts */
  .toast {
    display: flex; align-items: center; justify-content: space-between;
    padding: 12px 16px; border-radius: 12px; margin-bottom: 24px;
    border: 1px solid transparent; font-size: 13px; font-weight: 600;
  }
  .toast.error { background: rgba(239, 68, 68, 0.08); border-color: rgba(239, 68, 68, 0.2); color: #f87171; }
  .toast.success { background: rgba(34, 197, 94, 0.08); border-color: rgba(34, 197, 94, 0.2); color: #4ade80; }
  .toast-content { display: flex; align-items: center; gap: 10px; }
  .toast-close { background: none; border: none; color: inherit; font-size: 18px; cursor: pointer; opacity: 0.5; }

  /* Grid */
  .section-label { font-size: 12px; font-weight: 800; color: var(--text-sec); text-transform: uppercase; letter-spacing: 0.08em; margin-bottom: 16px; }
  .provider-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 12px; margin-bottom: 40px; }

  .card {
    background: var(--card); border: 1px solid var(--border); border-radius: 14px;
    padding: 14px; display: flex; align-items: center; justify-content: space-between;
    transition: 0.2s; cursor: pointer; text-align: left;
  }
  .card:hover { border-color: rgba(255,255,255,0.15); background: var(--card-hover); transform: translateY(-1px); }
  .card.configured { border-color: rgba(108, 71, 255, 0.2); }

  .card-main { display: flex; align-items: center; gap: 12px; }
  .card-icon {
    width: 36px; height: 36px; border-radius: 10px; background: rgba(255,255,255,0.03);
    display: grid; place-items: center; position: relative; color: var(--icon-color);
  }
  .status-indicator {
    position: absolute; top: -2px; right: -2px; width: 7px; height: 7px;
    background: #4ade80; border-radius: 50%; border: 1.5px solid var(--card);
  }
  .provider-name { font-size: 14px; font-weight: 700; }
  .provider-status { font-size: 11px; font-weight: 600; margin-top: 1px; }
  .provider-status.on { color: #4ade80; }
  .provider-status.paused { color: #facc15; }
  .provider-status.off { color: var(--text-sec); opacity: 0.5; }

  .card-action { opacity: 0.2; transition: 0.2s; transform: translateX(-4px); }
  .card:hover .card-action { opacity: 0.6; transform: translateX(0); }

  /* Guide */
  .guide-box { background: var(--card); border: 1px solid var(--border); border-radius: 20px; padding: 24px; }
  .guide-header { display: flex; align-items: center; gap: 16px; margin-bottom: 20px; }
  .guide-icon { width: 38px; height: 38px; background: var(--primary); border-radius: 12px; display: grid; place-items: center; color: #fff; }
  .guide-titles h3 { font-size: 16px; font-weight: 800; margin: 0 0 2px; }
  .guide-titles p { font-size: 13px; color: var(--text-sec); margin: 0; }

  .code-bar {
    background: #000; border: 1px solid var(--border); border-radius: 12px;
    padding: 6px 6px 6px 14px; display: flex; align-items: center; gap: 12px; margin-bottom: 12px;
  }
  .code { flex: 1; font-family: 'JetBrains Mono', monospace; font-size: 12px; color: #9aa5ff; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .copy-btn {
    width: 30px; height: 30px; border-radius: 8px; background: #111; border: 1px solid var(--border);
    color: var(--text-sec); cursor: pointer; display: grid; place-items: center; transition: 0.2s;
  }
  .copy-btn:hover { background: #222; color: #fff; }
  .copy-btn.copied { background: rgba(34, 197, 94, 0.1); color: #4ade80; border-color: #059669; }
  .guide-note { font-size: 12px; color: var(--text-sec); margin: 0; }
  .guide-note code { color: #f87171; background: rgba(248, 113, 113, 0.05); padding: 1px 4px; border-radius: 4px; }

  /* Modal */
  .modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.8); backdrop-filter: blur(10px); z-index: 1000; display: flex; align-items: center; justify-content: center; padding: 20px; }
  .modal-content { width: 100%; max-width: 420px; background: #0a0a0e; border: 1px solid rgba(255,255,255,0.12); border-radius: 24px; overflow: hidden; }

  .modal-header { padding: 24px; display: flex; justify-content: space-between; align-items: flex-start; }
  .header-info { display: flex; gap: 16px; align-items: center; }
  .id-icon { width: 44px; height: 44px; border-radius: 12px; display: grid; place-items: center; }
  .modal-header h2 { font-size: 18px; font-weight: 850; margin: 0 0 2px; }
  .modal-header p { font-size: 13px; color: var(--text-sec); margin: 0; }
  .modal-close { background: none; border: none; font-size: 24px; color: var(--text-sec); cursor: pointer; padding: 0 8px; align-self: flex-start; }

  .modal-body { padding: 0 24px 24px; }
  .form-field { margin-bottom: 20px; }
  .form-field label { display: block; font-size: 12px; font-weight: 700; color: var(--text-sec); margin-bottom: 8px; text-transform: uppercase; }
  .form-field input { width: 100%; background: #000; border: 1px solid var(--border); border-radius: 12px; padding: 10px 14px; color: #fff; font-size: 14px; transition: border-color 0.2s; }
  .form-field input:focus { border-color: var(--primary); outline: none; }
  .password-wrap { position: relative; }
  .password-wrap svg { position: absolute; right: 14px; top: 50%; translate: 0 -50%; opacity: 0.2; }

  .divider { height: 1px; background: var(--border); margin: 24px 0; }

  .toggle-row { display: flex; align-items: center; justify-content: space-between; }
  .toggle-title { font-size: 15px; font-weight: 700; color: #fff; }
  .toggle-desc { font-size: 12px; color: var(--text-sec); }

  .switch { position: relative; width: 42px; height: 22px; }
  .switch input { opacity: 0; width: 0; height: 0; }
  .slider { position: absolute; inset: 0; background: #27272a; border-radius: 22px; transition: 0.3s; cursor: pointer; }
  .slider::before { content: ''; position: absolute; height: 16px; width: 16px; left: 3px; bottom: 3px; background: #fff; border-radius: 50%; transition: 0.3s; }
  input:checked + .slider { background: var(--primary); }
  input:checked + .slider::before { transform: translateX(20px); }

  .modal-footer { padding: 20px 24px; background: rgba(255,255,255,0.015); border-top: 1px solid var(--border); display: flex; justify-content: flex-end; gap: 12px; }
  .btn-ghost { background: none; border: none; color: var(--text-sec); font-weight: 700; cursor: pointer; padding: 10px 16px; border-radius: 10px; }
  .btn-primary { background: var(--primary); color: #fff; border: none; font-weight: 800; font-size: 14px; padding: 10px 20px; border-radius: 12px; cursor: pointer; transition: 0.2s; }
  .btn-primary:hover { transform: translateY(-1px); background: #7d5fff; }
</style>
