<svelte:head><title>Multi-Factor Auth — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Multi-Factor Authentication</h1><p class="page-subtitle">Configure MFA methods — TOTP, SMS, backup codes</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-3 gap-4">
    {#each [
      {name:'TOTP Authenticator',desc:'Google Authenticator, Authy, 1Password — time-based one-time passwords',icon:'📱',enabled:true},
      {name:'SMS OTP',desc:'One-time codes sent via SMS. Requires Twilio or AWS SNS integration.',icon:'💬',enabled:false},
      {name:'Backup Codes',desc:'10 one-time-use recovery codes printed when MFA is enabled',icon:'🔑',enabled:true},
    ] as m}
      <div class="card">
        <div style="font-size:28px;margin-bottom:10px">{m.icon}</div>
        <div style="font-weight:700;font-size:14px;margin-bottom:4px">{m.name}</div>
        <p style="font-size:12px;color:var(--text-secondary);margin-bottom:14px">{m.desc}</p>
        <div style="display:flex;align-items:center;justify-content:space-between">
          <span class="badge {m.enabled ? 'badge-success' : 'badge-neutral'}">{m.enabled ? 'Enabled' : 'Disabled'}</span>
          <button class="btn btn-sm {m.enabled ? 'btn-danger' : 'btn-primary'}">{m.enabled ? 'Disable' : 'Enable'}</button>
        </div>
      </div>
    {/each}
  </div>
  <div class="card" style="margin-top:24px">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:10px">MFA Enforcement</h3>
    <p style="font-size:12px;color:var(--text-secondary);margin-bottom:10px">Require MFA for all users, specific roles, or make it optional.</p>
    <select class="input" style="max-width:300px">
      <option>Optional — users can choose to enable MFA</option>
      <option>Mandatory — all users must set up MFA</option>
      <option>Role-based — enforce for admin role only</option>
    </select>
  </div>
</div>
