<svelte:head><title>A/B Testing — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">A/B Testing</h1><p class="page-subtitle">Define experiments, assign users to variants, and measure conversion impact</p></div><button class="btn btn-primary">＋ New Experiment</button></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Experiment</th><th>Variants</th><th>Users Enrolled</th><th>Duration</th><th>Winner</th><th>Status</th></tr></thead>
      <tbody>
        {#each [
          {name:'Signup CTA button color',variants:['Control (blue)','Treatment (purple)'],users:4200,dur:'14 days',winner:'Treatment (+24%)',status:'complete'},
          {name:'Onboarding flow v2',variants:['Old 5-step','New 3-step'],users:1800,dur:'7 days',winner:'',status:'running'},
          {name:'Pricing page layout',variants:['Current','Card layout','Comparison table'],users:0,dur:'—',winner:'',status:'draft'},
        ] as e}
          <tr>
            <td style="font-size:13px">{e.name}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{e.variants.join(' vs ')}</td>
            <td class="cell-mono">{e.users > 0 ? e.users.toLocaleString() : '—'}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{e.dur}</td>
            <td style="font-size:12px;color:#00e676">{e.winner || '—'}</td>
            <td><span class="badge {e.status==='complete'?'badge-success':e.status==='running'?'badge-info':'badge-neutral'}">{e.status}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
