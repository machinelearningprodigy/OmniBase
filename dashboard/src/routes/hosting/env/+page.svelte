<svelte:head><title>Env Variables — OmniBase Hosting</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Environment Variables</h1><p class="page-subtitle">Per-deployment environment variables — different values for preview vs production</p></div><button class="btn btn-primary">＋ Add Variable</button></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Name</th><th>Preview Value</th><th>Production Value</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {name:'VITE_API_URL',preview:'http://localhost:8000',prod:'https://api.yourdomain.com'},
          {name:'VITE_APP_ENV',preview:'preview',prod:'production'},
        ] as v}
          <tr>
            <td class="cell-mono">{v.name}</td>
            <td class="cell-mono" style="color:var(--text-secondary)">{v.preview}</td>
            <td class="cell-mono">{v.prod}</td>
            <td style="display:flex;gap:6px"><button class="btn btn-secondary btn-sm">Edit</button><button class="btn btn-danger btn-sm">Delete</button></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
