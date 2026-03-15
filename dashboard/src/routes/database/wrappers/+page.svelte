<svelte:head><title>Wrappers — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Wrappers</h1><p class="page-subtitle">Foreign Data Wrappers — query external data sources directly from PostgreSQL</p></div>
  <button class="btn btn-primary">＋ Add Wrapper</button>
</div>
<div class="page-content">
  <div class="grid-cols-3 gap-4">
    {#each [
      {name:'Stripe',desc:'Query Stripe invoices, customers, and payments as Postgres tables',icon:'💳',enabled:false},
      {name:'Firebase',desc:'Import Firebase Firestore documents into Postgres directly',icon:'🔥',enabled:false},
      {name:'S3',desc:'Query files in any S3-compatible bucket as a Postgres table',icon:'🪣',enabled:true},
      {name:'Redis / Valkey',desc:'Read Redis key-value data in SQL queries',icon:'⚡',enabled:false},
      {name:'ClickHouse',desc:'Federate analytical queries to a ClickHouse cluster',icon:'📊',enabled:false},
      {name:'MongoDB',desc:'Read MongoDB collections as foreign Postgres tables',icon:'🍃',enabled:false},
    ] as w}
      <div class="card">
        <div style="font-size:28px;margin-bottom:8px">{w.icon}</div>
        <div style="font-weight:700;font-size:14px;margin-bottom:4px">{w.name}</div>
        <p style="font-size:12px;color:var(--text-secondary);margin-bottom:12px">{w.desc}</p>
        <button class="btn btn-sm {w.enabled ? 'btn-danger' : 'btn-primary'}">{w.enabled ? 'Disconnect' : 'Connect'}</button>
      </div>
    {/each}
  </div>
</div>
