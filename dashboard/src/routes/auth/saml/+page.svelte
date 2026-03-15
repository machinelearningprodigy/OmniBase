<svelte:head><title>SAML / SSO — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">SAML / SSO</h1><p class="page-subtitle">Enterprise Single Sign-On with any SAML 2.0 Identity Provider — Okta, Azure AD, Google Workspace</p></div>
  <button class="btn btn-primary">＋ Add Provider</button>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4" style="margin-bottom:24px">
    {#each [
      {name:'Okta',desc:'Enterprise-grade workforce identity provider',icon:'🔐',status:'configured'},
      {name:'Azure Active Directory',desc:'Microsoft enterprise identity and access',icon:'🔷',status:'not_configured'},
      {name:'Google Workspace',desc:'G Suite / Google Cloud Identity SAML provider',icon:'🔴',status:'not_configured'},
      {name:'Custom SAML IdP',desc:'Any SAML 2.0 compliant identity provider',icon:'⚙',status:'not_configured'},
    ] as p}
      <div class="card" style="display:flex;align-items:center;gap:16px">
        <div style="font-size:28px">{p.icon}</div>
        <div style="flex:1">
          <div style="font-weight:700;font-size:14px;margin-bottom:2px">{p.name}</div>
          <p style="font-size:12px;color:var(--text-secondary)">{p.desc}</p>
        </div>
        <span class="badge {p.status === 'configured' ? 'badge-success' : 'badge-neutral'}">{p.status === 'configured' ? 'Active' : 'Not set up'}</span>
        <button class="btn btn-secondary btn-sm">{p.status === 'configured' ? 'Edit' : 'Configure'}</button>
      </div>
    {/each}
  </div>
</div>
