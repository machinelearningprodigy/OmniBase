<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { getRealtimeUrl } from '$lib/api'

  let logs = $state<{ time: string; event: string; payload: unknown }[]>([])
  let ws: WebSocket | null = null
  let status = $state('disconnected')

  onMount(() => {
    connect()
  })

  onDestroy(() => {
    if (ws) ws.close()
  })

  function connect() {
    status = 'connecting'
    ws = new WebSocket(getRealtimeUrl())

    ws.onopen = () => {
      status = 'connected'
      // Subscribe to all changes in public schema
      ws?.send(JSON.stringify({
        type: 'subscribe',
        channel: 'inspector',
        config: {
          postgres_changes: [
            { event: '*', schema: 'public', table: '*' }
          ]
        }
      }))
    }

    ws.onmessage = (e) => {
      const data = JSON.parse(e.data)
      logs = [{
        time: new Date().toLocaleTimeString(),
        event: data.type || data.event || 'message',
        payload: data
      }, ...logs].slice(0, 100) // keep last 100
    }

    ws.onclose = () => {
      status = 'disconnected'
      setTimeout(connect, 5000)
    }
  }
</script>

<svelte:head>
  <title>Realtime Inspector — OmniBase</title>
</svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">Realtime Inspector</h1>
    <p class="page-subtitle">Watch your WAL database changes and WebSocket messages live</p>
  </div>
  <div class="flex gap-2">
    <div class="badge {status === 'connected' ? 'badge-success' : 'badge-warning'}">
      {status}
    </div>
    <button class="btn btn-secondary btn-sm" onclick={() => logs = []}>Clear</button>
  </div>
</div>

<div class="page-content">
  <div class="card" style="padding: 0; display:flex; flex-direction:column; height: calc(100vh - 160px);">
    <div style="flex: 1; overflow-y: auto; background: #0a0e17; font-family: var(--font-mono); font-size: 12px; padding: 16px;">
      {#if logs.length === 0}
        <div style="color: var(--text-muted); text-align: center; margin-top: 40px;">
          Listening for Realtime Postgres WAL changes...<br>
          <span style="font-size: 10px; opacity: 0.6;">(Make sure replication is enabled in your database settings)</span>
        </div>
      {/if}
      
      {#each logs as log}
        <div style="margin-bottom: 12px; border-bottom: 1px solid rgba(255,255,255,0.05); padding-bottom: 12px; word-break: break-all;">
          <div style="display: flex; gap: 8px; margin-bottom: 6px;">
            <span style="color: var(--text-muted); width: 80px; flex-shrink: 0;">[{log.time}]</span>
            <span style="color: {log.event === 'INSERT' ? '#4ade80' : log.event === 'UPDATE' ? '#60a5fa' : log.event === 'DELETE' ? '#f87171' : '#facc15'}; font-weight: 600;">{log.event}</span>
          </div>
          <pre style="margin: 0; margin-left: 88px; color: #a5b4fc; background: rgba(0,0,0,0.3); padding: 8px; border-radius: 4px; white-space: pre-wrap;">{JSON.stringify(log.payload, null, 2)}</pre>
        </div>
      {/each}
    </div>
  </div>
</div>
