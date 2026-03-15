<svelte:head><title>Deployments — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Function Deployments</h1><p class="page-subtitle">Version history — deploy, rollback, and compare function versions</p></div></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Version</th><th>Function</th><th>Deployed</th><th>Runtime</th><th>Invocations</th><th>Status</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {ver:'v14',fn:'send-email',deployed:'Mar 15 07:00',rt:'Node.js 20',calls:2100,active:true},
          {ver:'v13',fn:'send-email',deployed:'Mar 14 12:00',rt:'Node.js 20',calls:4200,active:false},
          {ver:'v8',fn:'process-payment',deployed:'Mar 15 06:00',rt:'Node.js 20',calls:840,active:true},
          {ver:'v7',fn:'process-payment',deployed:'Mar 10 09:00',rt:'Node.js 20',calls:1200,active:false},
        ] as d}
          <tr>
            <td class="cell-mono">{d.ver}</td>
            <td class="cell-mono">{d.fn}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{d.deployed}</td>
            <td><span class="badge badge-neutral">{d.rt}</span></td>
            <td class="cell-mono">{d.calls.toLocaleString()}</td>
            <td><span class="badge {d.active ? 'badge-success' : 'badge-neutral'}">{d.active ? '● Live' : 'Archived'}</span></td>
            <td>{#if !d.active}<button class="btn btn-secondary btn-sm">Rollback</button>{:else}<span style="font-size:11px;color:var(--text-muted)">current</span>{/if}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
