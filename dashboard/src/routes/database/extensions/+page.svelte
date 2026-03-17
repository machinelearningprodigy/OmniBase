<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getOmniBaseUrl, getHeaders, getErrorMessage } from '$lib/api'

  interface ExtensionMeta {
    name: string
    schema: string
    version: string
    description: string
    installed: boolean
  }

  let extensions = $state<ExtensionMeta[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)
  let searchQuery = $state('')
  let actionLoading = $state<string | null>(null)

  async function loadExtensions() {
    try {
      loading = true
      error = null
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/extensions`, { headers: getHeaders() })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to load extensions'))
      extensions = await resp.json()
    } catch (e) {
      error = e instanceof Error ? e.message : 'Unknown error'
    } finally {
      loading = false
    }
  }

  async function toggleExtension(ext: ExtensionMeta) {
    if (!ext.installed) {
        if (!confirm(`Enable extension "${ext.name}"?`)) return
    } else {
        if (!confirm(`Disable extension "${ext.name}"? This might break views/functions that depend on it.`)) return
    }

    actionLoading = ext.name
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/extensions/${ext.installed ? 'disable' : 'enable'}${ext.installed ? `?name=${encodeURIComponent(ext.name)}` : ''}`, {
        method: ext.installed ? 'DELETE' : 'POST',
        headers: getHeaders(true),
        ...(ext.installed ? {} : { body: JSON.stringify({ name: ext.name, schema: 'extensions' }) }) // Install into the extensions schema mostly, or public. We default to 'public' if schema string isn't provided but passing 'extensions' is standard for supabase-like setups. Actually let's just pass 'public' for simplicity for now.
      })
      
      // Let's modify the body to pass public schema
      if (!ext.installed && resp.status === 400) {
           throw new Error(await getErrorMessage(resp, 'Failed to toggle extension'))
      } else if (!resp.ok) {
           throw new Error(await getErrorMessage(resp, 'Failed to toggle extension'))
      }
      
      await loadExtensions()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      actionLoading = null
    }
  }

  onMount(() => {
    loadExtensions()
  })

  // Known logo mapping or colorful badge categories for popular extensions
  function getExtCategory(name: string) {
      if (name.includes('vector') || name.includes('ai') || name.includes('bm25')) return 'AI/Search'
      if (name.includes('graphql') || name.includes('rest')) return 'API'
      if (name.includes('fdw') || name.includes('cron') || name.includes('audit')) return 'Infrastructure'
      if (name.includes('postgis') || name.includes('tiger')) return 'Geospatial'
      if (name.includes('pgcrypto') || name.includes('auth')) return 'Security'
      return 'Utility'
  }

  let filteredExtensions = $derived(
      extensions.filter(ext => 
          ext.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
          ext.description.toLowerCase().includes(searchQuery.toLowerCase())
      )
  )
</script>

<svelte:head><title>Extensions — OmniBase</title></svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">Extensions</h1>
    <p class="page-subtitle">Enable and manage PostgreSQL extensions to add powerful capabilities</p>
  </div>
  <div style="margin-left: auto;">
      <input type="text" class="input" placeholder="Search extensions..." bind:value={searchQuery} style="width: 250px" />
  </div>
</div>

<div class="page-content">
  {#if loading}
    <div class="grid-cols-2 gap-4">
      {#each Array(6) as _}
        <div class="skeleton" style="height: 120px; border-radius: var(--radius-md);"></div>
      {/each}
    </div>
  {:else if error}
    <div class="alert alert-error">{error}</div>
  {:else if filteredExtensions.length === 0}
    <div class="empty-state">
      <div class="empty-icon">🔌</div>
      <h3>No Extensions Found</h3>
      <p>No extensions matching your search query.</p>
    </div>
  {:else}
    <div class="grid-cols-2 gap-4">
      {#each filteredExtensions as ext}
        <div class="card" style="display:flex;align-items:start;gap:16px; {ext.installed ? 'border-color: rgba(167, 139, 250, 0.3); background: linear-gradient(to bottom, rgba(167, 139, 250, 0.03), transparent);' : ''}">
          <div style="flex:1">
            <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
              <span style="font-weight:700;font-family:var(--font-mono);font-size:14px">{ext.name}</span>
              <span class="badge badge-neutral" style="font-size:10px">v{ext.version}</span>
              {#if ext.installed}<span class="badge badge-success">Enabled</span>{/if}
              <span class="badge" style="background: rgba(255,255,255,0.05); color: var(--text-secondary); font-size: 10px">{getExtCategory(ext.name)}</span>
            </div>
            <p style="font-size:13px;color:var(--text-secondary); line-height: 1.5">{ext.description || 'No description available for this extension.'}</p>
            {#if ext.installed && ext.schema}
               <div style="margin-top: 12px; font-size: 11px; color: var(--text-tertiary); font-family: var(--font-mono)">
                 Schema: {ext.schema}
               </div>
            {/if}
          </div>
          <div>
            {#if actionLoading === ext.name}
                <button class="btn btn-sm btn-secondary" disabled>Working...</button>
            {:else if ext.installed}
                <button class="btn btn-sm btn-danger" onclick={() => toggleExtension(ext)}>Disable</button>
            {:else}
                <button class="btn btn-sm btn-primary" onclick={() => toggleExtension(ext)}>Enable</button>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
