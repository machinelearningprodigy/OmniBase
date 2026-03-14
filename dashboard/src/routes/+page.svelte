<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { getHeaders, getOmniBaseUrl, getRealtimeUrl } from '$lib/api'
  import { serviceHealth } from '$lib/stores/health'
  import type { ServiceStatus } from '$lib/stores/health'

  interface StatCard {
    label: string
    value: string
    delta?: string
    positive?: boolean
    loading: boolean
  }

  let stats = $state<StatCard[]>([
    { label: 'Total Tables', value: '—', loading: true },
    { label: 'Active Users', value: '—', delta: 'all time', positive: true, loading: true },
    { label: 'Storage Buckets', value: '—', loading: true },
    { label: 'Connected Clients', value: '—', delta: 'realtime', positive: true, loading: true },
  ])

  interface ActivityEvent {
    type: string
    table: string
    time: string
  }

  let recentActivity = $state<ActivityEvent[]>([])
  let ws: WebSocket | null = null
  let wsStatus = $state<'connected' | 'connecting' | 'disconnected'>('disconnected')

  let health = $derived($serviceHealth)

  async function loadStats() {
    // Load table count
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/pg/tables?schema=public`, { headers: getHeaders() })
      if (resp.ok) {
        const tables = await resp.json()
        stats[0] = { ...stats[0], value: String(tables.length), loading: false }
      } else {
        stats[0] = { ...stats[0], value: 'N/A', loading: false }
      }
    } catch {
      stats[0] = { ...stats[0], value: 'N/A', loading: false }
    }

    // Load user count
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/admin/users?page=1&per_page=1`, { headers: getHeaders() })
      if (resp.ok) {
        const data = await resp.json()
        stats[1] = { ...stats[1], value: String(data.total || 0), loading: false }
      } else {
        stats[1] = { ...stats[1], value: 'N/A', loading: false }
      }
    } catch {
      stats[1] = { ...stats[1], value: 'N/A', loading: false }
    }

    // Load bucket count
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/storage/v1/bucket`, { headers: getHeaders() })
      if (resp.ok) {
        const buckets = await resp.json()
        stats[2] = { ...stats[2], value: String(Array.isArray(buckets) ? buckets.length : 0), loading: false }
      } else {
        stats[2] = { ...stats[2], value: 'N/A', loading: false }
      }
    } catch {
      stats[2] = { ...stats[2], value: 'N/A', loading: false }
    }

    // Connected clients will update via WebSocket status
    stats[3] = { ...stats[3], value: wsStatus === 'connected' ? '1' : '0', loading: false }
  }

  function connectRealtime() {
    wsStatus = 'connecting'
    try {
      ws = new WebSocket(getRealtimeUrl())

      ws.onopen = () => {
        wsStatus = 'connected'
        stats[3] = { ...stats[3], value: '1', loading: false }
        ws?.send(JSON.stringify({
          type: 'subscribe',
          channel: 'dashboard_feed',
          config: {
            postgres_changes: [{ event: '*', schema: 'public', table: '*' }]
          }
        }))
      }

      ws.onmessage = (e) => {
        try {
          const data = JSON.parse(e.data)
          if (data.event && data.table) {
            recentActivity = [{
              type: data.event || data.type,
              table: data.table,
              time: 'just now'
            }, ...recentActivity].slice(0, 10)
          }
        } catch {}
      }

      ws.onclose = () => {
        wsStatus = 'disconnected'
        stats[3] = { ...stats[3], value: '0', loading: false }
        // Auto reconnect after 5s
        setTimeout(connectRealtime, 5000)
      }

      ws.onerror = () => { wsStatus = 'disconnected' }
    } catch {
      wsStatus = 'disconnected'
    }
  }

  function statusBadgeClass(status: ServiceStatus['status']) {
    return {
      'healthy': 'badge-success',
      'degraded': 'badge-warning',
      'down': 'badge-error',
      'unknown': 'badge-neutral',
    }[status] ?? 'badge-neutral'
  }

  onMount(async () => {
    serviceHealth.refresh()
    await loadStats()
    connectRealtime()
  })

  onDestroy(() => {
    if (ws) ws.close()
  })
</script>

<svelte:head>
  <title>OmniBase — Dashboard</title>
  <meta name="description" content="OmniBase project overview and health status" />
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">Project Overview</h1>
    <p class="page-subtitle">Welcome to OmniBase — your open-source BaaS platform</p>
  </div>
  <div class="flex gap-2">
    <a href="/database/editor" class="btn btn-secondary btn-sm">SQL Editor</a>
    <a href="/database" class="btn btn-primary btn-sm">Table Editor</a>
  </div>
</div>

<div class="page-content">
  <!-- Stat Cards -->
  <div style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px;">
    {#each stats as stat, i}
      <div class="stat-card animate-fade-in">
        <div class="stat-label">{stat.label}</div>
        {#if stat.loading}
          <div class="skeleton" style="height: 32px; width: 60px; border-radius: 6px; margin: 6px 0;"></div>
        {:else}
          <div class="stat-value" style="
            {i === 0 ? 'color: var(--brand-primary)' : i === 1 ? 'color: var(--brand-secondary)' :
             i === 2 ? 'color: #4ade80' : 'color: #f59e0b'}
          ">{stat.value}</div>
        {/if}
        {#if stat.delta}
          <div class="stat-delta {stat.positive ? 'positive' : ''}">{stat.delta}</div>
        {/if}
      </div>
    {/each}
  </div>

  <div style="display: grid; grid-template-columns: 1fr 340px; gap: 20px; align-items: start;">
    <!-- Services Health -->
    <div class="card animate-fade-in">
      <div class="flex items-center justify-between mb-4" style="margin-bottom: 16px;">
        <h2 style="font-size: 15px; font-weight: 600; margin: 0;">Service Health</h2>
        <div class="flex gap-2 items-center">
          {#if health.lastChecked}
            <span style="font-size: 11px; color: var(--text-muted);">
              Checked {Math.round((Date.now() - health.lastChecked.getTime()) / 1000)}s ago
            </span>
          {/if}
          <button class="btn btn-secondary btn-sm" onclick={() => serviceHealth.refresh()} disabled={health.loading}>
            {health.loading ? '⟳' : 'Refresh'}
          </button>
        </div>
      </div>

      <div style="display: flex; flex-direction: column; gap: 10px;">
        {#each health.services as svc}
          <div style="
            padding: 12px 16px; background: var(--bg-elevated);
            border-radius: var(--radius-md); border: 1px solid var(--border-subtle);
            display: flex; align-items: center; justify-content: space-between;
            transition: border-color 0.2s;
            {svc.status === 'healthy' ? 'border-color: rgba(74, 222, 128, 0.2)' :
             svc.status === 'down' ? 'border-color: rgba(255,82,82,0.2)' : ''}
          ">
            <div class="flex items-center gap-2">
              <span class="status-dot {svc.status === 'healthy' ? 'online' : svc.status === 'down' ? 'error' : 'offline'}"></span>
              <div>
                <div style="font-size: 13px; font-weight: 500;">{svc.name}</div>
                <div style="font-size: 11px; color: var(--text-muted); font-family: var(--font-mono);">{svc.url}</div>
              </div>
            </div>
            <div class="flex items-center gap-2">
              {#if svc.latency}
                <span style="font-size: 11px; color: var(--text-secondary); font-family: var(--font-mono);">
                  {svc.latency}ms
                </span>
              {/if}
              <span class="badge {statusBadgeClass(svc.status)}">{svc.status}</span>
            </div>
          </div>
        {/each}
      </div>

      <!-- Overall health summary -->
      {#if health.services.length > 0}
        {@const healthyCount = health.services.filter(s => s.status === 'healthy').length}
        <div style="margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border-subtle); display: flex; justify-content: space-between; align-items: center;">
          <span style="font-size: 12px; color: var(--text-muted);">
            {healthyCount}/{health.services.length} services healthy
          </span>
          <div style="height: 6px; flex: 1; margin: 0 12px; background: var(--bg-base); border-radius: 99px; overflow: hidden;">
            <div style="
              height: 100%;
              width: {(healthyCount / health.services.length * 100).toFixed(0)}%;
              background: linear-gradient(90deg, var(--brand-primary), var(--brand-secondary));
              border-radius: 99px;
              transition: width 0.5s ease;
            "></div>
          </div>
        </div>
      {/if}
    </div>

    <div style="display: flex; flex-direction: column; gap: 16px;">
      <!-- Realtime Feed -->
      <div class="card animate-fade-in">
        <div class="flex items-center justify-between" style="margin-bottom: 12px;">
          <h2 style="font-size: 15px; font-weight: 600; margin: 0;">Realtime Feed</h2>
          <div class="flex items-center gap-2">
            <span class="status-dot {wsStatus === 'connected' ? 'online' : wsStatus === 'connecting' ? 'offline' : 'error'}"></span>
            <span style="font-size: 11px; color: var(--text-muted);">{wsStatus}</span>
          </div>
        </div>

        <div style="display: flex; flex-direction: column; gap: 8px; min-height: 100px;">
          {#if recentActivity.length === 0}
            <div style="text-align: center; color: var(--text-muted); font-size: 12px; padding: 20px 0;">
              {wsStatus === 'connected' ? 'Listening for DB changes...' : 'Connecting to realtime service...'}
            </div>
          {:else}
            {#each recentActivity as event}
              <div class="flex items-center gap-2" style="font-size: 12px;">
                <span class="badge {event.type === 'INSERT' ? 'badge-success' : event.type === 'UPDATE' ? 'badge-info' : 'badge-error'}"
                  style="width: 52px; justify-content: center; flex-shrink: 0;">
                  {event.type}
                </span>
                <span style="color: var(--text-secondary); font-family: var(--font-mono);">{event.table}</span>
                <span style="color: var(--text-muted); margin-left: auto;">{event.time}</span>
              </div>
            {/each}
          {/if}
        </div>
      </div>

      <!-- Quick Links -->
      <div class="card animate-fade-in">
        <h2 style="font-size: 15px; font-weight: 600; margin-bottom: 12px;">Quick Actions</h2>
        <div style="display: flex; flex-direction: column; gap: 6px;">
          {#each [
            { label: '📊 Table Editor', href: '/database', desc: 'Browse and edit data' },
            { label: '💻 SQL Editor', href: '/database/editor', desc: 'Run raw SQL queries' },
            { label: '🔒 RLS Policies', href: '/database/rls', desc: 'Manage row-level security' },
            { label: '👤 Manage Users', href: '/auth/users', desc: 'Auth user management' },
            { label: '📦 Storage', href: '/storage', desc: 'File buckets and objects' },
            { label: '⚡ Realtime', href: '/realtime', desc: 'Live change inspector' },
          ] as link}
            <a href={link.href} style="
              display: flex; align-items: center; justify-content: space-between;
              padding: 10px 12px; border-radius: var(--radius-md);
              text-decoration: none; color: var(--text-primary);
              border: 1px solid var(--border-subtle);
              transition: all 0.15s;
            " onmouseenter={(e) => (e.currentTarget as HTMLElement).style.borderColor = 'var(--border-brand)'}
               onmouseleave={(e) => (e.currentTarget as HTMLElement).style.borderColor = 'var(--border-subtle)'}>
              <div>
                <div style="font-size: 13px; font-weight: 500;">{link.label}</div>
                <div style="font-size: 11px; color: var(--text-muted);">{link.desc}</div>
              </div>
              <span style="color: var(--text-muted);">→</span>
            </a>
          {/each}
        </div>
      </div>
    </div>
  </div>

  <!-- Quick Start SDK -->
  <div class="card animate-fade-in" style="margin-top: 20px; background: linear-gradient(135deg, rgba(108, 71, 255, 0.08) 0%, rgba(0, 212, 255, 0.04) 100%); border-color: var(--border-brand);">
    <h2 style="font-size: 15px; font-weight: 600; margin-bottom: 16px;">Quick Start with OmniBase SDK</h2>
    <pre style="margin: 0;"><code><span style="color: #7dd3fc;">import</span> <span style="color: #f0f0ff;">{'{ createClient }'}</span> <span style="color: #7dd3fc;">from</span> <span style="color: #86efac;">'@omnibase/omnibase-js'</span>

<span style="color: #7dd3fc;">const</span> omni = createClient(
  <span style="color: #86efac;">'http://localhost:8000'</span>,  <span style="color: #6272a4;">// Your OmniBase URL</span>
  <span style="color: #86efac;">'your-anon-key'</span>           <span style="color: #6272a4;">// From project settings</span>
)

<span style="color: #6272a4;">// Query your database — fully typed</span>
<span style="color: #7dd3fc;">const</span> {'{ data, error }'} = <span style="color: #7dd3fc;">await</span> omni
  .from(<span style="color: #86efac;">'posts'</span>)
  .select(<span style="color: #86efac;">'*'</span>)
  .eq(<span style="color: #86efac;">'published'</span>, <span style="color: #ff79c6;">true</span>)
  .order(<span style="color: #86efac;">'created_at'</span>, {'{ ascending: false }'})</code></pre>
  </div>
</div>

<style>
  .page-content { padding: 0 24px 24px; }
</style>
