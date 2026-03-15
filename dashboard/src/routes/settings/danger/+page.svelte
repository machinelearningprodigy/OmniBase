<svelte:head><title>Danger Zone — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Danger Zone</h1><p class="page-subtitle">Irreversible project-level actions — proceed with extreme caution</p></div></div>
<div class="page-content">
  <div style="display:flex;flex-direction:column;gap:12px">
    {#each [
      {title:'Pause Project',desc:'Suspend all services. Database remains intact. You can resume at any time.',btn:'Pause Project',level:'warning'},
      {title:'Reset Database',desc:'DROP all tables and data in the public schema. This cannot be undone.',btn:'Reset Database',level:'danger'},
      {title:'Delete Project',desc:'Permanently delete this project and all associated data, users, storage, and functions. Completely irreversible.',btn:'Delete Project',level:'danger'},
    ] as d}
      <div class="card" style="border-color:{d.level==='danger'?'rgba(255,82,82,0.3)':'rgba(255,171,64,0.3)'}">
        <div style="display:flex;align-items:center;justify-content:space-between">
          <div>
            <div style="font-weight:700;font-size:14px;margin-bottom:4px">{d.title}</div>
            <p style="font-size:12px;color:var(--text-secondary)">{d.desc}</p>
          </div>
          <button class="btn btn-danger btn-sm" style="flex-shrink:0;margin-left:24px">{d.btn}</button>
        </div>
      </div>
    {/each}
  </div>
</div>
