<svelte:head><title>Templates — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Message Templates</h1><p class="page-subtitle">Reusable notification templates with variable substitution — for push, email, and SMS</p></div><button class="btn btn-primary">＋ New Template</button></div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    {#each [
      {name:'welcome_push',type:'Push',title:'Welcome, {{name}}!',body:'Thanks for joining. Your account is ready.'},
      {name:'order_shipped',type:'Push + Email',title:'Your order #{{order_id}} has shipped!',body:'Track it at {{tracking_url}}'},
      {name:'password_changed',type:'Email + SMS',title:'Security Alert: Password Changed',body:'Your password was changed on {{device}} at {{time}}.'},
    ] as t}
      <div class="card">
        <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
          <code style="font-size:12px">{t.name}</code>
          <span class="badge badge-info">{t.type}</span>
        </div>
        <div style="font-weight:600;font-size:13px;margin-bottom:4px">{t.title}</div>
        <p style="font-size:12px;color:var(--text-secondary);margin-bottom:12px">{t.body}</p>
        <div style="display:flex;gap:8px">
          <button class="btn btn-secondary btn-sm">Edit</button>
          <button class="btn btn-secondary btn-sm">Preview</button>
        </div>
      </div>
    {/each}
  </div>
</div>
