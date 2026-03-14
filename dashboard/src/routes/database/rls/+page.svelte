<script lang="ts">
  import { onMount } from 'svelte'
  import { getErrorMessage, getHeaders, getOmniBaseUrl, apiFetch } from '$lib/api'

  interface Policy {
    name: string
    schema: string
    table: string
    action: string
    roles: string[]
    qualifier: string
  }

  interface Table {
    name: string
    has_rls: boolean
  }

  let policies = $state<Policy[]>([])
  let tables = $state<Table[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)

  async function loadData() {
    try {
      loading = true
      error = null
      
      const [tableResp, policyResp] = await Promise.all([
        apiFetch(`${getOmniBaseUrl()}/pg/tables?schema=public`, { headers: getHeaders() }),
        apiFetch(`${getOmniBaseUrl()}/pg/policies?schema=public`, { headers: getHeaders() })
      ])

      if (tableResp.ok && policyResp.ok) {
        tables = await tableResp.json()
        policies = await policyResp.json()
      } else {
        error = 'Failed to load policy data'
      }
    } catch (e) {
      error = 'Connection error'
    } finally {
      loading = false
    }
  }

  async function toggleRLS(table: string, current: boolean) {
    const sql = `ALTER TABLE public."${table}" ${current ? 'DISABLE' : 'ENABLE'} ROW LEVEL SECURITY;`
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/query`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ query: sql })
      })
      if (resp.ok) {
        await loadData()
      } else {
        alert(await getErrorMessage(resp, 'Failed to toggle RLS'))
      }
    } catch (e) {
      alert('Error toggling RLS')
    }
  }

  onMount(loadData)
</script>

<svelte:head>
  <title>RLS Policies — OmniBase</title>
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">Row Level Security (RLS)</h1>
    <p class="page-subtitle">Control access to your data at the row level</p>
  </div>
  <button class="btn btn-primary btn-sm" onclick={() => window.location.href = '/database/editor?query=CREATE POLICY...'}>
    Create Policy
  </button>
</div>

<div class="page-content">
  {#if loading}
    <div class="grid grid-cols-1 gap-4">
      {#each Array(3) as _}
        <div class="skeleton" style="height: 100px; border-radius: 12px;"></div>
      {/each}
    </div>
  {:else if error}
    <div class="card error-card">
      <p>{error}</p>
      <button class="btn btn-secondary btn-sm" onclick={loadData}>Retry</button>
    </div>
  {:else}
    <div class="grid gap-6">
      <!-- Tables State -->
      <div class="card">
        <h2 style="font-size: 15px; margin-bottom: 16px;">Tables Summary</h2>
        <div class="table-wrapper">
          <table>
            <thead>
              <tr>
                <th>Table</th>
                <th>RLS Status</th>
                <th style="text-align: right;">Action</th>
              </tr>
            </thead>
            <tbody>
              {#each tables as table}
                <tr>
                  <td style="font-family: var(--font-mono); font-size: 13px;">{table.name}</td>
                  <td>
                    <span class="badge {table.has_rls ? 'badge-success' : 'badge-neutral'}">
                      {table.has_rls ? 'Enabled' : 'Disabled'}
                    </span>
                  </td>
                  <td style="text-align: right;">
                    <button 
                      class="btn btn-secondary btn-sm" 
                      onclick={() => toggleRLS(table.name, table.has_rls)}
                    >
                      {table.has_rls ? 'Disable RLS' : 'Enable RLS'}
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>

      <!-- Active Policies -->
      <div class="card">
        <h2 style="font-size: 15px; margin-bottom: 16px;">Active Policies</h2>
        {#if policies.length === 0}
          <div class="empty-state">
            <p>No policies found</p>
          </div>
        {:else}
          <div class="table-wrapper">
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Table</th>
                  <th>Action</th>
                  <th>Roles</th>
                  <th>Definition</th>
                </tr>
              </thead>
              <tbody>
                {#each policies as policy}
                  <tr>
                    <td style="font-weight: 500;">{policy.name}</td>
                    <td style="font-family: var(--font-mono); font-size: 12px;">{policy.table}</td>
                    <td><span class="badge badge-info">{policy.action}</span></td>
                    <td>
                      {#each policy.roles as role}
                        <span class="badge badge-neutral" style="margin-right: 4px;">{role}</span>
                      {/each}
                    </td>
                    <td style="font-family: var(--font-mono); font-size: 11px; color: var(--text-muted);">
                      {policy.qualifier}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .page-content { padding: 0 24px 24px; }
  .error-card { 
    border-color: var(--status-error);
    background: rgba(255, 82, 82, 0.05);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }
</style>
