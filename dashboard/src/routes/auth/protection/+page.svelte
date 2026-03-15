<svelte:head><title>Attack Guard — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Attack Guard</h1><p class="page-subtitle">Protection against brute-force, credential stuffing, and bot attacks</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4" style="margin-bottom:24px">
    <div class="stat-card">
      <div class="stat-value">24</div>
      <div class="stat-label">Blocked IPs (Last 24h)</div>
      <span class="stat-delta negative">+3 today</span>
    </div>
    <div class="stat-card">
      <div class="stat-value">1,840</div>
      <div class="stat-label">Bot Requests Blocked</div>
      <span class="stat-delta positive">-12%</span>
    </div>
  </div>
  <div class="card" style="margin-bottom:16px">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Active Protections</h3>
    {#each [
      {label:'Captcha on login after 3 failed attempts',enabled:true},
      {label:'IP auto-ban after 10 failed attempts within 5 min',enabled:true},
      {label:'Bot score filtering (Cloudflare Turnstile)',enabled:false},
      {label:'Device fingerprint anomaly detection',enabled:true},
      {label:'Impossible travel detection',enabled:false},
    ] as p}
      <div style="display:flex;align-items:center;justify-content:space-between;padding:10px 0;border-bottom:1px solid var(--border-subtle)">
        <span style="font-size:13px">{p.label}</span>
        <span class="badge {p.enabled ? 'badge-success' : 'badge-neutral'}">{p.enabled ? 'Active' : 'Off'}</span>
      </div>
    {/each}
  </div>
</div>
