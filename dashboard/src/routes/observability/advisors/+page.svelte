<svelte:head><title>Advisors — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Advisors</h1><p class="page-subtitle">Intelligent recommendations for query performance, security, and cost optimization</p></div></div>
<div class="page-content">
  <div style="display:flex;flex-direction:column;gap:12px">
    {#each [
      {type:'performance',icon:'⚡',title:'Index Missing on orders.user_id',desc:'Table "orders" has 48k rows but no index on user_id. 12 slow queries detected that filter on this column. Estimated speedup: 40×.',action:'Create Index',severity:'high'},
      {type:'security',icon:'🔒',title:'Table "documents" has no RLS policies',desc:'"documents" is accessible to all authenticated users without row-level filtering. Consider adding RLS policies to restrict per-user access.',action:'Configure RLS',severity:'critical'},
      {type:'performance',icon:'⚡',title:'8 idle replication slots detected',desc:'Unused replication slots hold WAL files and may cause disk to fill over time. Delete unused slots.',action:'Review Slots',severity:'warning'},
      {type:'cost',icon:'💰',title:'Embedding queue exceeding threshold',desc:'documents.embedding has 2,400 pending rows. Auto-embedding job is backlogged. Consider upgrading your embedding model quota.',action:'View Embeddings',severity:'medium'},
    ] as a}
      <div class="card" style="display:flex;gap:16px;border-left:3px solid {a.severity==='critical'?'#ff5252':a.severity==='high'?'#ff9800':a.severity==='warning'?'#ffab40':'#6c47ff'}">
        <div style="font-size:28px;flex-shrink:0">{a.icon}</div>
        <div style="flex:1">
          <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px">
            <span style="font-weight:700;font-size:13px">{a.title}</span>
            <span class="badge {a.severity==='critical'?'badge-error':a.severity==='high'?'badge-warning':'badge-neutral'}">{a.severity}</span>
          </div>
          <p style="font-size:12px;color:var(--text-secondary);margin-bottom:10px">{a.desc}</p>
          <button class="btn btn-primary btn-sm">{a.action} →</button>
        </div>
      </div>
    {/each}
  </div>
</div>
