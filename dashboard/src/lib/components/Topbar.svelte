<script lang="ts">
  import { onMount } from 'svelte'
  import { authStore } from '$lib/stores/auth'

  let schemas = $state<string[]>([])
  let currentSchema = $state('public')
  let loading = $state(false)

  const OMNIBASE_URL = 'http://localhost:8000'

  function getHeaders() {
    const session = typeof localStorage !== 'undefined' ? localStorage.getItem('omnibase.session') : null
    const token = session ? JSON.parse(session).access_token : null
    return {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
  }

  async function loadSchemas() {
    try {
      loading = true
      const resp = await fetch(`${OMNIBASE_URL}/pg/schemas`, { headers: getHeaders() })
      if (resp.ok) {
        schemas = await resp.json()
      }
    } catch {} finally {
      loading = false
    }
  }

  function switchSchema(name: string) {
    currentSchema = name
    // Notify app - in a real app would use a store
    localStorage.setItem('omnibase.current_schema', name)
    window.dispatchEvent(new CustomEvent('omnibase:schema-change', { detail: name }))
  }

  onMount(() => {
    loadSchemas()
    currentSchema = localStorage.getItem('omnibase.current_schema') || 'public'
  })
</script>

<header class="topbar">
  <!-- Left: breadcrumb/project switcher -->
  <div class="flex items-center gap-2" style="flex: 1;">
    <div style="font-size: 12px; color: var(--text-muted); display: flex; align-items: center; gap: 4px;">
      <span style="color: var(--text-secondary); font-weight: 500;">Project</span>
      <select 
        bind:value={currentSchema} 
        onchange={() => switchSchema(currentSchema)}
        style="background: none; border: none; color: var(--brand-primary); font-weight: 600; font-size: 12px; cursor: pointer; outline: none;"
      >
        {#each schemas as s}
          <option value={s}>{s}</option>
        {/each}
        {#if schemas.length === 0}
          <option value="public">public</option>
        {/if}
      </select>
    </div>
    <span style="color: var(--border-default);">/</span>
    <div style="font-size: 12px; color: var(--text-secondary);">omnibase-cloud</div>
  </div>

  <!-- Center: search -->
  <div style="flex: 1; max-width: 400px;">
    <div style="background: rgba(255,255,255,0.04); border: 1px solid var(--border-subtle); border-radius: var(--radius-md); padding: 6px 12px; display: flex; align-items: center; gap: 8px; color: var(--text-muted); font-size: 12px;">
      <span>⌕</span>
      <span>Search or jump to...</span>
      <kbd style="margin-left: auto; background: rgba(255,255,255,0.06); padding: 1px 5px; border-radius: 3px; font-size: 10px; font-family: var(--font-mono);">⌘K</kbd>
    </div>
  </div>

  <!-- Right: status + user -->
  <div class="flex items-center gap-2" style="flex: 1; justify-content: flex-end;">
    <div class="flex items-center gap-2" style="font-size: 11px; color: var(--text-secondary); background: rgba(0,230,118,0.08); border: 1px solid rgba(0,230,118,0.15); padding: 4px 10px; border-radius: 20px;">
      <span class="status-dot online"></span>
      All systems operational
    </div>

    <a href="https://docs.omnibase.dev" target="_blank" class="btn btn-secondary btn-sm" style="text-decoration: none;">
      Docs
    </a>

    <div style="width: 30px; height: 30px; border-radius: 50%; background: var(--gradient-brand); display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; cursor: pointer;" data-tooltip="admin@omnibase.dev">
      {$authStore.user?.email?.[0].toUpperCase() || 'A'}
    </div>
  </div>
</header>
