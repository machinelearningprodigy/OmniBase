<svelte:head><title>DB Backups — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Database Backups</h1><p class="page-subtitle">Automated and on-demand backups — restore to any point in time</p></div>
  <button class="btn btn-primary">⬇ Create Backup Now</button>
</div>
<div class="page-content">
  <div class="grid-cols-3 gap-4" style="margin-bottom:24px">
    {#each [{l:'Last Backup',v:'2h ago',ok:true},{l:'Retention',v:'30 days',ok:true},{l:'Storage Used',v:'4.2 GB',ok:true}] as s}
      <div class="card" style="text-align:center">
        <div style="font-size:22px;font-weight:700;color:{s.ok ? '#00e676' : '#ff5252'}">{s.v}</div>
        <div style="font-size:11px;color:var(--text-secondary);margin-top:4px">{s.l}</div>
      </div>
    {/each}
  </div>
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Backup ID</th><th>Created</th><th>Size</th><th>Type</th><th>Status</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {id:'bkp_01j2x...',created:'Mar 15 2026 06:00',size:'2.1 GB',type:'Scheduled',status:'complete'},
          {id:'bkp_01j2w...',created:'Mar 14 2026 06:00',size:'2.0 GB',type:'Scheduled',status:'complete'},
          {id:'bkp_01j2v...',created:'Mar 13 2026 14:22',size:'1.9 GB',type:'Manual',status:'complete'},
        ] as b}
          <tr>
            <td class="cell-mono">{b.id}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{b.created}</td>
            <td class="cell-mono">{b.size}</td>
            <td><span class="badge badge-neutral">{b.type}</span></td>
            <td><span class="badge badge-success">● {b.status}</span></td>
            <td style="display:flex;gap:6px">
              <button class="btn btn-secondary btn-sm">Restore</button>
              <button class="btn btn-secondary btn-sm">Download</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
