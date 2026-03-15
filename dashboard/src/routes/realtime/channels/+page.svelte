<svelte:head><title>Channels — OmniBase Live Sync</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Channels</h1><p class="page-subtitle">Manage realtime channels — subscribe clients, monitor traffic, set access rules</p></div>
  <button class="btn btn-primary">＋ New Channel</button>
</div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Channel</th><th>Active Subscribers</th><th>Messages/min</th><th>Type</th><th>RLS</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {name:'room:general',subs:48,msgs:120,type:'public',rls:false},
          {name:'user:alice',subs:1,msgs:4,type:'private',rls:true},
          {name:'project:abc123',subs:12,msgs:38,type:'authenticated',rls:true},
          {name:'broadcast:alerts',subs:142,msgs:3,type:'public',rls:false},
        ] as c}
          <tr>
            <td class="cell-mono">{c.name}</td>
            <td><span class="badge badge-info">{c.subs}</span></td>
            <td class="cell-mono">{c.msgs}</td>
            <td><span class="badge badge-neutral">{c.type}</span></td>
            <td><span class="badge {c.rls ? 'badge-success' : 'badge-neutral'}">{c.rls ? 'Enforced' : 'None'}</span></td>
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
