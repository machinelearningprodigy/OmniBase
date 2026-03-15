<svelte:head><title>Replication — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Replication</h1><p class="page-subtitle">Logical replication slots and streaming replication status</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Replication Slots</h3>
      <div class="table-wrapper">
        <table>
          <thead><tr><th>Slot Name</th><th>Plugin</th><th>Active</th><th>Lag</th></tr></thead>
          <tbody>
            {#each [{name:'omnibase_realtime',plugin:'pgoutput',active:true,lag:'0 B'},{name:'analytics_slot',plugin:'wal2json',active:false,lag:'142 MB'}] as s}
              <tr>
                <td class="cell-mono">{s.name}</td>
                <td class="cell-mono">{s.plugin}</td>
                <td><span class="badge {s.active ? 'badge-success' : 'badge-neutral'}">{s.active ? 'Active' : 'Paused'}</span></td>
                <td class="cell-mono">{s.lag}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">WAL Stats</h3>
      {#each [['WAL Level','logical'],['Max WAL Senders','10'],['Max Replication Slots','10'],['WAL Keep Size','1 GB']] as [k,v]}
        <div style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid var(--border-subtle)">
          <span style="font-size:12px;color:var(--text-secondary)">{k}</span>
          <span class="cell-mono" style="font-size:12px">{v}</span>
        </div>
      {/each}
    </div>
  </div>
</div>
