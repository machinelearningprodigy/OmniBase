<svelte:head><title>Cron Scheduler — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Cron Scheduler</h1><p class="page-subtitle">Schedule functions to run on a recurring basis using cron expressions</p></div><button class="btn btn-primary">＋ Schedule Function</button></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Schedule Name</th><th>Function</th><th>Cron Expression</th><th>Next Run</th><th>Last Run</th><th>Status</th></tr></thead>
      <tbody>
        {#each [
          {name:'nightly-digest',fn:'send-digest',cron:'0 9 * * *',next:'Tomorrow 09:00',last:'1hr ago',ok:true},
          {name:'hourly-sync',fn:'sync-analytics',cron:'0 * * * *',next:'In 12min',last:'48min ago',ok:true},
          {name:'weekly-report',fn:'generate-report',cron:'0 8 * * 1',next:'Mon 08:00',last:'6 days ago',ok:false},
        ] as j}
          <tr>
            <td class="cell-mono">{j.name}</td>
            <td class="cell-mono">{j.fn}</td>
            <td class="cell-mono">{j.cron}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{j.next}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{j.last}</td>
            <td><span class="badge {j.ok ? 'badge-success' : 'badge-error'}">{j.ok ? '● Active' : '● Error'}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
