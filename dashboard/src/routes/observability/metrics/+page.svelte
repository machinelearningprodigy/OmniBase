<svelte:head><title>Metrics — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Metrics</h1><p class="page-subtitle">OpenTelemetry metrics — request rates, error rates, and resource utilization</p></div></div>
<div class="page-content">
  <div class="grid-cols-4 gap-4" style="margin-bottom:24px">
    {#each [{l:'Req/sec',v:'24.8k',d:'+12%',pos:true},{l:'Error Rate',v:'0.03%',d:'-0.01%',pos:true},{l:'CPU',v:'28%',d:'+4%',pos:false},{l:'Memory',v:'1.8 GB',d:'-200MB',pos:true}] as s}
      <div class="stat-card"><div class="stat-value">{s.v}</div><div class="stat-label">{s.l}</div><span class="stat-delta {s.pos?'positive':'negative'}">{s.d}</span></div>
    {/each}
  </div>
  <div class="card" style="margin-bottom:16px">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Service Health</h3>
    <div class="table-wrapper">
      <table>
        <thead><tr><th>Service</th><th>Status</th><th>Uptime</th><th>p99 Latency</th><th>Error Rate</th></tr></thead>
        <tbody>
          {#each [
            {svc:'API Gateway',ok:true,up:'99.99%',lat:'8ms',err:'0.01%'},
            {svc:'Auth Service',ok:true,up:'99.95%',lat:'45ms',err:'0.08%'},
            {svc:'Database Service',ok:true,up:'100%',lat:'12ms',err:'0.00%'},
            {svc:'Storage Service',ok:true,up:'99.98%',lat:'120ms',err:'0.02%'},
            {svc:'Realtime Engine',ok:true,up:'99.90%',lat:'18ms',err:'0.04%'},
            {svc:'Functions Runtime',ok:false,up:'98.2%',lat:'200ms',err:'1.80%'},
          ] as r}
            <tr>
              <td>{r.svc}</td>
              <td><span class="badge {r.ok?'badge-success':'badge-error'}">● {r.ok?'Healthy':'Degraded'}</span></td>
              <td class="cell-mono">{r.up}</td>
              <td class="cell-mono">{r.lat}</td>
              <td class="cell-mono" style="color:{parseFloat(r.err)>0.5?'#ff5252':'#00e676'}">{r.err}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>
