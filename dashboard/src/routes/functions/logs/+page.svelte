<svelte:head><title>Function Logs — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Function Logs</h1><p class="page-subtitle">Real-time streaming logs from all edge functions</p></div>
  <select class="input" style="width:200px"><option>All Functions</option><option>send-email</option><option>process-payment</option></select>
</div>
<div class="page-content">
  <div class="card" style="font-family:var(--font-mono);font-size:12px;line-height:1.8;background:var(--bg-elevated)">
    {#each [
      {ts:'07:47:33.421',fn:'send-email',level:'INFO',msg:'Sending email to alice@example.com'},
      {ts:'07:47:33.488',fn:'send-email',level:'INFO',msg:'Email dispatched via SMTP. Message-ID: abc123@mail'},
      {ts:'07:47:28.12',fn:'process-payment',level:'INFO',msg:'Processing payment intent pi_3Nx... amount=4999'},
      {ts:'07:47:28.340',fn:'process-payment',level:'WARN',msg:'Stripe response delay >200ms, retrying...'},
      {ts:'07:47:28.512',fn:'process-payment',level:'INFO',msg:'Payment succeeded: charge_id=ch_3Nx...'},
      {ts:'07:46:55.007',fn:'sync-crm',level:'ERROR',msg:'CRM API returned 503 Service Unavailable'},
      {ts:'07:46:55.008',fn:'sync-crm',level:'INFO',msg:'Scheduled retry in 30s (attempt 1/3)'},
    ] as l}
      <div style="display:flex;gap:16px;padding:3px 0;border-bottom:1px solid rgba(255,255,255,0.02)">
        <span style="color:#5050a0;flex-shrink:0">{l.ts}</span>
        <span style="color:#7dd3fc;flex-shrink:0;width:140px">{l.fn}</span>
        <span style="flex-shrink:0;width:50px;color:{l.level==='ERROR'?'#ff5252':l.level==='WARN'?'#ffab40':'#00e676'}">{l.level}</span>
        <span style="color:#c4b5fd">{l.msg}</span>
      </div>
    {/each}
  </div>
</div>
