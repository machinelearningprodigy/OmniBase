<script lang="ts">
  import { onMount } from 'svelte'
  import { authStore } from '$lib/stores/auth'
  import { apiFetch, getHeaders, getOmniBaseUrl } from '$lib/api'

  let schemas = $state<string[]>([])
  let currentSchema = $state('public')
  let loading = $state(false)
  let showUserMenu = $state(false)

  async function loadSchemas() {
    try {
      loading = true
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/schemas`, { headers: getHeaders() })
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

  function handleWindowClick(event: MouseEvent) {
    const target = event.target as HTMLElement | null
    if (!target?.closest('[data-user-menu]')) {
      showUserMenu = false
    }
  }

  onMount(() => {
    loadSchemas()
    currentSchema = localStorage.getItem('omnibase.current_schema') || 'public'
    window.addEventListener('click', handleWindowClick)
    return () => window.removeEventListener('click', handleWindowClick)
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

    {#if $authStore.user}
      <div class="user-menu" data-user-menu>
        <button
          class="user-menu-trigger"
          type="button"
          title={$authStore.user.email}
          onclick={() => showUserMenu = !showUserMenu}
        >
          <div class="user-avatar">
            {$authStore.user.email?.[0].toUpperCase() || 'A'}
          </div>
        </button>

        {#if showUserMenu}
          <div class="user-menu-popover">
            <div class="user-menu-email">{$authStore.user.email}</div>
            <a href="/settings" class="user-menu-item" onclick={() => showUserMenu = false}>Settings</a>
            <button
              class="user-menu-item"
              type="button"
              onclick={() => {
                showUserMenu = false
                authStore.signOut()
              }}
            >
              Sign out
            </button>
          </div>
        {/if}
      </div>
    {:else}
      <a href="/auth/login" class="btn btn-primary btn-sm" style="text-decoration: none;">Sign in</a>
    {/if}
  </div>
</header>

<style>
  .user-menu {
    position: relative;
  }

  .user-menu-trigger {
    padding: 0;
    border: none;
    background: transparent;
    cursor: pointer;
  }

  .user-avatar {
    width: 34px;
    height: 34px;
    border-radius: 50%;
    background: var(--gradient-brand);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 700;
    color: #fff;
  }

  .user-menu-popover {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    min-width: 200px;
    padding: 8px;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    background: var(--bg-surface);
    box-shadow: var(--shadow-lg);
    z-index: 50;
  }

  .user-menu-email {
    padding: 8px 10px 10px;
    font-size: 12px;
    color: var(--text-muted);
    border-bottom: 1px solid var(--border-subtle);
    margin-bottom: 6px;
    word-break: break-word;
  }

  .user-menu-item {
    width: 100%;
    display: block;
    padding: 10px;
    border: none;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--text-primary);
    font-size: 13px;
    text-align: left;
    text-decoration: none;
    cursor: pointer;
  }

  .user-menu-item:hover {
    background: var(--bg-elevated);
  }
</style>
