<svelte:head><title>AI Query Builder — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">AI Query Builder</h1><p class="page-subtitle">Type a question in plain English — OmniBase writes the SQL for you</p></div><span class="badge badge-success">NEW</span></div>
<div class="page-content">
  <div class="card" style="margin-bottom:16px">
    <div style="display:flex;gap:12px;margin-bottom:12px">
      <input class="input" style="flex:1" placeholder="e.g. Show me the top 10 users by total revenue in the last 30 days" value="Show me the top 10 users by total revenue in the last 30 days" />
      <button class="btn btn-primary">✨ Generate SQL</button>
    </div>
    <div style="background:var(--bg-elevated);border-radius:8px;padding:16px;font-family:var(--font-mono);font-size:13px;line-height:1.7;border:1px solid var(--border-subtle)">
      <span style="color:#5050a0">-- AI-generated query</span><br/>
      <span style="color:#7dd3fc">SELECT</span> <span style="color:#f0f0ff">u.email, u.name,</span><br/>
      &nbsp;&nbsp;<span style="color:#a78bfa">SUM</span><span style="color:#f0f0ff">(o.total_amount)</span> <span style="color:#7dd3fc">AS</span> <span style="color:#f0f0ff">total_revenue</span><br/>
      <span style="color:#7dd3fc">FROM</span> <span style="color:#f0f0ff">users u</span><br/>
      <span style="color:#7dd3fc">JOIN</span> <span style="color:#f0f0ff">orders o</span> <span style="color:#7dd3fc">ON</span> <span style="color:#f0f0ff">o.user_id = u.id</span><br/>
      <span style="color:#7dd3fc">WHERE</span> <span style="color:#f0f0ff">o.created_at</span> <span style="color:#7dd3fc">&gt;=</span> <span style="color:#a78bfa">NOW</span><span style="color:#f0f0ff">() -</span> <span style="color:#7dd3fc">INTERVAL</span> <span style="color:#98f5a0">'30 days'</span><br/>
      <span style="color:#7dd3fc">GROUP BY</span> <span style="color:#f0f0ff">u.id, u.email, u.name</span><br/>
      <span style="color:#7dd3fc">ORDER BY</span> <span style="color:#f0f0ff">total_revenue</span> <span style="color:#7dd3fc">DESC</span><br/>
      <span style="color:#7dd3fc">LIMIT</span> <span style="color:#98f5a0">10</span><span style="color:#f0f0ff">;</span>
    </div>
    <div style="display:flex;gap:8px;margin-top:12px">
      <a href="/database/editor" class="btn btn-primary btn-sm">▶ Run in SQL Lab</a>
      <button class="btn btn-secondary btn-sm">📋 Copy</button>
      <button class="btn btn-secondary btn-sm">🔄 Regenerate</button>
    </div>
  </div>
  <div class="card">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:10px">Suggested Queries</h3>
    <div style="display:flex;flex-direction:column;gap:6px">
      {#each [
        'Count new users per day for the last week',
        'Find all tables with no Row Level Security policies',
        'Show me the 5 slowest queries from pg_stat_statements',
        'List all foreign key relationships in the public schema',
      ] as q}
        <button class="btn btn-secondary" style="text-align:left;justify-content:start;font-weight:400;font-size:12px">💡 {q}</button>
      {/each}
    </div>
  </div>
</div>
