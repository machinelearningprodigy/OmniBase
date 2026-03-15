<svelte:head><title>API Keys — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">API Keys</h1><p class="page-subtitle">Manage project API keys — anon key, service key, and custom scoped keys</p></div><button class="btn btn-primary">＋ Generate Key</button></div>
<div class="page-content">
  <div style="display:flex;flex-direction:column;gap:12px">
    {#each [
      {name:'Anon (Public) Key',key:'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.anon...A',desc:'Safe to use in browser code. Row Level Security enforced. Never bypass RLS with this key.',type:'anon'},
      {name:'Service Key (Private)',key:'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.service...B',desc:'Bypasses RLS. Never expose in client code.',type:'service'},
    ] as k}
      <div class="card">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:8px">
          <span style="font-weight:700;font-size:14px">{k.name}</span>
          <span class="badge {k.type==='anon'?'badge-success':'badge-warning'}">{k.type}</span>
        </div>
        <p style="font-size:12px;color:var(--text-secondary);margin-bottom:10px">{k.desc}</p>
        <div style="display:flex;gap:8px">
          <input class="input input-mono" value={k.key} type="password" style="flex:1;font-size:11px" readonly />
          <button class="btn btn-secondary btn-sm">👁 Show</button>
          <button class="btn btn-secondary btn-sm">📋 Copy</button>
        </div>
      </div>
    {/each}
  </div>
</div>
