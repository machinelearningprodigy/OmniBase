<svelte:head><title>Hybrid Search — OmniBase AI</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Hybrid Search</h1><p class="page-subtitle">BM25 full-text + vector cosine similarity combined via Reciprocal Rank Fusion (RRF)</p></div></div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Run Hybrid Search</h3>
      <div class="form-group"><label class="label">Query</label><input class="input" value="how to implement authentication in react" /></div>
      <div class="form-group"><label class="label">Table</label><select class="input"><option>documents</option><option>posts</option></select></div>
      <div class="form-group" style="display:flex;gap:12px">
        <div style="flex:1"><label class="label">BM25 Weight</label><input class="input" type="number" value="0.5" min="0" max="1" step="0.1" /></div>
        <div style="flex:1"><label class="label">Vector Weight</label><input class="input" type="number" value="0.5" min="0" max="1" step="0.1" /></div>
      </div>
      <button class="btn btn-primary">🔍 Hybrid Search</button>
    </div>
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">How It Works</h3>
      {#each [
        {step:'1', label:'BM25 Search', desc:'Scores documents by keyword frequency (TF-IDF style)'},
        {step:'2', label:'Vector Search', desc:'Scores documents by semantic cosine similarity to query embedding'},
        {step:'3', label:'RRF Fusion', desc:'Combines both rank lists using Reciprocal Rank Fusion for best-of-both results'},
        {step:'4', label:'Re-ranker (optional)', desc:'Cross-encoder model re-ranks top-K for highest quality'},
      ] as h}
        <div style="display:flex;gap:12px;padding:8px 0;border-bottom:1px solid var(--border-subtle)">
          <div style="width:22px;height:22px;border-radius:50%;background:linear-gradient(135deg,#6c47ff,#00d4ff);display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:700;color:white;flex-shrink:0">{h.step}</div>
          <div>
            <div style="font-size:13px;font-weight:600">{h.label}</div>
            <div style="font-size:12px;color:var(--text-secondary)">{h.desc}</div>
          </div>
        </div>
      {/each}
    </div>
  </div>
</div>
