<svelte:head><title>Runtimes — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Runtimes</h1><p class="page-subtitle">Available serverless function runtimes — write in your language of choice</p></div></div>
<div class="page-content">
  <div class="grid-cols-3 gap-4">
    {#each [
      {lang:'Node.js',version:'20 LTS',emoji:'🟩',status:'stable',coldStart:'~80ms'},
      {lang:'Python',version:'3.12',emoji:'🐍',status:'stable',coldStart:'~120ms'},
      {lang:'Go (WASM)',version:'1.22',emoji:'🐹',status:'stable',coldStart:'~2ms'},
      {lang:'Rust (WASM)',version:'stable',emoji:'🦀',status:'stable',coldStart:'~1ms'},
      {lang:'PHP',version:'8.3',emoji:'🐘',status:'beta',coldStart:'~200ms'},
      {lang:'Ruby',version:'3.3',emoji:'💎',status:'beta',coldStart:'~180ms'},
      {lang:'Java',version:'21 LTS',emoji:'☕',status:'planned',coldStart:'~400ms'},
      {lang:'Swift',version:'5.10',emoji:'🍎',status:'planned',coldStart:'~50ms'},
      {lang:'Dart',version:'3.x',emoji:'🎯',status:'planned',coldStart:'~60ms'},
    ] as r}
      <div class="card">
        <div style="font-size:28px;margin-bottom:8px">{r.emoji}</div>
        <div style="font-weight:700;font-size:14px;margin-bottom:2px">{r.lang}</div>
        <div style="font-size:12px;color:var(--text-secondary);margin-bottom:10px">{r.version}</div>
        <div style="display:flex;align-items:center;justify-content:space-between">
          <span class="badge {r.status==='stable'?'badge-success':r.status==='beta'?'badge-warning':'badge-neutral'}">{r.status}</span>
          <span style="font-size:11px;color:var(--text-muted)">cold: {r.coldStart}</span>
        </div>
      </div>
    {/each}
  </div>
</div>
