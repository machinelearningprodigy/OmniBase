<svelte:head><title>Embeddings — OmniBase AI</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Embeddings</h1><p class="page-subtitle">Auto-generate embeddings on INSERT/UPDATE using your chosen model adapter</p></div></div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Embedding Jobs</h3>
      <div class="table-wrapper">
        <table>
          <thead><tr><th>Table.Column</th><th>Model</th><th>Queue Depth</th><th>Status</th></tr></thead>
          <tbody>
            {#each [
              {col:'documents.embedding',model:'text-embedding-3-large',queue:0,ok:true},
              {col:'posts.content_vec',model:'nomic-embed-text',queue:42,ok:true},
              {col:'products.cat_vec',model:'text-embedding-ada-002',queue:8,ok:false},
            ] as j}
              <tr>
                <td class="cell-mono" style="font-size:11px">{j.col}</td>
                <td class="cell-mono" style="font-size:11px">{j.model}</td>
                <td><span class="badge {j.queue>0?'badge-warning':'badge-success'}">{j.queue} pending</span></td>
                <td><span class="badge {j.ok?'badge-success':'badge-error'}">{j.ok?'● Running':'● Error'}</span></td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Model Adapters</h3>
      {#each [
        {name:'OpenAI text-embedding-3-large',dims:3072,status:'active'},
        {name:'OpenAI text-embedding-ada-002',dims:1536,status:'active'},
        {name:'Cohere embed-multilingual-v3',dims:1024,status:'inactive'},
        {name:'Ollama nomic-embed-text (local)',dims:768,status:'active'},
      ] as m}
        <div style="display:flex;align-items:center;gap:12px;padding:8px 0;border-bottom:1px solid var(--border-subtle)">
          <div style="flex:1;font-size:12px">{m.name}</div>
          <span class="badge badge-neutral" style="font-size:10px">{m.dims}d</span>
          <span class="badge {m.status==='active'?'badge-success':'badge-neutral'}">{m.status}</span>
        </div>
      {/each}
    </div>
  </div>
</div>
