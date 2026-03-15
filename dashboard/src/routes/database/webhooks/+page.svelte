<svelte:head><title>Webhooks — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Database Webhooks</h1><p class="page-subtitle">Trigger HTTP requests when database rows change</p></div>
  <button class="btn btn-primary">＋ New Webhook</button>
</div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Name</th><th>Table</th><th>Events</th><th>Endpoint</th><th>Last 24h</th><th>Status</th></tr></thead>
      <tbody>
        {#each [
          {name:'sync-crm',table:'users',events:'INSERT,UPDATE',url:'https://hooks.crm.example.com/sync',calls:142,status:'active'},
          {name:'notify-slack',table:'orders',events:'INSERT',url:'https://hooks.slack.com/T.../B.../xxx',calls:38,status:'active'},
        ] as w}
          <tr>
            <td class="cell-mono">{w.name}</td>
            <td>{w.table}</td>
            <td><div style="display:flex;gap:4px">{#each w.events.split(',') as e}<span class="badge badge-info">{e}</span>{/each}</div></td>
            <td class="cell-mono" style="font-size:11px;max-width:180px;overflow:hidden;text-overflow:ellipsis">{w.url}</td>
            <td>{w.calls} calls</td>
            <td><span class="badge badge-success">● {w.status}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
