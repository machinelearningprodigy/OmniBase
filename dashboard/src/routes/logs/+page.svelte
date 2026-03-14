<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { getHeaders, getOmniBaseUrl } from '$lib/api'

  interface LogEntry {
    id: string
    method: string
    path: string
    status: number
    latency_ms: number
    ip: string
    timestamp: string
  }

  let logs = $state<LogEntry[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)
  let pollInterval: any

  async function loadLogs() {
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/admin/v1/logs`, { headers: getHeaders() })
      if (resp.ok) {
        logs = await resp.json()
        error = null
      } else {
        const data = await resp.json().catch(() => ({}))
        error = data.message || 'Failed to fetch logs'
      }
    } catch {
      error = 'Connection error — is the gateway running?'
    } finally {
      loading = false
    }
  }

  onMount(() => {
    loadLogs()
    pollInterval = setInterval(loadLogs, 2000)
  })

  onDestroy(() => {
    if (pollInterval) clearInterval(pollInterval)
  })

  function formatTime(iso: string) {
    try {
      return new Date(iso).toLocaleTimeString(undefined, {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        fractionalSecondDigits: 3,
      })
    } catch {
      return iso
    }
  }

  function getMethodColor(method: string) {
    const colors: Record<string, string> = {
      GET: 'var(--brand-primary)',
      POST: 'var(--status-success)',
      PUT: 'var(--status-warning)',
      DELETE: 'var(--status-error)',
      PATCH: 'var(--status-warning)',
      OPTIONS: 'var(--text-muted)'
    }
    return colors[method] || 'var(--text-primary)'
  }

  function getStatusColor(status: number) {
    if (status >= 500) return 'var(--status-error)'
    if (status >= 400) return 'var(--status-warning)'
    if (status >= 300) return 'var(--brand-secondary)'
    if (status >= 200) return 'var(--status-success)'
    return 'var(--text-muted)'
  }

  let searchQuery = $state('')
  let filteredLogs = $derived(logs.filter(l => 
    l.path.toLowerCase().includes(searchQuery.toLowerCase()) || 
    l.method.toLowerCase().includes(searchQuery.toLowerCase()) ||
    l.id.toLowerCase().includes(searchQuery.toLowerCase()) 
  ))
</script>

<svelte:head>
  <title>API Logs — OmniBase</title>
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">API Logs</h1>
    <p class="page-subtitle">Real-time HTTP request tracing across all services</p>
  </div>
  <div class="flex gap-2">
    <div class="search-box">
      <span style="opacity: 0.5;">🔍</span>
      <input type="text" bind:value={searchQuery} placeholder="Filter logs..." class="input" style="width: 200px; padding-left: 8px; border: none; background: none; font-size: 13px;" />
    </div>
    <button class="btn btn-secondary btn-sm" onclick={loadLogs}>Refresh</button>
  </div>
</div>

<div class="page-content animate-fade-in">
  <div class="card" style="padding: 0; overflow: hidden; display: flex; flex-direction: column; height: calc(100vh - 200px);">
    <div class="log-header">
      <div class="col-time">Time</div>
      <div class="col-method">Method</div>
      <div class="col-path">Path</div>
      <div class="col-status">Status</div>
      <div class="col-latency">Latency</div>
      <div class="col-ip">IP Address</div>
      <div class="col-trace">Trace ID</div>
    </div>

    <div class="log-body" style="flex: 1; overflow-y: auto;">
      {#if loading && logs.length === 0}
        {#each Array(10) as _}
          <div class="log-row skeleton" style="height: 40px; margin: 4px; border-radius: 4px;"></div>
        {/each}
      {:else if error}
        <div style="padding: 40px; text-align: center; color: var(--status-error);">
          <div style="font-size: 32px; margin-bottom: 16px;">⚠️</div>
          {error}
        </div>
      {:else if filteredLogs.length === 0}
        <div style="padding: 40px; text-align: center; color: var(--text-muted);">
          No logs match the current filter, or no requests have been made yet.
        </div>
      {:else}
        {#each filteredLogs as log (log.id)}
          <div class="log-row">
            <div class="col-time">{formatTime(log.timestamp)}</div>
            <div class="col-method" style="color: {getMethodColor(log.method)}; font-weight: 600;">{log.method}</div>
            <div class="col-path" title={log.path}>{log.path}</div>
            <div class="col-status">
              <span class="badge" style="background: transparent; border-color: {getStatusColor(log.status)}; color: {getStatusColor(log.status)}; font-size: 11px;">
                {log.status}
              </span>
            </div>
            <div class="col-latency" style={log.latency_ms > 500 ? 'color: var(--status-warning)' : ''}>
              {log.latency_ms.toFixed(1)}ms
            </div>
            <div class="col-ip">{log.ip}</div>
            <div class="col-trace">{log.id}</div>
          </div>
        {/each}
      {/if}
    </div>
  </div>
</div>

<style>
  .page-content { padding: 0 24px 24px; }
  .search-box {
    display: flex;
    align-items: center;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 0 12px;
  }
  
  .log-header {
    display: flex;
    padding: 12px 16px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    position: sticky;
    top: 0;
    z-index: 10;
  }
  
  .log-row {
    display: flex;
    padding: 10px 16px;
    border-bottom: 1px solid var(--border-subtle);
    font-family: var(--font-mono);
    font-size: 12px;
    align-items: center;
    transition: background 0.15s;
  }
  .log-row:hover { background: var(--bg-surface); }
  .log-row:last-child { border-bottom: none; }

  .col-time { width: 120px; color: var(--text-muted); flex-shrink: 0; }
  .col-method { width: 70px; flex-shrink: 0; }
  .col-path { flex: 1; min-width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: var(--text-primary); }
  .col-status { width: 70px; flex-shrink: 0; display: flex; justify-content: flex-start; }
  .col-latency { width: 90px; flex-shrink: 0; text-align: right; padding-right: 20px; color: var(--text-secondary); }
  .col-ip { width: 130px; flex-shrink: 0; color: var(--text-muted); }
  .col-trace { width: 180px; flex-shrink: 0; color: var(--text-muted); font-size: 11px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
</style>
