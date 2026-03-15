<svelte:head><title>Delivery Logs — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Delivery Logs</h1><p class="page-subtitle">Track every notification delivery — sent, delivered, opened, failed, bounced</p></div></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>ID</th><th>Channel</th><th>Recipient</th><th>Sent At</th><th>Status</th><th>Opened</th></tr></thead>
      <tbody>
        {#each [
          {id:'msg_x1',ch:'push',to:'+1 555-0100',sent:'3 min ago',status:'delivered',opened:true},
          {id:'msg_x2',ch:'email',to:'alice@example.com',sent:'5 min ago',status:'delivered',opened:false},
          {id:'msg_x3',ch:'sms',to:'+44 7700 900011',sent:'8 min ago',status:'failed',opened:false},
          {id:'msg_x4',ch:'push',to:'device:abc123',sent:'12 min ago',status:'delivered',opened:true},
        ] as l}
          <tr>
            <td class="cell-mono" style="font-size:11px">{l.id}</td>
            <td><span class="badge badge-neutral">{l.ch}</span></td>
            <td style="font-size:12px">{l.to}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{l.sent}</td>
            <td><span class="badge {l.status==='delivered'?'badge-success':'badge-error'}">{l.status}</span></td>
            <td><span class="badge {l.opened?'badge-success':'badge-neutral'}">{l.opened?'Yes':'No'}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
