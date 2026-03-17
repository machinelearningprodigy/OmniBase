<svelte:head><title>Broadcast — OmniBase Live Sync</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Broadcast</h1><p class="page-subtitle">Send ephemeral messages to all clients in a channel — game state, collaborative ops, notifications</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4" style="margin-bottom:24px">
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Send Test Broadcast</h3>
      <div class="form-group">
        <label class="label">Channel</label>
        <input class="input input-mono" value="room:general" />
      </div>
      <div class="form-group">
        <label class="label">Event Type</label>
        <input class="input input-mono" value="cursor_move" />
      </div>
      <div class="form-group">
        <label class="label">Payload (JSON)</label>
        <textarea class="input input-mono" style="height:100px;display:block">{"{\"x\": 42, \"y\": 88, \"user\": \"alice\"}"}</textarea>
      </div>
      <button class="btn btn-primary">⚡ Broadcast</button>
    </div>
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Recent Broadcasts (last 60s)</h3>
      {#each [
        {ts:'07:47:12',ch:'room:general',event:'cursor_move',size:'42 B'},
        {ts:'07:47:08',ch:'room:general',event:'selection_change',size:'18 B'},
        {ts:'07:47:01',ch:'project:abc',event:'comment_added',size:'128 B'},
        {ts:'07:46:58',ch:'broadcast:alerts',event:'alert',size:'64 B'},
      ] as b}
        <div style="display:flex;align-items:center;gap:8px;padding:6px 0;border-bottom:1px solid var(--border-subtle);font-size:12px">
          <span class="cell-mono" style="color:var(--text-muted);font-size:11px">{b.ts}</span>
          <code style="flex:1">{b.ch}</code>
          <span style="color:var(--text-secondary)">{b.event}</span>
          <span style="color:var(--text-muted)">{b.size}</span>
        </div>
      {/each}
    </div>
  </div>
</div>
