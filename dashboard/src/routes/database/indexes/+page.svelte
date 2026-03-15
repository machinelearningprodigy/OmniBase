<svelte:head><title>Indexes — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Indexes</h1><p class="page-subtitle">Manage database indexes to optimize query performance</p></div>
  <button class="btn btn-primary">＋ New Index</button>
</div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Index Name</th><th>Table</th><th>Columns</th><th>Type</th><th>Size</th><th>Scans</th></tr></thead>
      <tbody>
        {#each [
          {name:'users_email_idx',table:'users',cols:'email',type:'BTREE',size:'48 kB',scans:'1.2M'},
          {name:'posts_created_at_idx',table:'posts',cols:'created_at DESC',type:'BTREE',size:'112 kB',scans:'856k'},
          {name:'documents_embedding_idx',table:'documents',cols:'embedding',type:'HNSW',size:'2.4 MB',scans:'24k'},
        ] as idx}
          <tr>
            <td class="cell-mono">{idx.name}</td>
            <td>{idx.table}</td>
            <td class="cell-mono">{idx.cols}</td>
            <td><span class="badge badge-info">{idx.type}</span></td>
            <td class="cell-mono">{idx.size}</td>
            <td>{idx.scans}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
