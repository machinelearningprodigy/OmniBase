<svelte:head><title>Push Notifications — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Push Notifications</h1><p class="page-subtitle">Cross-platform push — APNs (iOS/macOS), FCM (Android), Web Push (VAPID)</p></div>
  <button class="btn btn-primary">＋ Send Notification</button>
</div>
<div class="page-content">
  <div class="grid-cols-4 gap-4" style="margin-bottom:24px">
    {#each [{l:'Devices Registered',v:'8,420'},{l:'Sent Today',v:'12,800'},{l:'Delivered',v:'99.1%'},{l:'Opened',v:'34.2%'}] as s}
      <div class="stat-card"><div class="stat-value">{s.v}</div><div class="stat-label">{s.l}</div></div>
    {/each}
  </div>
  <div class="grid-cols-2 gap-4">
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Providers</h3>
      {#each [
        {name:'Apple APNs',platform:'iOS / macOS / watchOS',icon:'🍎',status:'connected'},
        {name:'FCM (Android)',platform:'Android 8+',icon:'🤖',status:'connected'},
        {name:'Web Push (VAPID)',platform:'Chrome, Firefox, Edge',icon:'🌐',status:'connected'},
      ] as p}
        <div style="display:flex;align-items:center;gap:12px;padding:10px 0;border-bottom:1px solid var(--border-subtle)">
          <span style="font-size:20px">{p.icon}</span>
          <div style="flex:1">
            <div style="font-size:13px;font-weight:600">{p.name}</div>
            <div style="font-size:11px;color:var(--text-secondary)">{p.platform}</div>
          </div>
          <span class="badge badge-success">● {p.status}</span>
        </div>
      {/each}
    </div>
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Quick Send</h3>
      <div class="form-group"><label class="label">Title</label><input class="input" placeholder="Notification title" /></div>
      <div class="form-group"><label class="label">Body</label><textarea class="input" style="height:80px;display:block" placeholder="Notification body..."></textarea></div>
      <div class="form-group"><label class="label">Target</label>
        <select class="input">
          <option>All Users</option><option>iOS Only</option><option>Android Only</option><option>Topic: news</option>
        </select>
      </div>
      <button class="btn btn-primary">Send Push</button>
    </div>
  </div>
</div>
