<script lang="ts">
  import { onMount } from 'svelte'
  import { getHeaders, getOmniBaseUrl } from '$lib/api'

  interface DBFunction {
    name: string
    schema: string
    return_type: string
    language: string
    definition: string
  }

  let functions = $state<DBFunction[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)
  let currentSchema = $state('public')
  let selectedFn = $state<DBFunction | null>(null)

  async function loadFunctions() {
    try {
      loading = true
      error = null
      const resp = await fetch(`${getOmniBaseUrl()}/pg/functions?schema=${currentSchema}`, {
        headers: getHeaders()
      })
      if (resp.ok) {
        functions = await resp.json()
      } else {
        error = 'Failed to load functions'
      }
    } catch {
      error = 'Connection error'
    } finally {
      loading = false
    }
  }

  onMount(() => {
    currentSchema = localStorage.getItem('omnibase.current_schema') || 'public'
    loadFunctions()

    const handleSchemaChange = (e: any) => {
      currentSchema = e.detail
      selectedFn = null
      loadFunctions()
    }

    window.addEventListener('omnibase:schema-change', handleSchemaChange)
    return () => window.removeEventListener('omnibase:schema-change', handleSchemaChange)
  })
</script>

<svelte:head>
  <title>Database Functions — OmniBase</title>
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">Database Functions</h1>
    <p class="page-subtitle">Custom SQL and PL/pgSQL procedures in {currentSchema}</p>
  </div>
  <button class="btn btn-primary btn-sm" onclick={() => window.location.href = '/database/editor'}>
    Create Function
  </button>
</div>

<div style="display: grid; grid-template-columns: 300px 1fr; height: calc(100vh - 160px); padding: 0 24px 24px; gap: 24px; overflow: hidden;">
  <!-- List -->
  <div class="card" style="display: flex; flex-direction: column; overflow: hidden; padding: 0;">
    <div style="padding: 16px; border-bottom: 1px solid var(--border-subtle);">
      <input type="text" class="input input-sm" placeholder="Filter functions..." />
    </div>
    <div style="flex: 1; overflow-y: auto;">
      {#if loading}
        <div style="padding: 20px;"><div class="skeleton" style="height: 100px;"></div></div>
      {:else if functions.length === 0}
        <div class="empty-state" style="padding: 40px 10px;">
          <p>No functions in {currentSchema}</p>
        </div>
      {:else}
        {#each functions as fn}
          <button 
            class="nav-item {selectedFn?.name === fn.name ? 'active' : ''}"
            onclick={() => selectedFn = fn}
            style="width: 100%; border: none; background: none; border-bottom: 1px solid var(--border-subtle); padding: 12px 16px; text-align: left; cursor: pointer; display: flex; flex-direction: column; gap: 4px;"
          >
            <span style="font-family: var(--font-mono); font-size: 13px; font-weight: 500;">{fn.name}</span>
            <span style="font-size: 11px; color: var(--text-muted);">{fn.return_type} ({fn.language})</span>
          </button>
        {/each}
      {/if}
    </div>
  </div>

  <!-- Detail -->
  <div class="card" style="overflow: hidden; display: flex; flex-direction: column;">
    {#if !selectedFn}
      <div class="empty-state" style="flex: 1;">
        <p>Select a function to view its definition</p>
      </div>
    {:else}
      <div style="padding: 16px; border-bottom: 1px solid var(--border-subtle); display: flex; justify-content: space-between; align-items: center;">
        <h3 style="font-family: var(--font-mono); margin: 0;">{selectedFn.name}</h3>
        <button class="btn btn-secondary btn-sm" onclick={() => window.location.href = `/database/editor?query=${encodeURIComponent(selectedFn?.definition || '')}`}>Edit in SQL Editor</button>
      </div>
      <div style="flex: 1; overflow: auto; background: #0d1117; padding: 20px;">
        <pre style="color: #e6edf3; font-family: var(--font-mono); font-size: 13px; margin: 0; white-space: pre-wrap;">{selectedFn.definition}</pre>
      </div>
    {/if}
  </div>
</div>

<style>
  .nav-item:hover { background: rgba(255,255,255,0.04); }
  .nav-item.active { background: rgba(var(--brand-primary-rgb), 0.1); border-left: 2px solid var(--brand-primary); }
</style>
