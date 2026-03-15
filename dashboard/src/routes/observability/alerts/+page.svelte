<svelte:head><title>Alerts — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Alerts</h1><p class="page-subtitle">Configure threshold-based alerts for errors, latency, and service health</p></div><button class="btn btn-primary">＋ New Alert</button></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Alert Name</th><th>Condition</th><th>Threshold</th><th>Notify</th><th>State</th></tr></thead>
      <tbody>
        {#each [
          {name:'High Error Rate',cond:'error_rate > X% for 5min',thresh:'2%',notify:'Email + Slack',state:'ok'},
          {name:'Auth Service Latency',cond:'p99 > Xms sustained',thresh:'500ms',notify:'Slack',state:'ok'},
          {name:'DB Connections',cond:'connection count > X',thresh:'180',notify:'Email',state:'firing'},
          {name:'Storage 80% Full',cond:'storage_used > X%',thresh:'80%',notify:'Email',state:'ok'},
        ] as a}
          <tr>
            <td>{a.name}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{a.cond}</td>
            <td class="cell-mono">{a.thresh}</td>
            <td style="font-size:12px">{a.notify}</td>
            <td><span class="badge {a.state==='firing'?'badge-error':'badge-success'}">{a.state==='firing'?'🔴 Firing':'● OK'}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
