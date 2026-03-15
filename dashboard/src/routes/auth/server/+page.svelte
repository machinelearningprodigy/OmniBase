<svelte:head><title>Auth Server — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Auth Server</h1><p class="page-subtitle">OmniBase OAuth 2.0 authorization server — issue tokens to your own OAuth apps</p></div>
  <span class="badge badge-warning">BETA</span>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4" style="margin-bottom:24px">
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Server Endpoints</h3>
      {#each [
        ['Authorization URL','/oauth/authorize'],
        ['Token URL','/oauth/token'],
        ['UserInfo URL','/oauth/userinfo'],
        ['JWKS URL','/.well-known/jwks.json'],
        ['Discovery',  '/.well-known/openid-configuration'],
      ] as [label, path]}
        <div style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid var(--border-subtle)">
          <span style="font-size:12px;color:var(--text-secondary)">{label}</span>
          <code style="font-size:11px">{path}</code>
        </div>
      {/each}
    </div>
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Registered Clients</h3>
      <div class="empty-state" style="padding:30px">
        <h3>No OAuth clients yet</h3>
        <p>Register your first OAuth application to enable third-party login to your platform.</p>
        <button class="btn btn-primary" style="margin-top:12px">Register Client</button>
      </div>
    </div>
  </div>
</div>
