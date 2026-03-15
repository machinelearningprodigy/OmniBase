<svelte:head><title>Job Queues — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Job Queues</h1><p class="page-subtitle">NATS JetStream-backed message queues — reliable async processing with DLQ support</p></div><button class="btn btn-primary">＋ New Queue</button></div>
<div class="page-content">
  <div class="grid-cols-3 gap-4" style="margin-bottom:24px">
    {#each [{l:'Total Jobs Today',v:'12,440'},{l:'Processing',v:'14'},{l:'Dead Letter',v:'3'}] as s}
      <div class="stat-card"><div class="stat-value">{s.v}</div><div class="stat-label">{s.l}</div></div>
    {/each}
  </div>
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Queue Name</th><th>Consumer</th><th>Pending</th><th>Processing</th><th>DLQ</th><th>Throughput</th></tr></thead>
      <tbody>
        {#each [
          {name:'email-queue',fn:'send-email',pending:4,proc:2,dlq:0,tput:'120/min'},
          {name:'payment-queue',fn:'process-payment',pending:1,proc:1,dlq:1,tput:'24/min'},
          {name:'webhook-queue',fn:'fire-webhook',pending:18,proc:5,dlq:2,tput:'480/min'},
        ] as q}
          <tr>
            <td class="cell-mono">{q.name}</td>
            <td class="cell-mono">{q.fn}</td>
            <td><span class="badge badge-info">{q.pending}</span></td>
            <td><span class="badge badge-success">{q.proc}</span></td>
            <td><span class="badge {q.dlq > 0 ? 'badge-error' : 'badge-neutral'}">{q.dlq}</span></td>
            <td class="cell-mono">{q.tput}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
