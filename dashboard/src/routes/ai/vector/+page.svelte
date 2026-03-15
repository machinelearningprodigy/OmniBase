<svelte:head><title>Vector Search — OmniBase AI</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Vector Search</h1><p class="page-subtitle">Store and query high-dimensional embeddings with pgvector — cosine, L2, inner product similarity</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-3 gap-4" style="margin-bottom:24px">
    {#each [{l:'Vector Columns',v:'4'},{l:'Total Embeddings',v:'142k'},{l:'Avg Query Time',v:'8ms'}] as s}
      <div class="stat-card"><div class="stat-value">{s.v}</div><div class="stat-label">{s.l}</div></div>
    {/each}
  </div>
  <div class="grid-cols-2 gap-4">
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Vector Columns</h3>
      <div class="table-wrapper">
        <table>
          <thead><tr><th>Table.Column</th><th>Dimensions</th><th>Index</th><th>Rows</th></tr></thead>
          <tbody>
            {#each [
              {col:'documents.embedding',dims:1536,idx:'HNSW',rows:'48k'},
              {col:'products.search_vec',dims:768,idx:'IVFFlat',rows:'12k'},
              {col:'posts.content_vec',dims:1536,idx:'HNSW',rows:'82k'},
            ] as c}
              <tr>
                <td class="cell-mono">{c.col}</td>
                <td class="cell-mono">{c.dims}</td>
                <td><span class="badge badge-info">{c.idx}</span></td>
                <td>{c.rows}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Similarity Search Test</h3>
      <div class="form-group"><label class="label">Table.Column</label>
        <select class="input"><option>documents.embedding</option><option>products.search_vec</option></select>
      </div>
      <div class="form-group"><label class="label">Query Text (auto-embedded)</label>
        <input class="input" placeholder="e.g. machine learning for beginners" />
      </div>
      <div class="form-group"><label class="label">Top K Results</label>
        <input class="input" type="number" value="10" style="max-width:100px" />
      </div>
      <button class="btn btn-primary">🔍 Search</button>
    </div>
  </div>
</div>
