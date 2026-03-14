<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import { apiFetch, getErrorMessage, getHeaders, getOmniBaseUrl } from '$lib/api'

  const HISTORY_KEY = 'omnibase.sql_history'
  const MAX_HISTORY = 50

  interface HistoryItem {
    id: string
    query: string
    timestamp: number
    name?: string
    pinned?: boolean
  }

  let activeTabId = $state('new')
  let query = $state('SELECT * FROM storage.buckets;')
  let results = $state<any[]>([])
  let error = $state<string | null>(null)
  let loading = $state(false)
  let columns = $state<string[]>([])
  let migrationName = $state('')
  let queryHistory = $state<HistoryItem[]>([])
  let commandSummary = $state<{ command: string; rowsAffected: number; table?: string; schema?: string } | null>(null)
  
  // Menu/Dropdown states
  let itemMenuId = $state<string | null>(null)
  let renamingId = $state<string | null>(null)
  let newName = $state('')

  // Sort history: pinned first, then by timestamp
  let sortedHistory = $derived([...queryHistory].sort((a, b) => {
    if (a.pinned && !b.pinned) return -1
    if (!a.pinned && b.pinned) return 1
    return b.timestamp - a.timestamp
  }))

  function extractCreatedTable(sql: string) {
    const normalized = sql.trim().replace(/\s+/g, ' ')
    const match = normalized.match(/^create\s+table\s+(if\s+not\s+exists\s+)?(?:"?([a-zA-Z_][\w$]*)"?\.)?"?([a-zA-Z_][\w$]*)"?/i)
    if (!match) return null
    return { schema: match[2] || 'public', table: match[3] }
  }

  function rememberCreatedTable(schema: string, table: string) {
    const payload = { schema, table, createdAt: Date.now() }
    localStorage.setItem('omnibase.last_created_table', JSON.stringify(payload))
    window.dispatchEvent(new CustomEvent('omnibase:table-created', { detail: payload }))
  }

  function loadHistory() {
    try {
      const raw = localStorage.getItem(HISTORY_KEY)
      queryHistory = raw ? JSON.parse(raw) : []
    } catch {
      queryHistory = []
    }
  }

  function syncHistory(list: HistoryItem[]) {
    queryHistory = list
    try {
      localStorage.setItem(HISTORY_KEY, JSON.stringify(list))
    } catch {}
  }

  function saveToHistory(q: string) {
    const trimmed = q.trim()
    if (!trimmed) return
    
    // Check if exactly same query exists (ignoring whitespace)
    const existing = queryHistory.find((x) => x.query.trim() === trimmed)
    if (existing) {
      // If it exists, just update timestamp to move to top (if not pinned differently)
      const list = queryHistory.map(item => 
        item.id === existing.id ? { ...item, timestamp: Date.now() } : item
      )
      syncHistory(list)
      return
    }

    const newItem: HistoryItem = {
      id: Math.random().toString(36).substring(2, 9),
      query: trimmed,
      timestamp: Date.now()
    }
    
    let list = [newItem, ...queryHistory].slice(0, MAX_HISTORY)
    syncHistory(list)
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

  function pickFile(item: HistoryItem) {
    query = item.query
    activeTabId = item.id
    results = []
    error = null
    columns = []
    commandSummary = null
  }

  function newFile() {
    query = ''
    activeTabId = 'new'
    results = []
    error = null
    columns = []
    commandSummary = null
    migrationName = ''
  }

  function deleteHistoryItem(id: string, e: Event) {
    e.stopPropagation()
    const filtered = queryHistory.filter(item => item.id !== id)
    syncHistory(filtered)
    if (activeTabId === id) {
      newFile()
    }
    itemMenuId = null
  }

  function clearHistory() {
    if (confirm('Are you sure you want to clear all query history?')) {
      syncHistory([])
      newFile()
    }
  }

  function togglePin(id: string, e: Event) {
    e.stopPropagation()
    const list = queryHistory.map(item => 
      item.id === id ? { ...item, pinned: !item.pinned } : item
    )
    syncHistory(list)
    itemMenuId = null
  }

  function startRename(item: HistoryItem, e: Event) {
    e.stopPropagation()
    renamingId = item.id
    newName = item.name || item.query.slice(0, 20).replace(/\s+/g, ' ')
    itemMenuId = null
  }

  function submitRename(id: string) {
    const list = queryHistory.map(item => {
      if (item.id === id) {
        return { ...item, name: newName.trim() || undefined }
      }
      return item
    })
    syncHistory(list)
    renamingId = null
  }

  function copyQuery(q: string, e: Event) {
    e.stopPropagation()
    navigator.clipboard.writeText(q)
    itemMenuId = null
  }

  function downloadResults() {
    if (results.length === 0) return
    const keys = Object.keys(results[0])
    const csv = [
      keys.join(','),
      ...results.map(row => keys.map(k => JSON.stringify(row[k])).join(','))
    ].join('\n')
    
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `query_results_${Date.now()}.csv`
    a.click()
  }
  
  function formatTime(ts: number) {
    const d = new Date(ts)
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
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

    const handleOutside = () => itemMenuId = null
    window.addEventListener('click', handleOutside)
    return () => window.removeEventListener('click', handleOutside)
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

<div class="editor-layout animate-fade-in">
  <!-- Sidebar -->
  <aside class="editor-sidebar">
    <div class="sidebar-header">
      <h2 class="sidebar-title">SQL Explorer</h2>
      <div class="flex gap-1">
        <button class="btn-icon" onclick={newFile} title="New Query">
           <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
        </button>
        <button class="btn-icon delete-all" onclick={clearHistory} title="Clear All History">
           <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
        </button>
      </div>
    </div>
    
    <div class="sidebar-section">
      <div class="section-title">HISTORY</div>
      <div class="file-list">
        {#each sortedHistory as item (item.id)}
          <div class="file-item-container">
            <button class="file-item {activeTabId === item.id ? 'active' : ''}" onclick={() => pickFile(item)}>
              <div class="indicator-stack">
                <svg class="file-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line></svg>
                {#if item.pinned}
                  <svg class="pin-indicator" width="8" height="8" viewBox="0 0 24 24" fill="currentColor"><path d="M21 14l-2.07-2.07A2 2 0 0 0 17.51 11H17V5h1V3H6v2h1v6h-.51a2 2 0 0 0-1.42.59L3 14h18zM11 19h2v2h-2v-2z"></path></svg>
                {/if}
              </div>
              <div class="file-meta">
                {#if renamingId === item.id}
                  <input 
                    class="rename-input" 
                    bind:value={newName} 
                    onkeydown={(e) => e.key === 'Enter' && submitRename(item.id)}
                    onblur={() => submitRename(item.id)}
                    autofocus
                    onclick={(e) => e.stopPropagation()}
                  />
                {:else}
                  <span class="file-preview">
                    {item.name || (item.query.slice(0, 24).replace(/\s+/g, ' ') + (item.query.length > 24 ? '...' : ''))}
                  </span>
                {/if}
                <span class="file-time">{formatTime(item.timestamp)}</span>
              </div>
            </button>
            <button class="item-menu-btn" onclick={(e) => { e.stopPropagation(); itemMenuId = itemMenuId === item.id ? null : item.id }}>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="1"></circle><circle cx="12" cy="5" r="1"></circle><circle cx="12" cy="19" r="1"></circle></svg>
            </button>
            
            {#if itemMenuId === item.id}
              <div class="context-menu" onclick={(e) => e.stopPropagation()}>
                <button onclick={(e) => togglePin(item.id, e)}>
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 14l-2.07-2.07A2 2 0 0 0 17.51 11H17V5h1V3H6v2h1v6h-.51a2 2 0 0 0-1.42.59L3 14h18zM11 19h2v2h-2v-2z"></path></svg>
                  {item.pinned ? 'Unpin' : 'Pin to Top'}
                </button>
                <button onclick={(e) => startRename(item, e)}>
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                  Rename
                </button>
                <button onclick={(e) => copyQuery(item.query, e)}>
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                  Copy Query
                </button>
                <div class="menu-divider"></div>
                <button class="delete-opt" onclick={(e) => deleteHistoryItem(item.id, e)}>
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                  Delete
                </button>
              </div>
            {/if}
          </div>
        {/each}
        {#if queryHistory.length === 0}
          <div class="sidebar-empty">No queries ran yet</div>
        {/if}
      </div>
    </div>
  </aside>

  <!-- Main Area -->
  <div class="editor-main">
    <div class="tab-bar">
      <div class="tab active">
        <svg style="margin-right: 6px; opacity: 0.7;" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline></svg>
        {#if activeTabId === 'new'}
          untitled.sql
        {:else}
          {queryHistory.find(i => i.id === activeTabId)?.name || `query_${activeTabId}.sql`}
        {/if}
        <span class="tab-close" onclick={newFile}>×</span>
      </div>
      <div class="tab-actions">
         <span style="font-size: 11px; color: var(--text-muted); margin-right: 12px; font-family: var(--font-sans);">Ctrl + Enter to run</span>
         <input class="migration-input" bind:value={migrationName} placeholder="Migration name (opt)" />
         <button class="btn btn-primary btn-sm" onclick={runQuery} disabled={loading}>
           {#if loading}
              <svg class="spinner" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 2v4"/></svg> Run...
           {:else}
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 4px;"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>
              Run Query
           {/if}
         </button>
      </div>
    </div>
    
    <div class="query-area">
      <div class="line-numbers">
        {#each query.split('\n') as _, i}
          <div class="line-num">{i + 1}</div>
        {/each}
      </div>
      <textarea
        class="code-input"
        bind:value={query}
        onkeydown={handleKeydown}
        placeholder="-- Type your SQL here&#10;SELECT * FROM public.users;"
        spellcheck="false"
      ></textarea>
    </div>

    <div class="results-area">
      <div class="results-header">
        <div class="flex items-center gap-4">
          <span class="results-title">Results Output</span>
          {#if results.length > 0}
            <span class="results-count">{results.length} rows returned</span>
          {/if}
        </div>
        {#if results.length > 0}
          <button class="btn btn-secondary btn-sm" onclick={downloadResults}>
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 4px;"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
            Export CSV
          </button>
        {/if}
      </div>
      
      <div class="results-content">
        {#if error}
          <div class="error-msg">
            <span style="font-weight: 600;">Postgres Error:</span><br>
            {error}
          </div>
        {:else if loading}
          <div class="flex items-center justify-center p-8">
            <div class="skeleton" style="width: 100%; height: 120px; border-radius: 8px;"></div>
          </div>
        {:else if results.length === 0 && !commandSummary}
          <div class="empty-results">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" style="opacity: 0.2; margin-bottom: 12px;"><path d="M4 14.899A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 2.5 8.242"></path><path d="M12 12v9"></path><path d="M8 17l4 4 4-4"></path></svg>
            <p>Run a query to see results</p>
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
            </div>
          </div>
        {:else}
          <div class="table-wrapper">
            <table>
              <thead>
                <tr>
                  <th style="width: 40px; text-align: center; color: var(--text-muted);">#</th> 
                  {#each columns as col}
                    <th>{col}</th>
                  {/each}
                </tr>
              </thead>
              <tbody>
                {#each results as row, i}
                  <tr>
                    <td style="text-align: center; color: var(--text-muted); font-size: 11px; background: rgba(0,0,0,0.1);">{i + 1}</td>
                    {#each columns as col}
                      <td>
                        <div class="cell-content" title={JSON.stringify(row[col])}>
                          {#if row[col] === null}
                            <span style="color: rgba(255,255,255,0.2); font-style: italic;">NULL</span>
                          {:else if typeof row[col] === 'boolean'}
                            <span style="color: var(--brand-blue);">{row[col]}</span>
                          {:else if typeof row[col] === 'number'}
                            <span style="color: var(--brand-green);">{row[col]}</span>
                          {:else if typeof row[col] === 'object'}
                            <span style="color: var(--status-warning);">{JSON.stringify(row[col])}</span>
                          {:else}
                            {row[col]}
                          {/if}
                        </div>
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
</div>

<style>
  .editor-layout {
    height: calc(100vh - var(--topbar-height));
    display: grid;
    grid-template-columns: 260px 1fr;
    background: var(--bg-base);
    overflow: hidden;
  }

  /* Sidebar */
  .editor-sidebar {
    background: #0d0d12;
    border-right: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .sidebar-header {
    padding: 16px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border-subtle);
  }
  .sidebar-title { font-size: 13px; font-weight: 600; color: #fff; margin: 0; }
  .btn-icon { background: transparent; border: none; color: #888; cursor: pointer; padding: 4px; border-radius: 4px; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
  .btn-icon:hover { background: rgba(255,255,255,0.05); color: #fff; }
  .delete-all:hover { color: var(--status-error); background: rgba(255, 82, 82, 0.1); }
  
  .sidebar-section { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
  .section-title { font-size: 10px; font-weight: 700; color: #555; padding: 16px 20px 8px; letter-spacing: 0.1em; }
  .file-list { flex: 1; overflow-y: auto; padding: 0 10px 10px; }
  
  .file-item-container { position: relative; display: flex; align-items: center; group-hover: flex; }
  .file-item {
    flex: 1;
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 10px 12px;
    background: transparent;
    border: none;
    border-radius: 8px;
    text-align: left;
    cursor: pointer;
    transition: all 0.15s;
    margin-bottom: 2px;
    overflow: hidden;
  }
  .file-item:hover { background: rgba(255,255,255,0.03); }
  .file-item.active { background: rgba(108, 71, 255, 0.1); }
  
  .indicator-stack { position: relative; flex-shrink: 0; }
  .file-icon { color: #666; margin-top: 2px; flex-shrink: 0; }
  .file-item.active .file-icon { color: var(--brand-primary); }
  .pin-indicator { position: absolute; bottom: -2px; right: -4px; color: var(--brand-primary); filter: drop-shadow(0 0 2px rgba(0,0,0,0.8)); }
  
  .file-meta { display: flex; flex-direction: column; gap: 4px; overflow: hidden; flex: 1; }
  .file-preview { font-family: var(--font-mono); font-size: 12px; color: #ccc; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .file-item.active .file-preview { color: #fff; font-weight: 500; }
  .file-time { font-size: 10px; color: #555; }

  .item-menu-btn {
    opacity: 0;
    background: transparent;
    border: none;
    color: #666;
    padding: 8px;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.2s;
    margin-right: 4px;
  }
  .file-item-container:hover .item-menu-btn { opacity: 1; }
  .item-menu-btn:hover { background: rgba(255,255,255,0.05); color: #fff; }

  .rename-input {
    background: #1a1a24;
    border: 1px solid var(--brand-primary);
    border-radius: 4px;
    color: #fff;
    font-family: var(--font-mono);
    font-size: 12px;
    padding: 2px 4px;
    width: 100%;
    outline: none;
  }

  .context-menu {
    position: absolute;
    top: 40px;
    right: 0;
    background: #1a1a24;
    border: 1px solid var(--border-subtle);
    border-radius: 8px;
    box-shadow: var(--shadow-xl);
    z-index: 100;
    min-width: 140px;
    padding: 4px;
    display: flex;
    flex-direction: column;
  }
  .context-menu button {
    background: transparent;
    border: none;
    color: #ccc;
    padding: 8px 12px;
    font-size: 12px;
    text-align: left;
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.15s;
  }
  .context-menu button:hover { background: rgba(255,255,255,0.05); color: #fff; }
  .context-menu .delete-opt { color: #ff5252; }
  .context-menu .delete-opt:hover { background: rgba(255, 82, 82, 0.1); }
  .menu-divider { height: 1px; background: #222; margin: 4px 0; }

  .sidebar-empty {
    padding: 20px 10px; font-size: 12px; color: #555; text-align: center;
  }

  /* Main Area */
  .editor-main {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: #08080c;
  }

  .tab-bar {
    height: 48px;
    min-height: 48px;
    background: #0d0d12;
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-right: 16px;
  }
  .tab {
    display: flex;
    align-items: center;
    height: 100%;
    padding: 0 20px;
    border-right: 1px solid var(--border-subtle);
    background: rgba(255,255,255,0.02);
    color: #888;
    font-size: 13px;
    font-family: var(--font-mono);
    cursor: pointer;
    position: relative;
  }
  .tab.active { background: #08080c; color: #fff; border-top: 2px solid var(--brand-primary); }
  .tab-close { margin-left: 12px; opacity: 0; transition: opacity 0.2s; font-size: 16px; line-height: 1; }
  .tab:hover .tab-close { opacity: 0.5; }
  .tab-close:hover { opacity: 1 !important; color: #ff5252; }

  .tab-actions { display: flex; align-items: center; gap: 12px; }
  .migration-input { height: 28px; background: rgba(0,0,0,0.3); border: 1px solid #222; border-radius: 4px; padding: 0 10px; color: #eee; font-size: 12px; outline: none; transition: border-color 0.2s; width: 160px; }
  .migration-input:focus { border-color: var(--brand-primary); }
  
  .query-area {
    height: 50%;
    display: flex;
    border-bottom: 1px solid var(--border-subtle);
    background: #08080c;
  }
  .line-numbers {
    width: 48px;
    padding: 24px 0;
    background: #050508;
    border-right: 1px solid #1a1a24;
    text-align: right;
    user-select: none;
    overflow: hidden;
  }
  .line-num {
    padding-right: 12px; font-family: var(--font-mono); font-size: 14px; line-height: 24px; color: #333; height: 24px;
  }
  .code-input {
    flex: 1;
    padding: 24px;
    background: transparent;
    color: #e6edf3;
    font-family: var(--font-mono);
    font-size: 14px;
    line-height: 24px;
    border: none;
    resize: none;
    outline: none;
  }

  .results-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: var(--bg-surface);
    overflow: hidden;
  }
  .results-header {
    padding: 12px 24px;
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: #0d0d12;
  }
  .results-title { font-size: 12px; font-weight: 700; color: #888; text-transform: uppercase; letter-spacing: 0.05em; }
  .results-count { font-size: 12px; color: var(--brand-blue); background: rgba(96, 165, 250, 0.1); padding: 2px 8px; border-radius: 4px; font-family: var(--font-mono); }

  .results-content { flex: 1; overflow: auto; padding: 0; }
  
  .empty-results { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; color: var(--text-muted); font-size: 14px; }
  
  .error-msg { margin: 24px; padding: 16px; background: rgba(255, 82, 82, 0.05); border: 1px solid rgba(255, 82, 82, 0.2); border-radius: 8px; color: var(--status-error); font-size: 13px; font-family: var(--font-mono); }
  
  .command-card { margin: 24px; padding: 24px; border: 1px solid rgba(0, 230, 118, 0.18); border-radius: 12px; background: linear-gradient(135deg, rgba(0, 230, 118, 0.04), rgba(108, 71, 255, 0.03)), var(--bg-surface); }
  .command-badge { display: inline-flex; padding: 5px 9px; border-radius: 6px; background: rgba(0, 230, 118, 0.12); color: #7fffb3; font-size: 11px; font-weight: 700; letter-spacing: 0.04em; font-family: var(--font-mono); }
  .command-title { margin: 16px 0 8px; font-size: 20px; color: #fff; }
  .command-copy { margin: 0; color: var(--text-secondary); font-size: 13px; }
  .command-meta { display: flex; gap: 10px; margin-top: 16px; }
  .command-meta span { padding: 6px 12px; border-radius: 6px; background: rgba(255,255,255,0.05); color: var(--text-muted); font-size: 12px; font-family: var(--font-mono); }

  .table-wrapper { width: 100%; height: 100%; overflow: auto; }
  table { width: 100%; border-collapse: collapse; text-align: left; }
  th { position: sticky; top: 0; background: #111116; padding: 10px 16px; font-size: 12px; color: #888; font-weight: 600; font-family: var(--font-mono); border-bottom: 1px solid var(--border-subtle); border-right: 1px solid var(--border-subtle); white-space: nowrap; z-index: 10;}
  td { padding: 8px 16px; font-size: 13px; font-family: var(--font-mono); color: #ddd; border-bottom: 1px solid rgba(255,255,255,0.03); border-right: 1px solid rgba(255,255,255,0.02); white-space: nowrap; max-width: 400px; }
  tr:hover td { background: rgba(255,255,255,0.01); }
  .cell-content { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

  @keyframes spin { 100% { transform: rotate(360deg); } }
  .spinner { animation: spin 1s linear infinite; }
</style>
