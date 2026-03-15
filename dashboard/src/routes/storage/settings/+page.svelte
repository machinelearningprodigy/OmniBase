<svelte:head><title>Storage Settings — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Storage Settings</h1><p class="page-subtitle">Configure the object store — file size limits, upload restrictions, and S3 backend</p></div></div>
<div class="page-content">
  <div class="card" style="max-width:500px">
    {#each [
      {label:'Max Upload Size', val:'50', suffix:'MB'},
      {label:'Max Signed URL Expiry', val:'604800', suffix:'seconds (7 days)'},
    ] as s}
      <div class="form-group"><label class="label">{s.label}</label>
        <div style="display:flex;align-items:center;gap:8px">
          <input class="input" value={s.val} style="flex:1" />
          <span style="font-size:12px;color:var(--text-secondary)">{s.suffix}</span>
        </div>
      </div>
    {/each}
    <div class="form-group">
      <label class="label">Allowed MIME Types (comma-separated, empty = all)</label>
      <input class="input" placeholder="image/*, video/mp4, application/pdf" />
    </div>
    <button class="btn btn-primary">Save</button>
  </div>
</div>
