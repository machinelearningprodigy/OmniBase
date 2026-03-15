<script lang="ts">
  import { onMount } from 'svelte'
  import { getHeaders, getOmniBaseUrl } from '$lib/api'

  interface Migration {
    id: number
    name: string
    executed_at: string
  }

  let migrations = $state<Migration[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)
  let expandedMigrationId = $state<number | null>(null)

  function toggleExpand(id: number) {
    expandedMigrationId = expandedMigrationId === id ? null : id
  }

  async function loadMigrations() {
    try {
      loading = true
      error = null
      const resp = await fetch(`${getOmniBaseUrl()}/pg/migrations`, { headers: getHeaders() })
      const data = await resp.json()
      if (resp.ok) {
        migrations = data
      } else {
        error = data.error || 'Failed to load migrations'
      }
    } catch (e) {
      error = 'Connection error'
    } finally {
      loading = false
    }
  }

  onMount(loadMigrations)
</script>

<svelte:head>
  <title>Migrations — OmniBase</title>
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">Database Migrations</h1>
    <p class="page-subtitle">Track and manage versioned changes to your schema</p>
  </div>
  <div class="flex gap-2">
    <button class="btn btn-secondary btn-sm" onclick={loadMigrations}>Refresh</button>
    <button class="btn btn-primary btn-sm" onclick={() => window.location.href = '/database/editor'}>
      New Migration
    </button>
  </div>
</div>

<div class="page-content">
  {#if loading}
    <div class="skeleton" style="height: 300px; border-radius: 12px;"></div>
  {:else if error}
    <div class="card" style="border-color: var(--status-error);">
      <p style="color: var(--status-error);">{error}</p>
    </div>
  {:else if migrations.length === 0}
    <div class="empty-state card">
      <div style="font-size: 40px; margin-bottom: 16px;">📜</div>
      <h3>No migration history found</h3>
      <p>Once you run migrations via the SQL Editor or CLI, they will appear here.</p>
    </div>
  {:else}
    <div class="card animate-fade-in" style="padding: 0; overflow: hidden;">
      <div class="table-wrapper">
        <table>
          <thead>
            <tr>
              <th style="width: 60px;">ID</th>
              <th>Name</th>
              <th>Executed At</th>
              <th style="text-align: right;">Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each migrations as m}
              {@const isExpanded = expandedMigrationId === m.id}
              <tr>
                <td style="color: var(--text-muted); font-family: var(--font-mono);">{m.id}</td>
                <td style="font-weight: 500;">{m.name}</td>
                <td style="color: var(--text-secondary);">{new Date(m.executed_at).toLocaleString()}</td>
                <td style="text-align: right;">
                   <button class="btn btn-secondary btn-xs" onclick={() => toggleExpand(m.id)}>
                     {isExpanded ? 'Hide SQL' : 'View SQL'}
                   </button>
                </td>
              </tr>
              {#if isExpanded}
                <tr>
                  <td colspan="4" style="background: rgba(0,0,0,0.2); padding: 20px;">
                    <div style="background: #050508; border: 1px solid var(--border-subtle); border-radius: 8px; padding: 16px; position: relative;">
                      <pre style="margin: 0; font-size: 13px; color: #86efac; overflow: auto; max-height: 400px; white-space: pre-wrap;"><code>{m.query}</code></pre>
                      <div style="position: absolute; top: 10px; right: 10px; font-size: 10px; color: var(--text-muted); text-transform: uppercase; font-weight: 700;">SQL SOURCE</div>
                    </div>
                  </td>
                </tr>
              {/if}
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>

<style>
  .page-content { padding: 0 24px 24px; }
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 60px 20px;
    text-align: center;
  }
</style>
