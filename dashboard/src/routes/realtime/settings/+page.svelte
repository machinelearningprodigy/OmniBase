<svelte:head><title>Sync Settings — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Sync Settings</h1><p class="page-subtitle">Configure the realtime engine — max connections, heartbeat, and reconnection behavior</p></div></div>
<div class="page-content">
  <div class="card" style="max-width:500px">
    {#each [
      {label:'Max Concurrent Connections', val:'200', type:'number'},
      {label:'Heartbeat Interval (ms)', val:'25000', type:'number'},
      {label:'Reconnect Backoff Base (ms)', val:'1000', type:'number'},
      {label:'Max Reconnect Attempts', val:'5', type:'number'},
    ] as s}
      <div class="form-group"><label class="label">{s.label}</label><input class="input" type={s.type} value={s.val} /></div>
    {/each}
    <button class="btn btn-primary">Save Settings</button>
  </div>
</div>
