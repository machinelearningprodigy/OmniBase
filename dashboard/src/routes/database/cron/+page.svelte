<svelte:head><title>Cron Jobs — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Cron Jobs</h1><p class="page-subtitle">Schedule recurring jobs inside PostgreSQL using pg_cron</p></div>
  <button class="btn btn-primary">＋ New Job</button>
</div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Job Name</th><th>Schedule</th><th>Command</th><th>Last Run</th><th>Status</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {name:'cleanup_sessions',cron:'0 2 * * *',cmd:'DELETE FROM sessions WHERE expires_at < NOW()',last:'2 hrs ago',status:'success'},
          {name:'aggregate_stats',cron:'*/15 * * * *',cmd:'CALL refresh_stats_mv()',last:'12 min ago',status:'success'},
          {name:'send_digests',cron:'0 9 * * 1',cmd:'SELECT send_weekly_digest()',last:'6 days ago',status:'success'},
        ] as j}
          <tr>
            <td class="cell-mono">{j.name}</td>
            <td class="cell-mono">{j.cron}</td>
            <td class="cell-mono" style="max-width:200px;overflow:hidden;text-overflow:ellipsis">{j.cmd}</td>
            <td style="color:var(--text-secondary);font-size:12px">{j.last}</td>
            <td><span class="badge badge-success">● {j.status}</span></td>
            <td style="display:flex;gap:6px">
              <button class="btn btn-secondary btn-sm">Edit</button>
              <button class="btn btn-danger btn-sm">Delete</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
