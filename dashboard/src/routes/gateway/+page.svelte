<script lang="ts">
  const stats = [
    { label: 'Requests / sec', value: '24.8k', delta: '+12%', positive: true },
    { label: 'Avg Latency', value: '4.2ms', delta: '-8%', positive: true },
    { label: 'Error Rate', value: '0.03%', delta: '+0.01%', positive: false },
    { label: 'Active Connections', value: '1,842', delta: '+5%', positive: true },
  ]
</script>

<svelte:head><title>API Gateway — OmniBase</title></svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">API Gateway</h1>
    <p class="page-subtitle">Single entry point for all external traffic — routing, rate-limiting, WebSocket upgrades</p>
  </div>
  <button class="btn btn-primary">⚙ Configure</button>
</div>

<div class="page-content">
  <div class="grid-cols-4 gap-4" style="margin-bottom:24px">
    {#each stats as s}
      <div class="stat-card">
        <div class="stat-value">{s.value}</div>
        <div class="stat-label">{s.label}</div>
        <span class="stat-delta {s.positive ? 'positive' : 'negative'}">{s.delta}</span>
      </div>
    {/each}
  </div>

  <div class="card" style="margin-bottom:16px">
    <h3 style="font-size:14px;font-weight:600;margin-bottom:12px">Gateway Endpoints</h3>
    <div class="table-wrapper">
      <table>
        <thead><tr><th>Path Prefix</th><th>Target Service</th><th>Rate Limit</th><th>Auth Required</th><th>Status</th></tr></thead>
        <tbody>
          {#each [
            { path:'/api/v1/db',path2:'Database Service','rate':'10k/min',auth:true},
            { path:'/api/v1/auth',path2:'Identity Service','rate':'2k/min',auth:false},
            { path:'/api/v1/storage',path2:'Object Store','rate':'5k/min',auth:true},
            { path:'/api/v1/functions',path2:'Compute Service','rate':'1k/min',auth:true},
            { path:'/realtime',path2:'Live Sync Engine','rate':'unlimited',auth:true},
          ] as r}
            <tr>
              <td class="cell-mono">{r.path}</td>
              <td>{r.path2}</td>
              <td class="cell-mono">{r.rate}</td>
              <td><span class="badge {r.auth ? 'badge-success' : 'badge-neutral'}">{r.auth ? 'Yes' : 'No'}</span></td>
              <td><span class="badge badge-success">● Active</span></td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>
