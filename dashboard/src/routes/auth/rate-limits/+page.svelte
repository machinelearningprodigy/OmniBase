<svelte:head><title>Rate Limits — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Rate Limits</h1><p class="page-subtitle">Protect your auth endpoints from abuse and brute-force attacks</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    {#each [
      {label:'Email/Password Sign In',current:10,max:30,period:'minute'},
      {label:'Magic Link Send',current:3,max:5,period:'hour'},
      {label:'OTP SMS',current:2,max:5,period:'hour'},
      {label:'Password Reset Email',current:1,max:3,period:'hour'},
      {label:'Token Refresh',current:150,max:360,period:'hour'},
      {label:'Signup',current:5,max:10,period:'minute'},
    ] as r}
      <div class="card">
        <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:10px">
          <span style="font-size:13px;font-weight:600">{r.label}</span>
          <span class="badge badge-info">{r.current}/{r.max} per {r.period}</span>
        </div>
        <div style="background:rgba(255,255,255,0.05);border-radius:4px;height:6px;overflow:hidden">
          <div style="height:100%;width:{(r.current/r.max*100).toFixed(0)}%;background:linear-gradient(90deg,#6c47ff,#00d4ff);border-radius:4px"></div>
        </div>
        <div style="margin-top:8px;display:flex;justify-content:flex-end">
          <button class="btn btn-secondary btn-sm">Edit Limit</button>
        </div>
      </div>
    {/each}
  </div>
</div>
