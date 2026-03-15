<svelte:head><title>Build Pipeline — OmniBase Hosting</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Build Pipeline</h1><p class="page-subtitle">Auto-detect framework (Vite, Next.js, SvelteKit, Astro, Nuxt) and run build commands</p></div></div>
<div class="page-content">
  <div class="card">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Supported Frameworks (Auto-detected)</h3>
    <div class="grid-cols-4 gap-4">
      {#each ['Vite','Next.js','SvelteKit','Astro','Nuxt 3','Remix','Gatsby','Create React App'] as f}
        <div style="padding:12px;background:rgba(255,255,255,0.03);border:1px solid var(--border-subtle);border-radius:8px;text-align:center;font-size:13px;font-weight:500">{f}</div>
      {/each}
    </div>
  </div>
</div>
