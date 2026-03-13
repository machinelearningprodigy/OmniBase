<script lang="ts">
  import { onMount } from 'svelte'
  import { serviceHealth } from '$lib/stores/health'
  import type { ServiceStatus } from '$lib/stores/health'

  interface StatCard {
    label: string
    value: string
    delta?: string
    positive?: boolean
  }

  let stats = $state<StatCard[]>([
    { label: 'Database Rows', value: '—', delta: '+12% today', positive: true },
    { label: 'Active Users', value: '—', delta: '+3 today', positive: true },
    { label: 'API Requests', value: '—', delta: 'last 24h', positive: true },
    { label: 'Storage Used', value: '—' },
  ])

  let recentActivity = $state([
    { type: 'INSERT', table: 'users', time: '2 min ago', icon: '✦' },
    { type: 'UPDATE', table: 'posts', time: '5 min ago', icon: '✦' },
    { type: 'DELETE', table: 'sessions', time: '8 min ago', icon: '✦' },
    { type: 'INSERT', table: 'comments', time: '12 min ago', icon: '✦' },
  ])

  let health = $derived($serviceHealth)

  onMount(async () => {
    serviceHealth.refresh()
  })

  function statusBadgeClass(status: ServiceStatus['status']) {
    return {
      'healthy': 'badge-success',
      'degraded': 'badge-warning',
      'down': 'badge-error',
      'unknown': 'badge-neutral',
    }[status] ?? 'badge-neutral'
  }
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
    <a href="/database" class="btn btn-primary btn-sm">New Table</a>
  </div>
</div>

<div class="page-content">
  <!-- Stat Cards -->
  <div class="grid-cols-4 gap-4 mb-4" style="display: grid; gap: 16px; margin-bottom: 24px;">
    {#each stats as stat}
      <div class="stat-card animate-fade-in">
        <div class="stat-label">{stat.label}</div>
        <div class="stat-value">{stat.value}</div>
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
        <h2 style="font-size: 15px; font-weight: 600;">Service Health</h2>
        <button class="btn btn-secondary btn-sm" onclick={() => serviceHealth.refresh()}>
          Refresh
        </button>
      </div>

      <div style="display: flex; flex-direction: column; gap: 12px;">
        {#each health.services as svc}
          <div class="flex items-center justify-between" style="padding: 12px 16px; background: var(--bg-elevated); border-radius: var(--radius-md); border: 1px solid var(--border-subtle);">
            <div class="flex items-center gap-2">
              <span class="status-dot {svc.status === 'healthy' ? 'online' : svc.status === 'down' ? 'error' : 'offline'}"></span>
              <div>
                <div style="font-size: 13px; font-weight: 500;">{svc.name}</div>
                <div style="font-size: 11px; color: var(--text-muted);">{svc.url}</div>
              </div>
            </div>
            <div class="flex items-center gap-2">
              {#if svc.latency}
                <span style="font-size: 11px; color: var(--text-secondary); font-family: var(--font-mono);">{svc.latency}ms</span>
              {/if}
              <span class="badge {statusBadgeClass(svc.status)}">{svc.status}</span>
            </div>
          </div>
        {/each}
      </div>
    </div>

    <!-- Recent Activity -->
    <div class="card animate-fade-in">
      <h2 style="font-size: 15px; font-weight: 600; margin-bottom: 16px;">Recent Activity</h2>
      <div style="display: flex; flex-direction: column; gap: 10px;">
        {#each recentActivity as event}
          <div class="flex items-center gap-2" style="font-size: 12px;">
            <span class="badge {event.type === 'INSERT' ? 'badge-success' : event.type === 'UPDATE' ? 'badge-info' : 'badge-error'}"
              style="width: 52px; justify-content: center;"
            >
              {event.type}
            </span>
            <span style="color: var(--text-secondary); font-family: var(--font-mono);">{event.table}</span>
            <span style="color: var(--text-muted); margin-left: auto;">{event.time}</span>
          </div>
        {/each}
      </div>
    </div>
  </div>

  <!-- Quick Start -->
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
