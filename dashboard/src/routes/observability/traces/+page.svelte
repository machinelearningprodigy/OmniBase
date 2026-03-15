<svelte:head><title>Traces — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Distributed Traces</h1><p class="page-subtitle">OpenTelemetry traces — follow a request across all microservices</p></div></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Trace ID</th><th>Operation</th><th>Duration</th><th>Services</th><th>Status</th><th>Time</th></tr></thead>
      <tbody>
        {#each [
          {id:'ab12cd3e',op:'POST /api/v1/db/query',dur:'14ms',svcs:['gateway','db'],ok:true,ts:'2s ago'},
          {id:'ef45gh6i',op:'POST /auth/v1/token',dur:'48ms',svcs:['gateway','auth','db'],ok:true,ts:'4s ago'},
          {id:'jk78lm9n',op:'POST /functions/send-email',dur:'340ms',svcs:['gateway','functions','smtp'],ok:false,ts:'8s ago'},
          {id:'op01qr2s',op:'GET /storage/v1/object',dur:'22ms',svcs:['gateway','storage'],ok:true,ts:'11s ago'},
        ] as t}
          <tr>
            <td class="cell-mono" style="font-size:11px">{t.id}</td>
            <td class="cell-mono" style="font-size:11px">{t.op}</td>
            <td class="cell-mono" style="color:{parseInt(t.dur)>200?'#ff5252':'#00e676'}">{t.dur}</td>
            <td><div style="display:flex;gap:4px">{#each t.svcs as s}<span class="badge badge-neutral" style="font-size:10px">{s}</span>{/each}</div></td>
            <td><span class="badge {t.ok?'badge-success':'badge-error'}">{t.ok?'OK':'Error'}</span></td>
            <td style="font-size:12px;color:var(--text-secondary)">{t.ts}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
