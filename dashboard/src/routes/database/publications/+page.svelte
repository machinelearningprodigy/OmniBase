<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getOmniBaseUrl, getHeaders, getErrorMessage } from '$lib/api'

  interface Publication {
    pubname: string
    alltables: boolean
    insert: boolean
    update: boolean
    delete: boolean
    truncate: boolean
    tables: string[]
  }

  interface TableMeta {
    name: string
    schema: string
  }

  let pubs = $state<Publication[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)

  let tableList = $state<TableMeta[]>([])
  
  let showCreateModal = $state(false)
  let isCreating = $state(false)
  let newName = $state('')
  let isAllTables = $state(true)
  let selectedTables = $state<string[]>([])

  async function loadData() {
    try {
      loading = true
      error = null
      
      const pubResp = await apiFetch(`${getOmniBaseUrl()}/pg/publications`, { headers: getHeaders() })
      if (!pubResp.ok) throw new Error(await getErrorMessage(pubResp, 'Failed to load publications'))
      pubs = await pubResp.json()

      const tblResp = await apiFetch(`${getOmniBaseUrl()}/pg/tables`, { headers: getHeaders() })
      if (tblResp.ok) tableList = await tblResp.json()

    } catch (e) {
      error = e instanceof Error ? e.message : 'Unknown error'
    } finally {
      loading = false
    }
  }

  async function createPub() {
    if (!newName) return alert('Name is required')
    if (!isAllTables && selectedTables.length === 0) return alert('Select at least one table, or choose All Tables')

    isCreating = true
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/publications`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ name: newName, tables: isAllTables ? [] : selectedTables })
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to create publication'))
      
      showCreateModal = false
      newName = ''
      isAllTables = true
      selectedTables = []
      await loadData()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Unknown error')
    } finally {
      isCreating = false
    }
  }

  async function deletePub(name: string) {
    if (!confirm(`Are you sure you want to drop publication "${name}"? Subscribers will stop receiving updates.`)) return
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/publications?name=${name}`, {
        method: 'DELETE',
        headers: getHeaders(false)
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to delete publication'))
      await loadData()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Unknown error')
    }
  }

  function toggleTable(t: string) {
      if (selectedTables.includes(t)) {
          selectedTables = selectedTables.filter(x => x !== t)
      } else {
          selectedTables = [...selectedTables, t]
      }
  }

  onMount(() => loadData())
</script>

<svelte:head><title>Publications — OmniBase</title></svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">Publications</h1>
    <p class="page-subtitle">Postgres logical replication publications — control what gets streamed to the realtime engine</p>
  </div>
  <button class="btn btn-primary" onclick={() => showCreateModal = true}>＋ New Publication</button>
</div>

<div class="page-content">
  {#if loading}
    <div class="table-wrapper"><div class="skeleton" style="height: 200px; width: 100%; border-radius: var(--radius-md);"></div></div>
  {:else if error}
    <div class="alert alert-error">{error}</div>
  {:else if pubs.length === 0}
    <div class="empty-state">
      <div class="empty-icon">📡</div>
      <h3>No Publications Found</h3>
      <p>Publications act as broadcasters for Realtime subscriptions. OmniBase uses `omnibase_realtime` by default, but you can create custom ones too.</p>
      <button class="btn btn-primary btn-sm mt-4" onclick={() => showCreateModal = true}>Create Publication</button>
    </div>
  {:else}
    <div class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th>Publication</th>
            <th>Tables</th>
            <th>Operations</th>
            <th style="text-align: right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each pubs as p}
            <tr>
              <td>
                <span class="cell-mono font-bold">{p.pubname}</span>
                {#if p.pubname === 'omnibase_realtime'}<span class="badge badge-info ml-2">System</span>{/if}
              </td>
              <td>
                {#if p.alltables}
                  <span class="badge badge-success">All Tables</span>
                {:else if p.tables.length > 0}
                  <div style="display:flex;flex-wrap:wrap;gap:4px">
                    {#each p.tables.slice(0, 3) as t}
                      <span class="badge badge-neutral cell-mono">{t}</span>
                    {/each}
                    {#if p.tables.length > 3}
                       <span class="badge badge-neutral">+{p.tables.length - 3} more</span>
                    {/if}
                  </div>
                {:else}
                  <span class="badge" style="color:var(--text-secondary)">No Tables</span>
                {/if}
              </td>
              <td>
                <div style="display:flex;gap:4px">
                  {#if p.insert}<span class="badge badge-neutral" style="font-size:10px">INS</span>{/if}
                  {#if p.update}<span class="badge badge-neutral" style="font-size:10px">UPD</span>{/if}
                  {#if p.delete}<span class="badge badge-neutral" style="font-size:10px">DEL</span>{/if}
                  {#if p.truncate}<span class="badge badge-neutral" style="font-size:10px">TRUN</span>{/if}
                </div>
              </td>
              <td style="text-align: right">
                <button class="btn btn-danger btn-sm" onclick={() => deletePub(p.pubname)}>Drop</button>
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
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2 class="modal-title">Create Publication</h2>
        <button class="modal-close" onclick={() => showCreateModal = false} disabled={isCreating}>×</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label class="form-label">Publication Name</label>
          <input class="input" type="text" placeholder="e.g. real_time_events" bind:value={newName} disabled={isCreating} />
        </div>
        <div class="form-group">
          <label class="form-label">Target Tables</label>
          <div style="display:flex;flex-direction:column;gap:12px">
            <label style="display:flex;align-items:center;gap:8px">
              <input type="radio" value={true} bind:group={isAllTables} disabled={isCreating} />
              <span>All Tables (Current and Future)</span>
            </label>
            <label style="display:flex;align-items:center;gap:8px">
              <input type="radio" value={false} bind:group={isAllTables} disabled={isCreating} />
              <span>Specific Tables</span>
            </label>
          </div>
        </div>
        
        {#if !isAllTables}
          <div class="form-group">
            <div style="display:flex;flex-wrap:wrap;gap:8px;max-height:150px;overflow-y:auto;padding:12px;background:var(--bg-document);border-radius:var(--radius-md);border:1px solid var(--border-default)">
              {#each tableList as tbl}
                <label style="display:flex;align-items:center;gap:6px;padding:4px 8px;background:var(--bg-card);border:1px solid var(--border-default);border-radius:4px;cursor:pointer">
                  <input type="checkbox" checked={selectedTables.includes(tbl.name)} onchange={() => toggleTable(tbl.name)} disabled={isCreating}/>
                  <span class="cell-mono text-sm">{tbl.name}</span>
                </label>
              {/each}
            </div>
          </div>
        {/if}
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showCreateModal = false} disabled={isCreating}>Cancel</button>
        <button class="btn btn-primary" onclick={createPub} disabled={isCreating || !newName}>
          {isCreating ? 'Creating...' : 'Create Publication'}
        </button>
      </div>
    </div>
  </div>
{/if}
