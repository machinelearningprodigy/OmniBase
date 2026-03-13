<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'

  let query = $state('SELECT * FROM storage.buckets;')
  let results = $state<any[]>([])
  let error = $state<string | null>(null)
  let loading = $state(false)
  let columns = $state<string[]>([])

  const OMNIBASE_URL = 'http://localhost:8000'

  function getHeaders() {
    const session = typeof localStorage !== 'undefined' ? localStorage.getItem('omnibase.session') : null
    const token = session ? JSON.parse(session).access_token : null
    return {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
  }

  async function runQuery() {
    if (!query.trim()) return
    loading = true
    error = null
    results = []
    columns = []

    try {
      const resp = await fetch(`${OMNIBASE_URL}/pg/query`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ query })
      })
      const data = await resp.json()
      if (resp.ok) {
        results = data
        if (data.length > 0) {
          columns = Object.keys(data[0])
        }
      } else {
        error = data.error || 'Failed to execute query'
      }
    } catch (e) {
      error = 'Could not connect to the API gateway.'
    } finally {
      loading = false
    }
  }

  onMount(() => {
    const tableParam = $page.url.searchParams.get('table')
    if (tableParam) {
      query = `SELECT * FROM public."${tableParam}" LIMIT 100;`
      runQuery()
    }
  })

  function handleKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      runQuery()
    }
  }
</script>

<svelte:head>
  <title>SQL Editor — OmniBase</title>
</svelte:head>

<div class="editor-container animate-fade-in">
  <div class="editor-header">
    <div class="flex items-center gap-4">
      <h1 class="page-title" style="font-size: 16px; margin: 0;">SQL Editor</h1>
      <span style="font-size: 11px; color: var(--text-muted);">Ctrl + Enter to run</span>
    </div>
    <div class="flex gap-2">
      <button class="btn btn-secondary btn-sm" onclick={() => query = ''}>Clear</button>
      <button class="btn btn-primary btn-sm" onclick={runQuery} disabled={loading}>
        {loading ? 'Running...' : 'Run Query'}
      </button>
    </div>
  </div>

  <div class="editor-main">
    <div class="query-input-wrapper">
      <textarea
        bind:value={query}
        onkeydown={handleKeydown}
        placeholder="Enter your SQL here..."
        spellcheck="false"
      ></textarea>
    </div>

    <div class="results-container">
      {#if error}
        <div class="error-msg">
          <span style="font-weight: 600;">Postgres Error:</span><br>
          {error}
        </div>
      {:else if loading}
        <div class="flex items-center justify-center" style="height: 100px;">
          <div class="skeleton" style="width: 100%; height: 100%; border-radius: 8px;"></div>
        </div>
      {:else if results.length === 0}
        <div class="empty-results">
          {results.length === 0 && !loading && !error ? 'No results to display' : ''}
        </div>
      {:else}
        <div class="table-wrapper">
          <div style="font-size: 11px; color: var(--text-muted); margin-bottom: 8px;">{results.length} rows returned</div>
          <table>
            <thead>
              <tr>
                {#each columns as col}
                  <th>{col}</th>
                {/each}
              </tr>
            </thead>
            <tbody>
              {#each results as row}
                <tr>
                  {#each columns as col}
                    <td>
                      <span class="truncate" style="display: block; max-width: 300px;" title={JSON.stringify(row[col])}>
                        {row[col] === null ? 'NULL' : typeof row[col] === 'object' ? JSON.stringify(row[col]) : row[col]}
                      </span>
                    </td>
                  {/each}
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .editor-container {
    height: calc(100vh - var(--topbar-height));
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .editor-header {
    padding: 12px 24px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .editor-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .query-input-wrapper {
    height: 40%;
    min-height: 150px;
    border-bottom: 1px solid var(--border-subtle);
  }
  textarea {
    width: 100%;
    height: 100%;
    padding: 20px;
    background: #0d1117;
    color: #e6edf3;
    font-family: var(--font-mono);
    font-size: 14px;
    line-height: 1.5;
    border: none;
    resize: none;
    outline: none;
  }
  .results-container {
    flex: 1;
    overflow: auto;
    padding: 20px 24px;
    background: var(--bg-surface);
  }
  .error-msg {
    padding: 16px;
    background: rgba(255, 82, 82, 0.05);
    border: 1px solid rgba(255, 82, 82, 0.2);
    border-radius: 8px;
    color: var(--status-error);
    font-size: 13px;
    font-family: var(--font-mono);
  }
  .empty-results {
    color: var(--text-muted);
    font-size: 13px;
    text-align: center;
    margin-top: 40px;
  }
  .table-wrapper {
    overflow-x: auto;
  }
  .truncate {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
