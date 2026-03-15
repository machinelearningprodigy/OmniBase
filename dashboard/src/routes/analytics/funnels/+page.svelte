<svelte:head><title>Funnels — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Funnels</h1><p class="page-subtitle">Multi-step funnel analysis — see where users drop off in key flows</p></div><button class="btn btn-primary">＋ New Funnel</button></div>
<div class="page-content">
  <div class="card">
    <h3 style="font-size:14px;font-weight:700;margin-bottom:16px">Signup → Activation Funnel</h3>
    {#each [
      {step:'Visited Landing Page',count:10000,pct:100},
      {step:'Started Signup',count:3200,pct:32},
      {step:'Completed Signup',count:2100,pct:21},
      {step:'Created First Table',count:840,pct:8.4},
      {step:'Made First API Call',count:420,pct:4.2},
    ] as f, i}
      <div style="margin-bottom:12px">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:4px">
          <span style="font-size:13px">{i+1}. {f.step}</span>
          <div style="display:flex;align-items:center;gap:10px">
            <span class="cell-mono" style="font-size:12px">{f.count.toLocaleString()}</span>
            <span style="font-size:12px;color:{f.pct<10?'#ff5252':f.pct<30?'#ffab40':'#00e676'};font-weight:700;width:48px;text-align:right">{f.pct}%</span>
          </div>
        </div>
        <div style="background:rgba(255,255,255,0.05);border-radius:4px;height:8px;overflow:hidden">
          <div style="height:100%;width:{f.pct}%;background:linear-gradient(90deg,#6c47ff,#00d4ff);border-radius:4px;transition:width 500ms ease"></div>
        </div>
        {#if i < 4}
          <div style="font-size:11px;color:#ff5252;margin-top:2px">↳ {(100 - ([32,66,40,50][i] || 0)).toFixed(0)}% drop-off</div>
        {/if}
      </div>
    {/each}
  </div>
</div>
