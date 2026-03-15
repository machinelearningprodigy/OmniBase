<svelte:head><title>Auth Performance — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Auth Performance</h1><p class="page-subtitle">Authentication endpoint response times, throughput, and error rates</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-4 gap-4" style="margin-bottom:24px">
    {#each [
      {label:'Login p50',value:'12ms',delta:'-3ms',pos:true},
      {label:'Login p99',value:'84ms',delta:'+2ms',pos:false},
      {label:'Token Refresh p50',value:'4ms',delta:'-1ms',pos:true},
      {label:'Auth Error Rate',value:'0.12%',delta:'-0.04%',pos:true},
    ] as s}
      <div class="stat-card">
        <div class="stat-value">{s.value}</div>
        <div class="stat-label">{s.label}</div>
        <span class="stat-delta {s.pos ? 'positive' : 'negative'}">{s.delta}</span>
      </div>
    {/each}
  </div>
  <div class="card">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:10px">Latency Breakdown by Endpoint</h3>
    <div class="table-wrapper">
      <table>
        <thead><tr><th>Endpoint</th><th>p50</th><th>p95</th><th>p99</th><th>RPS</th><th>Error %</th></tr></thead>
        <tbody>
          {#each [
            {ep:'POST /auth/v1/token?grant_type=password',p50:'12ms',p95:'48ms',p99:'84ms',rps:120,err:'0.1%'},
            {ep:'POST /auth/v1/token?grant_type=refresh_token',p50:'4ms',p95:'18ms',p99:'32ms',rps:840,err:'0.0%'},
            {ep:'POST /auth/v1/signup',p50:'88ms',p95:'210ms',p99:'440ms',rps:14,err:'0.5%'},
            {ep:'POST /auth/v1/otp',p50:'120ms',p95:'380ms',p99:'800ms',rps:6,err:'2.1%'},
          ] as r}
            <tr>
              <td class="cell-mono" style="font-size:11px">{r.ep}</td>
              <td class="cell-mono" style="color:#00e676">{r.p50}</td>
              <td class="cell-mono">{r.p95}</td>
              <td class="cell-mono" style="color:#ff9800">{r.p99}</td>
              <td class="cell-mono">{r.rps} req/s</td>
              <td class="cell-mono" style="color:{parseFloat(r.err)>1?'#ff5252':'#00e676'}">{r.err}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>
