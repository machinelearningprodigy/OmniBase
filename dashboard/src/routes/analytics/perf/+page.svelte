<svelte:head><title>Performance — OmniBase Analytics</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Performance Monitoring</h1><p class="page-subtitle">Network traces, screen rendering FPS, custom traces, cold start metrics</p></div></div>
<div class="page-content">
  <div class="grid-cols-4 gap-4" style="margin-bottom:24px">
    {#each [{l:'App Cold Start',v:'1.2s',d:'-0.3s',pos:true},{l:'API p99 Latency',v:'284ms',d:'+12ms',pos:false},{l:'Avg FPS',v:'58.2',d:'+0.4',pos:true},{l:'HTTP Error Rate',v:'0.8%',d:'-0.2%',pos:true}] as s}
      <div class="stat-card"><div class="stat-value">{s.v}</div><div class="stat-label">{s.l}</div><span class="stat-delta {s.pos?'positive':'negative'}">{s.d}</span></div>
    {/each}
  </div>
  <div class="card">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Slowest Network Requests</h3>
    <div class="table-wrapper">
      <table>
        <thead><tr><th>URL</th><th>Method</th><th>p50</th><th>p95</th><th>p99</th><th>Count</th></tr></thead>
        <tbody>
          {#each [
            {url:'/api/v1/db/query',method:'POST',p50:'12ms',p95:'88ms',p99:'240ms',count:'12k'},
            {url:'/api/v1/functions/generate-report',method:'POST',p50:'800ms',p95:'2.4s',p99:'5.1s',count:'84'},
            {url:'/storage/v1/upload',method:'PUT',p50:'340ms',p95:'1.2s',p99:'3.8s',count:'2.1k'},
          ] as r}
            <tr>
              <td class="cell-mono" style="font-size:11px">{r.url}</td>
              <td><span class="badge badge-neutral">{r.method}</span></td>
              <td class="cell-mono" style="color:#00e676">{r.p50}</td>
              <td class="cell-mono">{r.p95}</td>
              <td class="cell-mono" style="color:#ff9800">{r.p99}</td>
              <td class="cell-mono">{r.count}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>
