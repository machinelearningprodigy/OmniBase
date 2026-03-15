<svelte:head><title>Graph Queries — OmniBase</title></svelte:head>
<div class="page-header">
  <div>
    <h1 class="page-title">Graph Queries</h1>
    <p class="page-subtitle">Property graph queries with Cypher language via Apache AGE — no separate graph database needed</p>
  </div>
  <span class="badge badge-success">NEW</span>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4" style="margin-bottom:24px">
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px;color:var(--text-secondary)">Cypher Query Editor</h3>
      <textarea
        class="input input-mono"
        style="height:180px;resize:vertical;display:block"
        placeholder="MATCH (u:User)-[:FOLLOWS]->(f:User)&#10;WHERE u.id = 'user-123'&#10;RETURN f.name, f.email&#10;LIMIT 10"
      ></textarea>
      <div style="margin-top:10px;display:flex;gap:8px">
        <button class="btn btn-primary btn-sm">▶ Run Query</button>
        <button class="btn btn-secondary btn-sm">Save</button>
      </div>
    </div>
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px;color:var(--text-secondary)">Graph Stats</h3>
      {#each [
        {label:'Vertex Labels', value:'8', note:'users, posts, tags, orgs...'},
        {label:'Edge Labels', value:'12', note:'follows, likes, owns, member_of...'},
        {label:'Total Vertices', value:'142k'},
        {label:'Total Edges', value:'1.8M'},
      ] as s}
        <div style="display:flex;justify-content:space-between;align-items:center;padding:8px 0;border-bottom:1px solid var(--border-subtle)">
          <span style="font-size:12px;color:var(--text-secondary)">{s.label}</span>
          <span style="font-weight:700;color:#a78bfa">{s.value}</span>
        </div>
      {/each}
    </div>
  </div>
  <div class="card">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Graph Visualizer</h3>
    <div style="height:300px;background:rgba(255,255,255,0.02);border-radius:8px;border:1px solid var(--border-subtle);display:flex;align-items:center;justify-content:center;color:var(--text-muted)">
      Run a MATCH query to visualize the graph nodes and edges
    </div>
  </div>
</div>
