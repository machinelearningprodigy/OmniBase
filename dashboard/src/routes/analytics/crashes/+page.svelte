<svelte:head><title>Crash Reports — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Crash Reports</h1><p class="page-subtitle">Automatic crash capture, stack trace symbolication, and user impact analysis</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-4 gap-4" style="margin-bottom:24px">
    {#each [{l:'Open Issues',v:'12',neg:true},{l:'Crashes Today',v:'84',neg:false},{l:'Affected Users',v:'342',neg:false},{l:'Crash-Free Rate',v:'99.6%',neg:false}] as s}
      <div class="stat-card"><div class="stat-value" style="color:{s.neg?'#ff5252':'inherit'}">{s.v}</div><div class="stat-label">{s.l}</div></div>
    {/each}
  </div>
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Issue</th><th>Occurrences</th><th>Users Affected</th><th>First Seen</th><th>Last Seen</th><th>Status</th></tr></thead>
      <tbody>
        {#each [
          {name:'NullPointerException in PaymentFlow',occ:42,users:38,first:'Mar 14',last:'2min ago',status:'open'},
          {name:'Network timeout in SyncService',occ:18,users:15,first:'Mar 12',last:'1hr ago',status:'open'},
          {name:'IndexOutOfBoundsException in TableView',occ:8,users:6,first:'Mar 13',last:'3hrs ago',status:'investigating'},
          {name:'WebSocket disconnect loop',occ:4,users:4,first:'Mar 15',last:'8min ago',status:'open'},
        ] as i}
          <tr>
            <td style="font-size:13px;max-width:280px">{i.name}</td>
            <td><span class="badge badge-error">{i.occ}</span></td>
            <td>{i.users} users</td>
            <td style="font-size:12px;color:var(--text-secondary)">{i.first}</td>
            <td style="font-size:12px;color:{i.last.includes('min')?'#ff5252':'var(--text-secondary)'}">{i.last}</td>
            <td><span class="badge {i.status==='open'?'badge-error':i.status==='investigating'?'badge-warning':'badge-success'}">{i.status}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
