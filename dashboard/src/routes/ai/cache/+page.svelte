<svelte:head><title>Semantic Cache — OmniBase AI</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Semantic Cache</h1><p class="page-subtitle">Cache LLM responses by semantic similarity — identical intent → cached response, zero latency</p></div><span class="badge badge-success">NEW</span></div>
<div class="page-content">
  <div class="grid-cols-4 gap-4" style="margin-bottom:24px">
    {#each [{l:'Cache Hits Today',v:'4,240'},{l:'Hit Rate',v:'68%'},{l:'Cost Saved',v:'$12.40'},{l:'Avg Similarity Threshold',v:'0.92'}] as s}
      <div class="stat-card"><div class="stat-value">{s.v}</div><div class="stat-label">{s.l}</div></div>
    {/each}
  </div>
  <div class="card" style="max-width:500px">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Cache Configuration</h3>
    <div class="form-group"><label class="label">Similarity Threshold (0-1)</label><input class="input" type="number" value="0.92" min="0" max="1" step="0.01" /></div>
    <div class="form-group"><label class="label">TTL (seconds)</label><input class="input" type="number" value="3600" /></div>
    <div class="form-group"><label class="label">Max Cache Entries</label><input class="input" type="number" value="10000" /></div>
    <button class="btn btn-primary">Save Config</button>
  </div>
</div>
