<svelte:head><title>Enum Types — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Enum Types</h1><p class="page-subtitle">Manage PostgreSQL enumerated types used across your database schema</p></div>
  <button class="btn btn-primary">＋ New Enum</button>
</div>
<div class="page-content">
  <div class="grid-cols-3 gap-4">
    {#each [
      {name:'user_role', values:['admin','member','viewer','guest']},
      {name:'order_status', values:['pending','processing','shipped','delivered','cancelled']},
      {name:'notification_type', values:['email','push','sms','in_app']},
    ] as e}
      <div class="card">
        <div style="font-size:12px;font-weight:700;color:#a78bfa;margin-bottom:8px;font-family:var(--font-mono)">{e.name}</div>
        <div style="display:flex;flex-wrap:wrap;gap:4px">
          {#each e.values as v}
            <span class="badge badge-neutral">{v}</span>
          {/each}
        </div>
        <div style="margin-top:12px;display:flex;gap:8px">
          <button class="btn btn-secondary btn-sm">Edit</button>
          <button class="btn btn-danger btn-sm">Delete</button>
        </div>
      </div>
    {/each}
  </div>
</div>
