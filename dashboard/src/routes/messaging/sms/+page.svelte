<svelte:head><title>SMS — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">SMS</h1><p class="page-subtitle">Send SMS messages via Twilio, AWS SNS, or Vonage</p></div><button class="btn btn-primary">＋ Send SMS</button></div>
<div class="page-content">
  <div class="grid-cols-2 gap-4" style="margin-bottom:24px">
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">SMS Provider</h3>
      {#each [{name:'Twilio',status:'connected'},{name:'AWS SNS',status:'not_configured'},{name:'Vonage',status:'not_configured'}] as p}
        <div style="display:flex;align-items:center;gap:12px;padding:10px 0;border-bottom:1px solid var(--border-subtle)">
          <div style="flex:1;font-size:13px">{p.name}</div>
          <span class="badge {p.status==='connected'?'badge-success':'badge-neutral'}">{p.status==='connected'?'● Connected':'Not set up'}</span>
          <button class="btn btn-secondary btn-sm">{p.status==='connected'?'Edit':'Connect'}</button>
        </div>
      {/each}
    </div>
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Send Test SMS</h3>
      <div class="form-group"><label class="label">Phone Number</label><input class="input" placeholder="+1 555 000 0000" /></div>
      <div class="form-group"><label class="label">Message</label><textarea class="input" style="height:80px;display:block" placeholder="SMS body..."></textarea></div>
      <button class="btn btn-primary">Send SMS</button>
    </div>
  </div>
</div>
