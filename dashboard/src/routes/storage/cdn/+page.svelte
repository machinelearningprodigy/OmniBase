<svelte:head><title>CDN Config — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">CDN Config</h1><p class="page-subtitle">Configure edge caching and content delivery for your stored assets</p></div></div>
<div class="page-content">
  <div class="card" style="max-width:560px">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Cache Settings</h3>
    {#each [
      {label:'Cache TTL for Public Buckets', val:'86400', suffix:'seconds'},
      {label:'Cache TTL for Transformed Images', val:'31536000', suffix:'seconds'},
      {label:'Stale-While-Revalidate', val:'3600', suffix:'seconds'},
    ] as s}
      <div class="form-group"><label class="label">{s.label}</label>
        <div style="display:flex;align-items:center;gap:8px">
          <input class="input" value={s.val} style="flex:1" />
          <span style="font-size:12px;color:var(--text-secondary)">{s.suffix}</span>
        </div>
      </div>
    {/each}
    <button class="btn btn-primary">Save CDN Settings</button>
  </div>
</div>
