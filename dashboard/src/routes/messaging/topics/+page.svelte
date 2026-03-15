<svelte:head><title>Topics — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Topics</h1><p class="page-subtitle">Subscribe devices & users to named topics — broadcast to all subscribers with one call</p></div><button class="btn btn-primary">＋ New Topic</button></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Topic</th><th>Subscribers</th><th>Last Broadcast</th><th>Total Sent</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {name:'news',subs:4200,last:'2hrs ago',total:'84k'},
          {name:'product-updates',subs:12800,last:'3 days ago',total:'42k'},
          {name:'alerts',subs:8400,last:'4min ago',total:'200k'},
          {name:'beta-testers',subs:340,last:'12hrs ago',total:'12k'},
        ] as t}
          <tr>
            <td class="cell-mono">{t.name}</td>
            <td class="cell-mono">{t.subs.toLocaleString()}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{t.last}</td>
            <td class="cell-mono">{t.total}</td>
            <td style="display:flex;gap:6px"><button class="btn btn-primary btn-sm">Broadcast</button><button class="btn btn-danger btn-sm">Delete</button></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
