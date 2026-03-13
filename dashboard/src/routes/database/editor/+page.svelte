<script lang="ts">
  import { onMount } from 'svelte'
  import { authStore } from '$lib/stores/auth'

  let query = 'SELECT * FROM users LIMIT 10;'
  let results: Record<string, unknown>[] | null = null
  let error: string | null = null
  let executing = false
  let execTime = 0

  const OMNIBASE_URL = 'http://localhost:8000'

  async function executeQuery() {
    executing = true
    error = null
    results = null
    const start = performance.now()

    try {
      // In a real app, you would pass the anon key or personal access token
      // Currently, the meta service requires service_role for raw SQL.
      // For local dev, we will assume we pass a token that resolves to service_role,
      // or we temporarily bypass if it's not setup. 
      // Based on our implementation, user needs a service_role token.
      
      const session = typeof localStorage !== 'undefined' ? localStorage.getItem('omnibase.session') : null
      const token = session ? JSON.parse(session).access_token : null

      const resp = await fetch(`${OMNIBASE_URL}/meta/query`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({ query })
      })

      const data = await resp.json()
      if (resp.ok) {
        results = data || []
      } else {
        error = data.error || 'Execution failed'
      }
    } catch (e) {
      error = 'Could not connect to API Gateway'
    } finally {
      execTime = Math.round(performance.now() - start)
      executing = false
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      executeQuery()
    }
  }
</script>

<svelte:head>
  <title>SQL Editor — OmniBase</title>
</svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">SQL Editor</h1>
    <p class="page-subtitle">Run raw SQL queries directly against your database</p>
  </div>
  <div class="flex gap-2">
    <button class="btn btn-primary" on:click={executeQuery} disabled={executing}>
      {executing ? 'Running...' : 'Run Query'} <kbd style="margin-left:8px;font-size:10px;">⌘↵</kbd>
    </button>
  </div>
</div>

<div class="page-content" style="display:flex; flex-direction:column; gap:16px; height: calc(100vh - 160px);">
  <!-- Editor Area -->
  <div class="card" style="flex: 1; padding: 0;">
    <textarea 
      class="input"
      bind:value={query} 
      on:keydown={handleKeydown}
      style="width: 100%; height: 100%; resize: none; border: none; font-family: var(--font-mono); font-size: 13px; padding: 16px; background: transparent; color: var(--text-primary); outline: none;"
      placeholder="Type your SQL query here..."></textarea>
  </div>

  <!-- Results Area -->
  <div class="card" style="flex: 1; display:flex; flex-direction:column; overflow: hidden; padding: 0;">
    <div style="padding: 12px 16px; border-bottom: 1px solid var(--border-subtle); display:flex; justify-content:space-between; align-items:center;">
      <h3 style="font-size: 13px; font-weight: 600;">Results</h3>
      {#if execTime > 0}
        <span style="font-size: 12px; color: var(--text-muted);">Executed in {execTime}ms</span>
      {/if}
    </div>
    <div style="flex: 1; overflow: auto; padding: 16px;">
      {#if error}
        <div style="color: var(--status-error); font-family: var(--font-mono); font-size: 13px;">
          ERROR: {error}
        </div>
      {:else if results}
        {#if results.length === 0}
          <div style="color: var(--text-muted); font-size: 13px;">Success. No rows returned.</div>
        {:else}
          <table style="width: 100%; border-collapse: collapse; font-family: var(--font-mono); font-size: 13px;">
            <thead style="text-align: left; border-bottom: 2px solid var(--border-subtle);">
              <tr>
                {#each Object.keys(results[0] || {}) as col}
                  <th style="padding: 8px;">{col}</th>
                {/each}
              </tr>
            </thead>
            <tbody>
              {#each results as row, i (i)}
                <tr style="border-bottom: 1px solid var(--border-subtle);">
                  {#each Object.keys(row) as col}
                    <td style="padding: 8px; color: var(--text-secondary);">
                      {row[col] === null ? 'NULL' : typeof row[col] === 'object' ? JSON.stringify(row[col]) : String(row[col])}
                    </td>
                  {/each}
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      {:else}
        <div style="color: var(--text-muted); font-size: 13px;">Run a query to see results here.</div>
      {/if}
    </div>
  </div>
</div>
