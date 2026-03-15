<svelte:head><title>Auth Hooks — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Auth Hooks</h1><p class="page-subtitle">Run custom logic at key auth lifecycle events — signup, login, password reset, and more</p></div>
  <span class="badge badge-warning">BETA</span>
  <button class="btn btn-primary">＋ New Hook</button>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    {#each [
      {event:'Before Signup',desc:'Validate or block signups — check invite codes, domain allowlists',fn:'validate_signup',enabled:true},
      {event:'After Signup',desc:'Create a user profile row, send welcome email, trigger onboarding flow',fn:'on_new_user',enabled:true},
      {event:'Before Login',desc:'Log login attempts, check account status, apply extra verification',fn:null,enabled:false},
      {event:'After Login',desc:'Update last-seen timestamp, sync to CRM, trigger analytics events',fn:'track_login',enabled:false},
      {event:'Password Reset',desc:'Custom password reset logic, rate limiting, audit logging',fn:null,enabled:false},
      {event:'Token Refresh',desc:'Validate session is still valid, check revocation list',fn:null,enabled:false},
    ] as h}
      <div class="card">
        <div style="display:flex;align-items:center;gap:8px;margin-bottom:6px">
          <span style="font-weight:700;font-size:13px">{h.event}</span>
          <span class="badge {h.enabled ? 'badge-success' : 'badge-neutral'}">{h.enabled ? 'Active' : 'Off'}</span>
        </div>
        <p style="font-size:12px;color:var(--text-secondary);margin-bottom:10px">{h.desc}</p>
        {#if h.fn}<code style="font-size:11px;display:block;margin-bottom:10px">{h.fn}()</code>{/if}
        <div style="display:flex;gap:8px">
          <button class="btn btn-secondary btn-sm">{h.fn ? 'Edit' : 'Configure'}</button>
          {#if h.enabled}<button class="btn btn-danger btn-sm">Disable</button>{/if}
        </div>
      </div>
    {/each}
  </div>
</div>
