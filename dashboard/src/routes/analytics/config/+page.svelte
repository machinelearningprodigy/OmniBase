<svelte:head><title>Remote Config — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Remote Config</h1><p class="page-subtitle">Change app behavior without shipping a new release — feature flags, rollouts, kill switches</p></div><button class="btn btn-primary">＋ New Parameter</button></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Key</th><th>Type</th><th>Default Value</th><th>Overrides</th><th>Last Updated</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {key:'show_ai_panel',type:'boolean',default:'false',overrides:'100% → true',updated:'1hr ago'},
          {key:'max_free_tables',type:'number',default:'10',overrides:'beta_users → 50',updated:'3 days ago'},
          {key:'feature_branching',type:'boolean',default:'false',overrides:'10% users → true',updated:'5 days ago'},
          {key:'welcome_message',type:'string',default:'"Welcome!"',overrides:'None',updated:'2 weeks ago'},
        ] as p}
          <tr>
            <td class="cell-mono">{p.key}</td>
            <td><span class="badge badge-neutral">{p.type}</span></td>
            <td class="cell-mono" style="font-size:12px">{p.default}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{p.overrides}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{p.updated}</td>
            <td style="display:flex;gap:6px"><button class="btn btn-secondary btn-sm">Edit</button><button class="btn btn-danger btn-sm">Delete</button></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
