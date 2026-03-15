<svelte:head><title>Agent Workflows — OmniBase AI</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Agent Workflows</h1><p class="page-subtitle">Multi-step AI agent pipelines with persistent state, tool calls, and async long-running execution</p></div><span class="badge badge-success">NEW</span><button class="btn btn-primary">＋ New Workflow</button></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Workflow</th><th>Steps</th><th>Runs Today</th><th>Avg Duration</th><th>Success Rate</th><th>Status</th></tr></thead>
      <tbody>
        {#each [
          {name:'code-review-agent',steps:5,runs:48,dur:'42s',success:'96%',status:'active'},
          {name:'document-indexer',steps:3,runs:140,dur:'8s',success:'99%',status:'active'},
          {name:'customer-support-bot',steps:8,runs:280,dur:'15s',success:'94%',status:'active'},
          {name:'report-generator',steps:4,runs:24,dur:'120s',success:'92%',status:'paused'},
        ] as w}
          <tr>
            <td class="cell-mono">{w.name}</td>
            <td><span class="badge badge-neutral">{w.steps} steps</span></td>
            <td class="cell-mono">{w.runs}</td>
            <td class="cell-mono">{w.dur}</td>
            <td style="color:#00e676;font-weight:600">{w.success}</td>
            <td><span class="badge {w.status==='active'?'badge-success':'badge-neutral'}">{w.status==='active'?'● Active':'Paused'}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
