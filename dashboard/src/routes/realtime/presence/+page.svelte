<svelte:head><title>Presence — OmniBase Live Sync</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Presence</h1><p class="page-subtitle">Track which users are currently online in each channel — like collaborative cursors</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4" style="margin-bottom:24px">
    <div class="stat-card"><div class="stat-value">142</div><div class="stat-label">Online Users Now</div></div>
    <div class="stat-card"><div class="stat-value">18</div><div class="stat-label">Active Channels with Presence</div></div>
  </div>
  <div class="card">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Live Presence Map</h3>
    <div style="display:flex;flex-direction:column;gap:10px">
      {#each [
        {channel:'room:general',users:['alice','bob','carol','dave','eve'],total:48},
        {channel:'project:abc123',users:['frank','grace','henry'],total:12},
        {channel:'docs:shared-doc-01',users:['irene','james'],total:2},
      ] as ch}
        <div style="display:flex;align-items:center;gap:16px;padding:10px;background:rgba(255,255,255,0.02);border-radius:8px">
          <code style="flex-shrink:0;width:200px">{ch.channel}</code>
          <div style="display:flex;gap:4px;flex:1">
            {#each ch.users as u}
              <div style="width:28px;height:28px;border-radius:50%;background:linear-gradient(135deg,#6c47ff,#00d4ff);display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:700;color:white">{u[0].toUpperCase()}</div>
            {/each}
            {#if ch.total > ch.users.length}
              <div style="width:28px;height:28px;border-radius:50%;background:rgba(255,255,255,0.08);display:flex;align-items:center;justify-content:center;font-size:11px;color:var(--text-secondary)">+{ch.total - ch.users.length}</div>
            {/if}
          </div>
          <span class="badge badge-success">● {ch.total} online</span>
        </div>
      {/each}
    </div>
  </div>
</div>
