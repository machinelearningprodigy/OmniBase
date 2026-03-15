<svelte:head><title>Email Flows — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Email Flows</h1><p class="page-subtitle">Customize transactional emails — confirmation, magic link, password reset, and more</p></div>
  <button class="btn btn-primary">＋ Custom Template</button>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    {#each [
      {name:'Account Confirmation',trigger:'On signup',provider:'SMTP',sent:1240,last:'2min ago'},
      {name:'Magic Link',trigger:'On passwordless login',provider:'SMTP',sent:380,last:'4min ago'},
      {name:'Password Reset',trigger:'On forgot-password',provider:'SMTP',sent:92,last:'1hr ago'},
      {name:'Email Change',trigger:'On email update',provider:'SMTP',sent:14,last:'2days ago'},
      {name:'Invite User',trigger:'Manual / API',provider:'SMTP',sent:56,last:'3hr ago'},
      {name:'MFA Verification',trigger:'On MFA enable',provider:'SMTP',sent:201,last:'8min ago'},
    ] as t}
      <div class="card" style="display:flex;align-items:center;gap:16px">
        <div style="font-size:24px">📧</div>
        <div style="flex:1">
          <div style="font-weight:600;font-size:13px;margin-bottom:2px">{t.name}</div>
          <div style="font-size:11px;color:var(--text-secondary)">{t.trigger} · {t.sent} sent · last {t.last}</div>
        </div>
        <span class="badge badge-success">{t.provider}</span>
        <button class="btn btn-secondary btn-sm">Edit</button>
      </div>
    {/each}
  </div>
</div>
