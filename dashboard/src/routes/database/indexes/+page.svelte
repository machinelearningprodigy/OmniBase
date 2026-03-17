<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getOmniBaseUrl, getHeaders, getErrorMessage } from '$lib/api'

  interface IndexMeta {
    name: string
    schema: string
    table: string
    columns: string
    index_type: string
    is_unique: boolean
    is_primary: boolean
    size: string
    scans: number
  }

  interface TableMeta {
    name: string
    schema: string
  }

  interface ColumnMeta {
    name: string
    type: string
  }

  let indexes = $state<IndexMeta[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)

  // Creation State
  let showCreateModal = $state(false)
  let isCreating = $state(false)
  
  let newName = $state('')
  let selectedTable = $state('')
  let newColumns = $state<string[]>([])
  let indexType = $state('BTREE')
  let isUnique = $state(false)

  // Options
  let tableList = $state<TableMeta[]>([])
  let columnList = $state<ColumnMeta[]>([])

  async function loadIndexes() {
    try {
      loading = true
      error = null
      const currentSchema = localStorage.getItem('omnibase.current_schema') || 'public'
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/indexes?schema=${currentSchema}`, { headers: getHeaders() })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to load indexes'))
      indexes = await resp.json()
    } catch (e) {
      error = e instanceof Error ? e.message : 'Unknown error'
    } finally {
      loading = false
    }
  }

  async function loadTables() {
    try {
      const currentSchema = localStorage.getItem('omnibase.current_schema') || 'public'
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/tables?schema=${currentSchema}`, { headers: getHeaders() })
      if (resp.ok) tableList = await resp.json()
    } catch (e) {
      console.error('Failed to load tables', e)
    }
  }

  async function loadColumns() {
    if (!selectedTable) {
      columnList = []
      newColumns = []
      return
    }
    try {
      const currentSchema = localStorage.getItem('omnibase.current_schema') || 'public'
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/columns?schema=${currentSchema}&table=${selectedTable}`, { headers: getHeaders(false) })
      if (resp.ok) columnList = await resp.json()
      // Filter out invalid columns that might have been selected previously
      newColumns = newColumns.filter(c => columnList.some(col => col.name === c))
    } catch (e) {
      console.error('Failed to load columns', e)
    }
  }

  async function createIndex() {
    if (!newName) return alert('Index name is required')
    if (!selectedTable) return alert('Target table is required')
    if (newColumns.length === 0) return alert('At least one column is required')

    isCreating = true
    try {
      const currentSchema = localStorage.getItem('omnibase.current_schema') || 'public'
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/indexes`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({
          name: newName,
          schema: currentSchema,
          table: selectedTable,
          columns: newColumns,
          index_type: indexType,
          is_unique: isUnique
        })
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to create index'))
      
      showCreateModal = false
      newName = ''
      selectedTable = ''
      newColumns = []
      indexType = 'BTREE'
      isUnique = false
      await loadIndexes()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Unknown error')
    } finally {
      isCreating = false
    }
  }

  async function deleteIndex(i: IndexMeta) {
    if (!confirm(`Are you sure you want to drop index "${i.name}"? This might impact query performance.`)) return
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/indexes?schema=${i.schema}&table=${i.table}&name=${i.name}`, {
        method: 'DELETE',
        headers: getHeaders(false)
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to delete index'))
      await loadIndexes()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Unknown error')
    }
  }

  function toggleColumnMapping(colName: string) {
    if (newColumns.includes(colName)) {
      newColumns = newColumns.filter(c => c !== colName)
    } else {
      newColumns = [...newColumns, colName]
    }
  }

  onMount(() => {
    loadIndexes()
    const handleSchemaChange = () => loadIndexes()
    window.addEventListener('omnibase:schema-change', handleSchemaChange)
    return () => window.removeEventListener('omnibase:schema-change', handleSchemaChange)
  })

  $effect(() => {
    // Only load options when modal is opened to save requests
    if (showCreateModal && tableList.length === 0) {
      loadTables()
    }
  })

  $effect(() => {
    if (selectedTable) {
        if (!newName) {
            // Auto-generate name based on table and columns if possible
            newName = `${selectedTable}_idx`
        }
        loadColumns()
    }
  })

  $effect(() => {
     if (newColumns.length > 0 && selectedTable && newName.endsWith('_idx')) {
         newName = `${selectedTable}_${newColumns.join('_')}_idx`
     }
  })
</script>

<svelte:head><title>Indexes — OmniBase</title></svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">Indexes</h1>
    <p class="page-subtitle">Manage database indexes to optimize query performance</p>
  </div>
  <button class="btn btn-primary" onclick={() => showCreateModal = true}>＋ New Index</button>
</div>

<div class="page-content">
  {#if loading}
    <div class="table-wrapper">
      <div class="skeleton" style="height: 200px; width: 100%; border-radius: var(--radius-md);"></div>
    </div>
  {:else if error}
    <div class="alert alert-error">{error}</div>
  {:else if indexes.length === 0}
      <div class="empty-state">
        <div class="empty-icon">↯</div>
        <h3>No Indexes</h3>
        <p>Indexes improve data retrieval speed. OmniBase automatically creates primary key indexes, but you can add more here.</p>
        <button class="btn btn-primary btn-sm" style="margin-top:16px" onclick={() => showCreateModal = true}>Create Index</button>
      </div>
  {:else}
    <div class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th>Index Name</th>
            <th>Table</th>
            <th>Columns</th>
            <th>Type</th>
            <th>Stats</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each indexes as idx}
            <tr>
              <td>
                <span class="cell-mono font-bold">{idx.name}</span>
                {#if idx.is_unique}<span class="badge badge-warning" style="margin-left: 8px">UNIQUE</span>{/if}
              </td>
              <td>{idx.schema}.{idx.table}</td>
              <td class="cell-mono">{idx.columns}</td>
              <td><span class="badge badge-info">{idx.index_type}</span></td>
              <td>
                <div style="display:flex;flex-direction:column;gap:4px;font-size:12px">
                  <span style="color:var(--text-secondary)">Size: {idx.size}</span>
                  <span style="color:var(--text-secondary)">Scans: {idx.scans.toLocaleString()}</span>
                </div>
              </td>
              <td style="text-align: right">
                <button class="btn btn-danger btn-sm" onclick={() => deleteIndex(idx)}>Drop</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

{#if showCreateModal}
  <div class="modal-backdrop" onclick={() => !isCreating && (showCreateModal = false)}>
    <div class="modal-content" style="max-width: 600px" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2 class="modal-title">Create Index</h2>
        <button class="modal-close" onclick={() => showCreateModal = false} disabled={isCreating}>×</button>
      </div>
      <div class="modal-body" style="display: flex; flex-direction: column; gap: 16px;">
        <div class="grid-cols-2 gap-4">
          <div class="form-group" style="margin: 0">
            <label class="form-label" for="index-table">Target Table</label>
            <select id="index-table" class="input" bind:value={selectedTable} disabled={isCreating}>
              <option value="">Select a table...</option>
              {#each tableList as tbl}
                <option value={tbl.name}>{tbl.name}</option>
              {/each}
            </select>
          </div>
          <div class="form-group" style="margin: 0">
            <label class="form-label" for="index-name">Index Name</label>
            <input id="index-name" class="input" type="text" placeholder="e.g. users_email_idx" bind:value={newName} disabled={isCreating} />
          </div>
        </div>

        <div class="grid-cols-2 gap-4">
          <div class="form-group" style="margin: 0">
            <label class="form-label" for="index-type">Index Type</label>
            <select id="index-type" class="input" bind:value={indexType} disabled={isCreating}>
              <option value="BTREE">B-Tree (Default)</option>
              <option value="HASH">Hash</option>
              <option value="GIN">GIN (Arrays/JSONB)</option>
              <option value="GIST">GiST (Geo/Trigram)</option>
              <option value="BRIN">BRIN (Sequential Data)</option>
              <option value="HNSW">HNSW (Vectors)</option>
            </select>
          </div>
          <div class="form-group" style="margin: 0; display:flex; align-items:center; height:100%; padding-top:24px;">
             <label style="display:flex;align-items:center;gap:8px;cursor:pointer">
                <input type="checkbox" bind:checked={isUnique} disabled={isCreating} />
                <span>Unique Index</span>
             </label>
          </div>
        </div>

        {#if selectedTable}
          <div class="form-group" style="margin: 0">
            <label class="form-label">Columns</label>
            {#if columnList.length === 0}
               <p class="form-hint">Loading columns...</p>
            {:else}
                <div style="display: flex; flex-wrap: wrap; gap: 8px; max-height: 200px; overflow-y: auto; padding: 12px; background: rgba(0,0,0,0.2); border-radius: var(--radius-md)">
                   {#each columnList as col}
                      <label style="display:flex;align-items:center;gap:6px;padding:4px 8px;background:var(--bg-card);border:1px solid var(--border-default);border-radius:4px;cursor:pointer">
                          <input type="checkbox" checked={newColumns.includes(col.name)} onchange={() => toggleColumnMapping(col.name)} disabled={isCreating} />
                          <span style="font-family:var(--font-mono);font-size:13px">{col.name}</span>
                          <span style="font-size:11px;color:var(--text-secondary)">{col.type}</span>
                      </label>
                   {/each}
                </div>
            {/if}
          </div>
        {/if}
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showCreateModal = false} disabled={isCreating}>Cancel</button>
        <button class="btn btn-primary" onclick={createIndex} disabled={isCreating || !newName || !selectedTable || newColumns.length === 0}>
          {isCreating ? 'Creating...' : 'Create Index'}
        </button>
      </div>
    </div>
  </div>
{/if}
