<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getHeaders, getOmniBaseUrl, getErrorMessage } from '$lib/api'
  import { Trash2, Zap, Play, Clock, Info, ShieldAlert } from 'lucide-svelte'

  interface Trigger {
    name: string
    schema: string
    table: string
    function: string
    events: string
    timing: string
    orientation: string
    enabled: string
    definition: string
  }

  interface DBFunction {
    name: string
    schema: string
  }

  interface Table {
    name: string
    schema: string
  }

  let triggers = $state<Trigger[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)
  let currentSchema = $state('public')

  let showCreateModal = $state(false)
  let createLoading = $state(false)
  let createError = $state<string | null>(null)

  let availableFunctions = $state<DBFunction[]>([])
  let availableTables = $state<Table[]>([])

  let newTrigger = $state({
    name: '',
    table: '',
    function: '',
    timing: 'BEFORE',
    events: [] as string[],
    orientation: 'ROW'
  })

  async function loadTriggers() {
    try {
      loading = true
      error = null
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/triggers?schema=${currentSchema}`)
      if (!resp.ok) {
        error = await getErrorMessage(resp, 'Failed to load triggers')
        return
      }
      triggers = await resp.json()
    } catch (e) {
      error = 'Failed to connect to backend'
    } finally {
      loading = false
    }
  }

  async function loadMetadata() {
    try {
      const [funcResp, tableResp] = await Promise.all([
        apiFetch(`${getOmniBaseUrl()}/pg/functions?schema=${currentSchema}`),
        apiFetch(`${getOmniBaseUrl()}/pg/tables?schema=${currentSchema}`)
      ])
      
      if (funcResp.ok) availableFunctions = await funcResp.json()
      if (tableResp.ok) availableTables = await tableResp.json()
    } catch {}
  }

  async function createTrigger() {
    if (!newTrigger.name || !newTrigger.table || !newTrigger.function || newTrigger.events.length === 0) {
      createError = 'Please fill in all required fields'
      return
    }

    try {
      createLoading = true
      createError = null
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/triggers`, {
        method: 'POST',
        headers: getHeaders(true),
        body: JSON.stringify({
          ...newTrigger,
          schema: currentSchema
        })
      })

      if (!resp.ok) {
        createError = await getErrorMessage(resp, 'Failed to create trigger')
        return
      }

      showCreateModal = false
      await loadTriggers()
      resetForm()
    } catch (e) {
      createError = 'Connection error'
    } finally {
      createLoading = false
    }
  }

  async function deleteTrigger(t: Trigger) {
    if (!confirm(`Are you sure you want to delete trigger "${t.name}"?`)) return

    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/triggers?schema=${t.schema}&table=${t.table}&name=${t.name}`, {
        method: 'DELETE'
      })
      if (!resp.ok) {
        alert(await getErrorMessage(resp, 'Failed to delete trigger'))
        return
      }
      await loadTriggers()
    } catch {
      alert('Connection error')
    }
  }

  function toggleEvent(event: string) {
    if (newTrigger.events.includes(event)) {
      newTrigger.events = newTrigger.events.filter(e => e !== event)
    } else {
      newTrigger.events = [...newTrigger.events, event]
    }
  }

  function resetForm() {
    newTrigger = {
      name: '',
      table: '',
      function: '',
      timing: 'BEFORE',
      events: [],
      orientation: 'ROW'
    }
    createError = null
  }

  onMount(() => {
    loadTriggers()
    loadMetadata()
  })

  // Watch schema change
  $effect(() => {
    const handleSchema = (e: any) => {
      currentSchema = e.detail
      loadTriggers()
      loadMetadata()
    }
    window.addEventListener('omnibase:schema-change', handleSchema as EventListener)
    return () => window.removeEventListener('omnibase:schema-change', handleSchema as EventListener)
  })
</script>

<svelte:head><title>Triggers — OmniBase</title></svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Triggers</h1>
      <p class="page-subtitle">Automate database actions on data changes</p>
    </div>
    <button class="btn btn-primary" onclick={() => showCreateModal = true}>
      <Zap size={16} /> New Trigger
    </button>
  </div>

  <div class="page-content">
    {#if error}
      <div class="alert alert-error">
        <ShieldAlert size={20} />
        <span>{error}</span>
      </div>
    {/if}

    {#if loading}
      <div class="loading-grid">
        {#each Array(3) as _}
          <div class="skeleton card" style="height: 120px"></div>
        {/each}
      </div>
    {:else if triggers.length === 0}
      <div class="card empty-state-card">
        <div class="icon-circle">
          <Zap size={32} class="icon-dim" />
        </div>
        <h3>No triggers found</h3>
        <p>Define triggers to run functions when rows are inserted, updated, or deleted.</p>
        <button class="btn btn-secondary" onclick={() => showCreateModal = true}>Create your first trigger</button>
      </div>
    {:else}
      <div class="trigger-grid">
        {#each triggers as t}
          <div class="card trigger-card">
            <div class="trigger-card-header">
              <div class="trigger-info">
                <div class="trigger-name-row">
                  <h3 class="trigger-name">{t.name}</h3>
                  <span class="badge badge-outline-success">{t.enabled}</span>
                </div>
                <div class="trigger-meta">
                  <span class="meta-item"><Info size={12} /> ON {t.table}</span>
                  <span class="meta-item"><Clock size={12} /> {t.timing}</span>
                </div>
              </div>
              <button class="btn-icon btn-danger-soft" onclick={() => deleteTrigger(t)} title="Delete Trigger">
                <Trash2 size={16} />
              </button>
            </div>
            
            <div class="trigger-card-body">
              <div class="event-capsules">
                {#each t.events.split(' OR ') as event}
                  <span class="capsule capsule-primary">{event}</span>
                {/each}
              </div>
              <div class="trigger-action">
                <Play size={14} class="text-dim" />
                <span class="func-name">{t.function}()</span>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

{#if showCreateModal}
  <div class="modal-overlay">
    <div class="modal card">
      <div class="modal-header">
        <h2>Create Database Trigger</h2>
        <button class="close-btn" onclick={() => showCreateModal = false}>&times;</button>
      </div>
      
      <div class="modal-body">
        {#if createError}
          <div class="alert alert-error" style="margin-bottom: 16px;">{createError}</div>
        {/if}

        <div class="form-group">
          <label>Trigger Name</label>
          <input type="text" class="input" placeholder="e.g. notify_on_new_user" bind:value={newTrigger.name} />
          <p class="input-hint">Unique name to identify this trigger</p>
        </div>

        <div class="form-grid">
          <div class="form-group">
            <label>Table</label>
            <select class="input" bind:value={newTrigger.table}>
              <option value="">Select a table...</option>
              {#each availableTables as table}
                <option value={table.name}>{table.name}</option>
              {/each}
            </select>
          </div>

          <div class="form-group">
            <label>Function to Run</label>
            <select class="input" bind:value={newTrigger.function}>
              <option value="">Select a function...</option>
              {#each availableFunctions as func}
                <option value={func.name}>{func.name}</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="form-group">
          <label>When to fire</label>
          <div class="radio-group">
            <button class="radio-item {newTrigger.timing === 'BEFORE' ? 'active' : ''}" onclick={() => newTrigger.timing = 'BEFORE'}>BEFORE</button>
            <button class="radio-item {newTrigger.timing === 'AFTER' ? 'active' : ''}" onclick={() => newTrigger.timing = 'AFTER'}>AFTER</button>
          </div>
        </div>

        <div class="form-group">
          <label>Events</label>
          <div class="checkbox-group">
            {#each ['INSERT', 'UPDATE', 'DELETE'] as event}
              <button class="check-item {newTrigger.events.includes(event) ? 'active' : ''}" onclick={() => toggleEvent(event)}>
                {event}
              </button>
            {/each}
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-ghost" onclick={() => showCreateModal = false}>Cancel</button>
        <button class="btn btn-primary" onclick={createTrigger} disabled={createLoading}>
          {createLoading ? 'Creating...' : 'Create Trigger'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page-container {
    padding: 24px;
    max-width: 1200px;
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 32px;
  }

  .page-title {
    font-size: 24px;
    font-weight: 700;
    margin: 0 0 4px 0;
  }

  .page-subtitle {
    color: var(--text-dim);
    margin: 0;
  }

  .trigger-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
    gap: 20px;
  }

  .trigger-card {
    display: flex;
    flex-direction: column;
    padding: 0;
    overflow: hidden;
    transition: transform 0.2s, box-shadow 0.2s;
    border: 1px solid var(--border-color);
  }

  .trigger-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.12);
  }

  .trigger-card-header {
    padding: 16px;
    display: flex;
    justify-content: space-between;
    background: rgba(255,255,255,0.02);
    border-bottom: 1px solid var(--border-color);
  }

  .trigger-name-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .trigger-name {
    font-size: 16px;
    font-weight: 600;
    margin: 0;
  }

  .trigger-meta {
    display: flex;
    gap: 12px;
    font-size: 12px;
    color: var(--text-dim);
  }

  .meta-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .trigger-card-body {
    padding: 16px;
  }

  .event-capsules {
    display: flex;
    gap: 6px;
    margin-bottom: 12px;
  }

  .capsule {
    padding: 2px 8px;
    border-radius: 99px;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.02em;
  }

  .capsule-primary {
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary-color);
  }

  .trigger-action {
    display: flex;
    align-items: center;
    gap: 8px;
    background: rgba(0,0,0,0.2);
    padding: 8px 12px;
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: 13px;
  }

  .func-name {
    color: #ffd76d;
  }

  /* Modal Styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0,0,0,0.7);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 20px;
  }

  .modal {
    width: 100%;
    max-width: 500px;
    padding: 0;
    background: #1a1b1e;
  }

  .modal-header {
    padding: 20px;
    border-bottom: 1px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .modal-header h2 { margin: 0; font-size: 18px; }

  .modal-body { padding: 20px; }

  .modal-footer {
    padding: 16px 20px;
    background: rgba(0,0,0,0.1);
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    border-top: 1px solid var(--border-color);
  }

  .form-group { margin-bottom: 20px; }
  .form-group label { display: block; margin-bottom: 8px; font-weight: 500; font-size: 14px; }
  .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

  .radio-group, .checkbox-group {
    display: flex;
    gap: 8px;
  }

  .radio-item, .check-item {
    flex: 1;
    padding: 8px;
    background: rgba(255,255,255,0.05);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .radio-item:hover, .check-item:hover { background: rgba(255,255,255,0.1); }
  .radio-item.active, .check-item.active {
    background: rgba(var(--primary-rgb), 0.15);
    border-color: var(--primary-color);
    color: var(--primary-color);
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-dim);
    font-size: 24px;
    cursor: pointer;
  }

  .empty-state-card {
    padding: 64px 32px;
    text-align: center;
    background: linear-gradient(to bottom, rgba(255,255,255,0.02), transparent);
  }

  .icon-circle {
    width: 80px;
    height: 80px;
    background: rgba(255,255,255,0.03);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 24px auto;
  }

  .icon-dim { opacity: 0.2; }
</style>
