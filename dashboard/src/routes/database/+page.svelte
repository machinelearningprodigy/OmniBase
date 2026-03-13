<script lang="ts">
  import { onMount } from 'svelte'

  interface Table {
    name: string
    schema: string
    row_count?: number
    size?: string
    has_rls: boolean
  }

  let tables = $state<Table[]>([])
  let loading = $state(true)
  let selectedTable = $state<string | null>(null)
  let tableData = $state<Record<string, unknown>[]>([])
  let columns = $state<{ name: string; type: string }[]>([])
  let dataLoading = $state(false)
  let error = $state<string | null>(null)
  let searchQuery = $state('')
  let showCreateModal = $state(false)

  const OMNIBASE_URL = 'http://localhost:8000'

  async function loadTables() {
    try {
      loading = true
      // Fetch tables from postgres-meta service
      const resp = await fetch(`${OMNIBASE_URL}/meta/tables?schema=public`, {
        headers: getHeaders()
      })
      if (resp.ok) {
        const data = await resp.json()
        tables = data
      } else {
        // Mock data for development without backend
        tables = [
          { name: 'users', schema: 'public', row_count: 127, size: '48 kB', has_rls: true },
          { name: 'posts', schema: 'public', row_count: 532, size: '124 kB', has_rls: true },
          { name: 'comments', schema: 'public', row_count: 1802, size: '256 kB', has_rls: false },
          { name: 'tags', schema: 'public', row_count: 45, size: '12 kB', has_rls: false },
        ]
      }
    } catch {
      // Use mock data if backend is not running
      tables = [
        { name: 'users', schema: 'public', row_count: 127, size: '48 kB', has_rls: true },
        { name: 'posts', schema: 'public', row_count: 532, size: '124 kB', has_rls: true },
        { name: 'comments', schema: 'public', row_count: 1802, size: '256 kB', has_rls: false },
      ]
    } finally {
      loading = false
    }
  }

  async function selectTable(table: Table) {
    selectedTable = table.name
    dataLoading = true
    error = null
    try {
      const resp = await fetch(`${OMNIBASE_URL}/rest/v1/${table.name}?limit=50&select=*`, {
        headers: getHeaders()
      })
      if (resp.ok) {
        tableData = await resp.json()
        columns = tableData.length > 0
          ? Object.keys(tableData[0]).map(k => ({ name: k, type: guessType(tableData[0][k]) }))
          : []
      } else {
        error = `Failed to load data: ${resp.statusText}`
        tableData = []
      }
    } catch (e) {
      error = 'Could not connect to the API gateway. Is it running?'
      tableData = []
    } finally {
      dataLoading = false
    }
  }

  function guessType(val: unknown): string {
    if (val === null) return 'null'
    if (typeof val === 'boolean') return 'boolean'
    if (typeof val === 'number') return 'number'
    if (typeof val === 'object') return 'jsonb'
    if (typeof val === 'string' && /^\d{4}-\d{2}-\d{2}/.test(String(val))) return 'timestamp'
    return 'text'
  }

  function getHeaders() {
    const session = typeof localStorage !== 'undefined' ? localStorage.getItem('omnibase.session') : null
    const token = session ? JSON.parse(session).access_token : null
    return {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
  }

  function formatValue(val: unknown): string {
    if (val === null) return 'NULL'
    if (typeof val === 'object') return JSON.stringify(val)
    return String(val)
  }

  function truncate(str: string, n = 60) {
    return str.length > n ? str.slice(0, n) + '…' : str
  }

  let filteredTables = $derived(tables.filter(t =>
    t.name.toLowerCase().includes(searchQuery.toLowerCase())
  ))

  onMount(loadTables)
</script>

<svelte:head>
  <title>Table Editor — OmniBase</title>
  <meta name="description" content="Browse and edit your database tables" />
</svelte:head>

<div style="display: grid; grid-template-columns: 220px 1fr; height: calc(100vh - var(--topbar-height)); overflow: hidden;">
  <!-- Table List Sidebar -->
  <aside style="background: var(--bg-surface); border-right: 1px solid var(--border-subtle); overflow-y: auto; display: flex; flex-direction: column;">
    <div style="padding: 16px; border-bottom: 1px solid var(--border-subtle);">
      <div class="flex items-center justify-between" style="margin-bottom: 12px;">
        <span style="font-size: 12px; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em;">Tables</span>
        <button class="btn btn-primary btn-sm" on:click={() => showCreateModal = true} id="create-table-btn">
          + New
        </button>
      </div>
      <input
        class="input"
        type="text"
        placeholder="Filter tables..."
        bind:value={searchQuery}
        id="table-search"
      />
    </div>

    <div style="overflow-y: auto; flex: 1; padding: 8px;">
      {#if loading}
        {#each Array(5) as _}
          <div class="skeleton" style="height: 36px; margin-bottom: 4px; border-radius: var(--radius-sm);"></div>
        {/each}
      {:else if filteredTables.length === 0}
        <div class="empty-state" style="padding: 30px 10px;">
          <p>No tables found</p>
        </div>
      {:else}
        {#each filteredTables as table}
          <button
            class="nav-item {selectedTable === table.name ? 'active' : ''}"
            on:click={() => selectTable(table)}
            id="table-{table.name}"
          >
            <span style="font-size: 12px; font-family: var(--font-mono);">⊞</span>
            <span style="flex: 1; text-align: left;">{table.name}</span>
            <div class="flex items-center gap-1">
              {#if table.has_rls}
                <span class="badge badge-success" style="font-size: 9px; padding: 1px 4px;">RLS</span>
              {/if}
              {#if table.row_count !== undefined}
                <span style="font-size: 10px; color: var(--text-muted);">{table.row_count}</span>
              {/if}
            </div>
          </button>
        {/each}
      {/if}
    </div>
  </aside>

  <!-- Table Data View -->
  <div style="display: flex; flex-direction: column; overflow: hidden;">
    {#if !selectedTable}
      <div class="empty-state" style="height: 100%;">
        <p style="font-size: 48px;">◫</p>
        <h3>Select a table to explore</h3>
        <p>Choose a table from the sidebar to view and edit your data</p>
        <button class="btn btn-primary" style="margin-top: 8px;" on:click={loadTables}>
          {loading ? 'Loading...' : 'Refresh Tables'}
        </button>
      </div>
    {:else}
      <!-- Table Header -->
      <div class="page-header" style="padding: 16px 24px;">
        <div>
          <h1 class="page-title" style="font-size: 16px; font-family: var(--font-mono);">
            public.{selectedTable}
          </h1>
          <p class="page-subtitle">{tableData.length} rows loaded · Live via PostgREST REST API</p>
        </div>
        <div class="flex gap-2">
          <button class="btn btn-secondary btn-sm">Insert Row</button>
          <a href="/database/editor?table={selectedTable}" class="btn btn-secondary btn-sm">Open in SQL Editor</a>
        </div>
      </div>

      <!-- Table Content -->
      <div style="flex: 1; overflow: auto; padding: 0 24px 24px;">
        {#if dataLoading}
          <div class="card" style="margin-top: 16px;">
            {#each Array(8) as _}
              <div class="skeleton" style="height: 44px; margin-bottom: 4px;"></div>
            {/each}
          </div>
        {:else if error}
          <div class="card" style="margin-top: 16px; border-color: rgba(255,82,82,0.3); background: rgba(255,82,82,0.05);">
            <div style="color: var(--status-error); font-size: 13px;">
              <strong>⚠ Error loading data</strong><br>
              {error}
            </div>
            <div style="margin-top: 12px; font-size: 12px; color: var(--text-muted);">
              Make sure your OmniBase stack is running: <code>docker compose up -d</code>
            </div>
          </div>
        {:else if columns.length === 0}
          <div class="empty-state" style="margin-top: 20px;">
            <h3>Table is empty</h3>
            <p>Insert some rows to see data here</p>
          </div>
        {:else}
          <div class="table-wrapper animate-fade-in" style="margin-top: 16px;">
            <table>
              <thead>
                <tr>
                  {#each columns as col}
                    <th>
                      <span>{col.name}</span>
                      <span style="color: var(--brand-primary); margin-left: 6px; font-size: 9px; font-weight: 400;">{col.type}</span>
                    </th>
                  {/each}
                </tr>
              </thead>
              <tbody>
                {#each tableData as row}
                  <tr>
                    {#each columns as col}
                      <td class="{col.type === 'text' || col.type === 'timestamp' ? '' : 'cell-mono'}">
                        {#if row[col.name] === null}
                          <span style="color: var(--text-muted); font-style: italic; font-size: 11px;">NULL</span>
                        {:else if col.type === 'boolean'}
                          <span class="badge {row[col.name] ? 'badge-success' : 'badge-neutral'}">{String(row[col.name])}</span>
                        {:else}
                          <span class="truncate" style="display: block; max-width: 250px;" title={formatValue(row[col.name])}>
                            {truncate(formatValue(row[col.name]))}
                          </span>
                        {/if}
                      </td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
