<svelte:head><title>Extensions — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Extensions</h1><p class="page-subtitle">Enable and manage PostgreSQL extensions to add powerful capabilities</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    {#each [
      {name:'pgvector',desc:'Store and query vector embeddings for AI/ML workloads',enabled:true,version:'0.7.0'},
      {name:'pg_graphql',desc:'Auto-generate GraphQL API from your Postgres schema',enabled:true,version:'1.5.0'},
      {name:'pg_cron',desc:'Run scheduled jobs inside PostgreSQL',enabled:true,version:'1.6.0'},
      {name:'pgaudit',desc:'Full audit trail of all database operations',enabled:false,version:'16.0'},
      {name:'timescaledb',desc:'Time-series data with continuous aggregates',enabled:false,version:'2.15'},
      {name:'apache_age',desc:'Property graph queries with Cypher language',enabled:false,version:'1.5.0'},
      {name:'pg_duckdb',desc:'Analytical queries 600x faster via DuckDB',enabled:false,version:'0.0.1'},
      {name:'pg_bm25',desc:'BM25 full-text ranking for hybrid search',enabled:false,version:'0.7.0'},
    ] as ext}
      <div class="card" style="display:flex;align-items:start;gap:16px">
        <div style="flex:1">
          <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px">
            <span style="font-weight:700;font-family:var(--font-mono);font-size:13px">{ext.name}</span>
            <span class="badge badge-neutral" style="font-size:10px">v{ext.version}</span>
            {#if ext.enabled}<span class="badge badge-success">Enabled</span>{/if}
          </div>
          <p style="font-size:12px;color:var(--text-secondary)">{ext.desc}</p>
        </div>
        <button class="btn btn-sm {ext.enabled ? 'btn-danger' : 'btn-primary'}">{ext.enabled ? 'Disable' : 'Enable'}</button>
      </div>
    {/each}
  </div>
</div>
