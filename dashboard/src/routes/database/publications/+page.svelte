<svelte:head><title>Publications — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Publications</h1><p class="page-subtitle">Postgres logical replication publications — control what gets streamed to the realtime engine</p></div>
  <button class="btn btn-primary">＋ New Publication</button>
</div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Publication</th><th>Tables</th><th>Operations</th><th>Active Subscribers</th></tr></thead>
      <tbody>
        {#each [
          {name:'omnibase_realtime',tables:'All Tables',ops:'INSERT, UPDATE, DELETE',subs:3},
          {name:'public_data',tables:'posts, comments',ops:'INSERT, UPDATE',subs:1},
        ] as p}
          <tr>
            <td class="cell-mono">{p.name}</td>
            <td>{p.tables}</td>
            <td class="cell-mono">{p.ops}</td>
            <td><span class="badge badge-info">{p.subs}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
