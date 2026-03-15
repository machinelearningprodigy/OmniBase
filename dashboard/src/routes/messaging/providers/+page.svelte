<svelte:head><title>Messaging Providers — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Messaging Providers</h1><p class="page-subtitle">Configure adapters for all notification channels</p></div></div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    {#each [
      {name:'Twilio',type:'SMS',status:'connected',icon:'📱'},
      {name:'Resend',type:'Email',status:'connected',icon:'📧'},
      {name:'AWS SES',type:'Email',status:'not_configured',icon:'📨'},
      {name:'Mailgun',type:'Email',status:'not_configured',icon:'📬'},
      {name:'Apple APNs',type:'Push',status:'connected',icon:'🍎'},
      {name:'FCM HTTP v1',type:'Push',status:'connected',icon:'🤖'},
      {name:'VAPID Web Push',type:'Push',status:'connected',icon:'🌐'},
      {name:'Vonage',type:'SMS',status:'not_configured',icon:'💬'},
    ] as p}
      <div class="card" style="display:flex;align-items:center;gap:16px">
        <div style="font-size:24px">{p.icon}</div>
        <div style="flex:1">
          <div style="font-weight:700;font-size:13px">{p.name}</div>
          <span class="badge badge-neutral" style="margin-top:2px">{p.type}</span>
        </div>
        <span class="badge {p.status==='connected'?'badge-success':'badge-neutral'}">{p.status==='connected'?'● Connected':'Not set up'}</span>
        <button class="btn btn-secondary btn-sm">{p.status==='connected'?'Edit':'Connect'}</button>
      </div>
    {/each}
  </div>
</div>
