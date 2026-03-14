<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import { apiFetch, getErrorMessage, getHeaders, getOmniBaseUrl } from '$lib/api'

  const HISTORY_KEY = 'omnibase.sql_history'
  const MAX_HISTORY = 50

  let query = $state('SELECT * FROM storage.buckets;')
  let results = $state<any[]>([])
  let error = $state<string | null>(null)
  let loading = $state(false)
  let columns = $state<string[]>([])
  let migrationName = $state('')
  let queryHistory = $state<string[]>([])
  let showHistory = $state(false)
  let commandSummary = $state<{ command: string; rowsAffected: number; table?: string; schema?: string } | null>(null)

  function extractCreatedTable(sql: string) {
    const normalized = sql.trim().replace(/\s+/g, ' ')
    const match = normalized.match(/^create\s+table\s+(if\s+not\s+exists\s+)?(?:"?([a-zA-Z_][\w$]*)"?\.)?"?([a-zA-Z_][\w$]*)"?/i)
    if (!match) {
      return null
    }

    return {
      schema: match[2] || 'public',
      table: match[3],
    }
  }

  function rememberCreatedTable(schema: string, table: string) {
    const payload = { schema, table, createdAt: Date.now() }
    localStorage.setItem('omnibase.last_created_table', JSON.stringify(payload))
    window.dispatchEvent(new CustomEvent('omnibase:table-created', { detail: payload }))
  }

  async function waitForTable(schema: string, table: string, attempts = 8) {
    for (let i = 0; i < attempts; i += 1) {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/tables?schema=${encodeURIComponent(schema)}`, {
        headers: getHeaders(false)
      })

      if (resp.ok) {
        const tables = await resp.json()
        if (Array.isArray(tables) && tables.some((entry) => entry?.name === table && entry?.schema === schema)) {
          return true
        }
      }

      await new Promise((resolve) => setTimeout(resolve, 300 + i * 150))
    }

    return false
  }

  function loadHistory() {
    try {
      const raw = localStorage.getItem(HISTORY_KEY)
      queryHistory = raw ? JSON.parse(raw) : []
    } catch {
      queryHistory = []
    }
  }

  function saveToHistory(q: string) {
    const trimmed = q.trim()
    if (!trimmed) return
    let list = [trimmed, ...queryHistory.filter((x) => x.trim() !== trimmed)].slice(0, MAX_HISTORY)
    queryHistory = list
    try {
      localStorage.setItem(HISTORY_KEY, JSON.stringify(list))
    } catch {}
  }

  async function runQuery() {
    if (!query.trim()) return
    loading = true
    error = null
    results = []
    columns = []
    commandSummary = null

    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/query`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ query, migration_name: migrationName.trim() || undefined })
      })
      if (resp.ok) {
        const data = await resp.json()
        results = data
        if (Array.isArray(data) && data.length > 0 && data[0]?.command) {
          const target = extractCreatedTable(query)
          commandSummary = {
            command: data[0].command,
            rowsAffected: Number(data[0].rows_affected ?? 0),
            table: target?.table,
            schema: target?.schema,
          }
          if (target && /^CREATE TABLE/i.test(String(data[0].command))) {
            rememberCreatedTable(target.schema, target.table)
          }
        } else if (Array.isArray(data) && data.length > 0) {
          columns = Object.keys(data[0])
        }
        saveToHistory(query)
      } else {
        error = await getErrorMessage(resp, 'Failed to execute query')
      }
    } catch (e) {
      error = 'Could not connect to the API gateway.'
    } finally {
      loading = false
    }
  }

  function pickFromHistory(q: string) {
    query = q
    showHistory = false
  }

  async function openCreatedTable() {
    if (!commandSummary?.table) {
      return
    }

    const schema = commandSummary.schema || 'public'
    await waitForTable(schema, commandSummary.table)
    window.location.href = `/database?schema=${encodeURIComponent(schema)}&table=${encodeURIComponent(commandSummary.table)}`
  }

  onMount(() => {
    loadHistory()
    const tableParam = $page.url.searchParams.get('table')
    const queryParam = $page.url.searchParams.get('query')
    if (queryParam) {
      query = queryParam
    }
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
      <input class="input" style="width: 220px;" bind:value={migrationName} placeholder="Optional migration name" />
    </div>
    <div class="flex gap-2" style="position: relative;">
      <button class="btn btn-secondary btn-sm" onclick={() => showHistory = !showHistory} title="Query history">
        History {queryHistory.length ? `(${queryHistory.length})` : ''}
      </button>
      {#if showHistory}
        <div class="history-dropdown">
          {#each queryHistory.slice(0, 15) as q}
            <button type="button" class="history-item" onclick={() => pickFromHistory(q)}>
              <span class="history-preview">{q.slice(0, 60)}{q.length > 60 ? '…' : ''}</span>
            </button>
          {/each}
          {#if queryHistory.length === 0}
            <div class="history-empty">No history yet</div>
          {/if}
        </div>
      {/if}
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
      {:else if commandSummary}
        <div class="command-card">
          <div class="command-badge">{commandSummary.command}</div>
          <h3 class="command-title">
            {#if commandSummary.table}
              Table `{commandSummary.schema}.{commandSummary.table}` is ready
            {:else}
              Query completed successfully
            {/if}
          </h3>
          <p class="command-copy">
            DDL statements change schema structure, so Postgres reports `0 rows affected` even when the command succeeds.
          </p>
          <div class="command-meta">
            <span>Rows affected: {commandSummary.rowsAffected}</span>
            {#if commandSummary.table}
              <span>Schema refreshed</span>
            {/if}
          </div>
          {#if commandSummary.table}
            <div class="command-actions">
              <button class="btn btn-primary btn-sm" type="button" onclick={openCreatedTable}>Open in Table Editor</button>
              <span class="command-hint">OmniBase will keep this table selected when you switch views.</span>
            </div>
          {/if}
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
  .command-card {
    padding: 18px 20px;
    border: 1px solid rgba(0, 230, 118, 0.18);
    border-radius: 12px;
    background:
      linear-gradient(135deg, rgba(0, 230, 118, 0.08), rgba(108, 71, 255, 0.06)),
      var(--bg-surface);
  }
  .command-badge {
    display: inline-flex;
    padding: 5px 9px;
    border-radius: 999px;
    background: rgba(0, 230, 118, 0.12);
    color: #7fffb3;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
  }
  .command-title {
    margin: 14px 0 8px;
    font-size: 20px;
  }
  .command-copy {
    margin: 0;
    color: var(--text-secondary);
    font-size: 13px;
    max-width: 640px;
  }
  .command-meta {
    display: flex;
    gap: 10px;
    margin-top: 14px;
    flex-wrap: wrap;
  }
  .command-meta span {
    padding: 6px 10px;
    border-radius: 999px;
    background: rgba(255,255,255,0.04);
    color: var(--text-muted);
    font-size: 12px;
  }
  .command-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 18px;
    flex-wrap: wrap;
  }
  .command-hint {
    color: var(--text-muted);
    font-size: 12px;
  }
  .truncate {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .history-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    margin-top: 4px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    max-height: 280px;
    overflow-y: auto;
    z-index: 100;
    min-width: 280px;
  }
  .history-item {
    display: block;
    width: 100%;
    padding: 10px 14px;
    text-align: left;
    background: none;
    border: none;
    color: var(--text-primary);
    font-size: 12px;
    font-family: var(--font-mono);
    cursor: pointer;
    border-bottom: 1px solid var(--border-subtle);
  }
  .history-item:hover {
    background: var(--bg-elevated);
  }
  .history-preview {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: block;
  }
  .history-empty {
    padding: 16px;
    color: var(--text-muted);
    font-size: 13px;
  }
</style>
