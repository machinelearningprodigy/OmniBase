<svelte:head><title>Integrations — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Integrations</h1><p class="page-subtitle">Connect OmniBase to your existing tools and services</p></div></div>
<div class="page-content">
  <div class="grid-cols-3 gap-4">
    {#each [
      {name:'GitHub',desc:'Sync repos, trigger deployments on push, auto-create DB branches per PR',icon:'🐙',status:'connected'},
      {name:'Slack',desc:'Send alert notifications and team updates to Slack channels',icon:'💬',status:'connected'},
      {name:'Grafana',desc:'Visualize metrics and traces in your Grafana dashboards',icon:'📊',status:'not_configured'},
      {name:'PagerDuty',desc:'Route critical alerts to PagerDuty for on-call escalation',icon:'🚨',status:'not_configured'},
      {name:'Datadog',desc:'Send OpenTelemetry data to Datadog APM and Log Management',icon:'🐕',status:'not_configured'},
      {name:'Vercel',desc:'Use OmniBase as the backend for your Vercel-deployed frontend',icon:'▲',status:'not_configured'},
      {name:'Stripe',desc:'Sync payment events to your database via webhooks',icon:'💳',status:'not_configured'},
      {name:'Sentry',desc:'Forward crash reports to Sentry for advanced error tracking',icon:'🔍',status:'not_configured'},
      {name:'Linear',desc:'Auto-create issues from crash reports and alerts',icon:'📐',status:'not_configured'},
    ] as i}
      <div class="card">
        <div style="font-size:28px;margin-bottom:8px">{i.icon}</div>
        <div style="font-weight:700;font-size:14px;margin-bottom:4px">{i.name}</div>
        <p style="font-size:12px;color:var(--text-secondary);margin-bottom:12px;flex:1">{i.desc}</p>
        <div style="display:flex;align-items:center;justify-content:space-between">
          <span class="badge {i.status==='connected'?'badge-success':'badge-neutral'}">{i.status==='connected'?'● Connected':'Not set up'}</span>
          <button class="btn btn-secondary btn-sm">{i.status==='connected'?'Manage':'Connect'}</button>
        </div>
      </div>
    {/each}
  </div>
</div>
