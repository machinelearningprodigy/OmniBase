<svelte:head><title>DB Branching — OmniBase</title></svelte:head>
<div class="page-header">
  <div>
    <h1 class="page-title">DB Branching</h1>
    <p class="page-subtitle">Fork your entire database per pull request — test migrations safely, merge like code</p>
  </div>
  <span class="badge badge-warning">BETA</span>
  <button class="btn btn-primary">⑂ Create Branch</button>
</div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Branch Name</th><th>Parent</th><th>Created</th><th>Size Delta</th><th>Status</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {name:'main',parent:'—',created:'origin',size:'4.2 GB',status:'production'},
          {name:'feature/ai-search',parent:'main',created:'Mar 14',size:'+12 MB',status:'active'},
          {name:'feat/new-schema',parent:'main',created:'Mar 12',size:'+4 MB',status:'active'},
          {name:'pr-142-hotfix',parent:'main',created:'Mar 10',size:'+0.5 MB',status:'merged'},
        ] as b}
          <tr>
            <td class="cell-mono">{b.name}</td>
            <td class="cell-mono" style="color:var(--text-secondary)">{b.parent}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{b.created}</td>
            <td class="cell-mono">{b.size}</td>
            <td><span class="badge {b.status === 'production' ? 'badge-success' : b.status === 'merged' ? 'badge-neutral' : 'badge-info'}">{b.status}</span></td>
            <td style="display:flex;gap:6px">
              {#if b.status === 'active'}
                <button class="btn btn-secondary btn-sm">Merge</button>
                <button class="btn btn-secondary btn-sm">Diff</button>
                <button class="btn btn-danger btn-sm">Delete</button>
              {:else if b.status === 'production'}
                <span style="font-size:11px;color:var(--text-muted)">protected</span>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
