<svelte:head><title>URL Config — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">URL Configuration</h1><p class="page-subtitle">Configure allowed redirect URLs and email link destinations for authentication flows</p></div>
  <button class="btn btn-primary">＋ Add URL</button>
</div>
<div class="page-content">
  <div class="card" style="margin-bottom:16px">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:8px">Site URL</h3>
    <p style="font-size:12px;color:var(--text-secondary);margin-bottom:12px">The base URL for your site — used as redirect after email confirmation and OAuth flows.</p>
    <input class="input" value="http://localhost:5173" style="max-width:400px" />
  </div>
  <div class="card">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:8px">Redirect Allow-list</h3>
    <p style="font-size:12px;color:var(--text-secondary);margin-bottom:12px">Additional allowed redirect URLs. Glob patterns supported.</p>
    {#each ['http://localhost:5173/**','https://app.example.com/**','https://*.preview.example.com/**'] as url}
      <div style="display:flex;align-items:center;gap:8px;padding:8px 0;border-bottom:1px solid var(--border-subtle)">
        <code style="flex:1">{url}</code>
        <button class="btn btn-danger btn-sm">Remove</button>
      </div>
    {/each}
    <button class="btn btn-secondary btn-sm" style="margin-top:10px">＋ Add URL</button>
  </div>
</div>
