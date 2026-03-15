<svelte:head><title>Model Adapters — OmniBase AI</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Model Adapters</h1><p class="page-subtitle">Connect to AI model providers for embedding generation, completions, and re-ranking</p></div><button class="btn btn-primary">＋ Add Adapter</button></div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    {#each [
      {name:'OpenAI',models:'gpt-4o, text-embedding-3-large',icon:'⚡',status:'connected',key:'sk-…R4xK'},
      {name:'Antropic Claude',models:'claude-3-5-sonnet-20241022',icon:'🤖',status:'not_configured',key:''},
      {name:'Cohere',models:'command-r-plus, embed-multilingual-v3',icon:'🌊',status:'not_configured',key:''},
      {name:'Ollama (Local)',models:'llama3.2, nomic-embed-text',icon:'🦙',status:'connected',key:'localhost:11434'},
      {name:'Mistral AI',models:'mistral-large, mistral-embed',icon:'🌪️',status:'not_configured',key:''},
      {name:'Google Gemini',models:'gemini-1.5-pro, text-embedding-004',icon:'💎',status:'not_configured',key:''},
    ] as a}
      <div class="card" style="display:flex;align-items:start;gap:16px">
        <div style="font-size:28px">{a.icon}</div>
        <div style="flex:1">
          <div style="font-weight:700;font-size:14px;margin-bottom:2px">{a.name}</div>
          <div style="font-size:11px;color:var(--text-secondary);margin-bottom:8px">{a.models}</div>
          {#if a.key}<code style="font-size:10px;color:var(--text-muted)">{a.key}</code>{/if}
        </div>
        <div style="display:flex;flex-direction:column;align-items:flex-end;gap:8px">
          <span class="badge {a.status==='connected'?'badge-success':'badge-neutral'}">{a.status==='connected'?'● Connected':'Not set up'}</span>
          <button class="btn btn-secondary btn-sm">{a.status==='connected'?'Edit':'Connect'}</button>
        </div>
      </div>
    {/each}
  </div>
</div>
